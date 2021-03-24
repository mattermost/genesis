// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	sdkAWS "github.com/aws/aws-sdk-go/aws"
	"github.com/gorilla/mux"
	"github.com/mattermost/genesis/internal/api"
	"github.com/mattermost/genesis/internal/genesis"
	"github.com/mattermost/genesis/internal/store"
	"github.com/mattermost/genesis/internal/supervisor"

	toolsAWS "github.com/mattermost/genesis/internal/aws"
	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	defaultLocalServerAPI = "http://localhost:8073"
)

var instanceID string

func init() {
	instanceID = model.NewID()

	// General
	serverCmd.PersistentFlags().String("database", "sqlite://genesis.db", "The database backing the genesis server.")
	serverCmd.PersistentFlags().String("listen", ":8073", "The interface and port on which to listen.")
	serverCmd.PersistentFlags().Bool("debug", false, "Whether to output debug logs.")
	serverCmd.PersistentFlags().Bool("machine-readable-logs", false, "Output the logs in machine readable format.")
	serverCmd.PersistentFlags().String("sso-user-email", "", "The email of the SSO admin user")
	serverCmd.PersistentFlags().String("sso-first-name", "", "The first name of the SSO admin user")
	serverCmd.PersistentFlags().String("sso-last-name", "", "The last name of the SSO admin user")
	serverCmd.PersistentFlags().String("managed-ou", "", "The managed organizational unit")
	serverCmd.PersistentFlags().String("control-tower-role", "", "The IAM role that will be assumed in the Control Tower account")
	serverCmd.PersistentFlags().String("control-tower-account", "", "The AWS account ID of the Control Tower account")

	serverCmd.MarkFlagRequired("sso-user-email")
	serverCmd.MarkFlagRequired("sso-first-name")
	serverCmd.MarkFlagRequired("sso-last-name")
	serverCmd.MarkFlagRequired("managed-ou")
	serverCmd.MarkFlagRequired("control-tower-role")
	serverCmd.MarkFlagRequired("control-tower-account")

	// Supervisors
	serverCmd.PersistentFlags().Int("poll", 30, "The interval in seconds to poll for background work.")
	serverCmd.PersistentFlags().Bool("account-supervisor", true, "Whether this server will run an account supervisor or not.")
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the genesis server.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		debug, _ := command.Flags().GetBool("debug")
		if debug {
			logger.SetLevel(logrus.DebugLevel)
		}

		machineLogs, _ := command.Flags().GetBool("machine-readable-logs")
		if machineLogs {
			logger.SetFormatter(&logrus.JSONFormatter{})
		}

		logger := logger.WithField("instance", instanceID)

		sqlStore, err := sqlStore(command)
		if err != nil {
			return err
		}

		currentVersion, err := sqlStore.GetCurrentVersion()
		if err != nil {
			return err
		}
		serverVersion := store.LatestVersion()

		// Require the schema to be at least the server version, and also the same major
		// version.
		if currentVersion.LT(serverVersion) || currentVersion.Major != serverVersion.Major {
			return errors.Errorf("server requires at least schema %s, current is %s", serverVersion, currentVersion)
		}

		accountSupervisor, _ := command.Flags().GetBool("account-supervisor")
		if !accountSupervisor {
			logger.Warn("Server will be running with no supervisors. Only API functionality will work.")
		}

		wd, err := os.Getwd()
		if err != nil {
			wd = "error getting working directory"
			logger.WithError(err).Error("Unable to get current working directory")
		}

		logger.WithFields(logrus.Fields{
			"build-hash":         model.BuildHash,
			"account-supervisor": accountSupervisor,
			"store-version":      currentVersion,
			"working-directory":  wd,
		}).Info("Starting Mattermost Genesis Server")

		deprecationWarnings(logger, command)

		awsConfig := &sdkAWS.Config{
			Region:     sdkAWS.String(toolsAWS.DefaultAWSRegion),
			MaxRetries: sdkAWS.Int(toolsAWS.DefaultAWSClientRetries),
		}
		awsClient := toolsAWS.NewAWSClientWithConfig(awsConfig, logger)

		environment, err := awsClient.GetCloudEnvironmentName()
		if err != nil {
			return errors.Wrap(err, "getting the AWS Cloud environment")
		}

		ssoUserEmail, _ := command.Flags().GetString("sso-user-email")
		ssoFirstName, _ := command.Flags().GetString("sso-first-name")
		ssoLastName, _ := command.Flags().GetString("sso-last-name")
		managedOU, _ := command.Flags().GetString("managed-ou")
		controlTowerRole, _ := command.Flags().GetString("control-tower-role")
		controlTowerAccountID, _ := command.Flags().GetString("control-tower-account")
		// Setup the provisioner for actually effecting changes to enterprise resources.
		genesisProvisioner := genesis.NewGenesisProvisioner(
			ssoUserEmail,
			ssoFirstName,
			ssoLastName,
			managedOU,
			controlTowerRole,
			controlTowerAccountID,
			logger,
		)

		var multiDoer supervisor.MultiDoer
		if accountSupervisor {
			multiDoer = append(multiDoer, supervisor.NewAccountSupervisor(sqlStore, genesisProvisioner, awsClient, instanceID, logger))
		}

		// Setup the supervisor to effect any requested changes. It is wrapped in a
		// scheduler to trigger it periodically in addition to being poked by the API
		// layer.
		poll, _ := command.Flags().GetInt("poll")
		if poll == 0 {
			logger.WithField("poll", poll).Info("Scheduler is disabled")
		}

		supervisor := supervisor.NewScheduler(multiDoer, time.Duration(poll)*time.Second)
		defer supervisor.Close()

		router := mux.NewRouter()

		api.Register(router, &api.Context{
			Store:       sqlStore,
			Supervisor:  supervisor,
			Genesis:     genesisProvisioner,
			Environment: environment,
			Logger:      logger,
		})

		listen, _ := command.Flags().GetString("listen")
		srv := &http.Server{
			Addr:           listen,
			Handler:        router,
			ReadTimeout:    180 * time.Second,
			WriteTimeout:   180 * time.Second,
			IdleTimeout:    time.Second * 180,
			MaxHeaderBytes: 1 << 20,
			ErrorLog:       log.New(&logrusWriter{logger}, "", 0),
		}

		go func() {
			logger.WithField("addr", srv.Addr).Info("Listening")
			err := srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.WithError(err).Error("Failed to listen and serve")
			}
		}()

		c := make(chan os.Signal, 1)
		// We'll accept graceful shutdowns when quit via:
		//  - SIGINT (Ctrl+C)
		//  - SIGTERM (Ctrl+/) (Kubernetes pod rolling termination)
		// SIGKILL and SIGQUIT will not be caught.
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)

		// Block until we receive a valid signal.
		sig := <-c
		logger.WithField("shutdown-signal", sig.String()).Info("Shutting down")

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		srv.Shutdown(ctx)

		return nil
	},
}

func checkRequirements(awsConfig *sdkAWS.Config, s3StateStore string) error {
	utilities := []string{
		"terraform",
	}

	for _, requiredUtility := range utilities {
		_, err := exec.LookPath(requiredUtility)
		if err != nil {
			return errors.Errorf("failed to find %s on the PATH", requiredUtility)
		}
	}

	client := toolsAWS.NewAWSClientWithConfig(awsConfig, logger)
	_, err := client.GetCloudEnvironmentName()
	if err != nil {
		return errors.Wrap(err, "failed to establish a connection with AWS")
	}

	return nil
}

// deprecationWarnings performs all checks for deprecated settings and warns if
// any are found.
func deprecationWarnings(logger logrus.FieldLogger, cmd *cobra.Command) {
	// Add deprecation logic here.
}

// getHumanReadableID  represents  a  best  effort  attempt  to  retrieve  an
// identifiable  human to  associate with  a given  provisioner. Since
// this is for dev workflows, any  error causes it to merely return a
// generic string.
func getHumanReadableID() string {
	envVar := os.Getenv("CLOUD_SERVER_OWNER")
	if envVar != "" {
		return envVar
	}

	cmd := exec.Command("git", "config",
		"--get", "user.email")

	output, err := cmd.Output()
	if err != nil {
		logger.Debugf("Couldn't determine username of developer with which to tag infrastructure due to error: %s", err.Error())
		if len(output) != 0 {
			logger.Debugf("Command output was: %s", string(output))
		}
		return "SRETeam"
	}

	return strings.TrimSpace(string(output))
}
