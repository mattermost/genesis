// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"encoding/json"
	"os"

	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/mattermost/genesis/model"
)

func init() {
	accountCmd.PersistentFlags().String("server", defaultLocalServerAPI, "The genesis server whose API will be queried.")
	accountCmd.PersistentFlags().Bool("dry-run", false, "When set to true, only print the API request without sending it.")

	accountCreateCmd.Flags().String("service-catalog-product", "", "The service catalog product id to provision a new account")
	accountCreateCmd.Flags().String("provider", "aws", "Cloud provider hosting the account.")
	accountCreateCmd.Flags().Bool("provision", false, "When set to true provision an account after creation.")

	accountCreateCmd.MarkFlagRequired("service-catalog-product")

	accountProvisionCmd.Flags().String("account", "", "The id of the account to be deleted.")
	accountProvisionCmd.MarkFlagRequired("account")

	accountDeleteCmd.Flags().String("account", "", "The id of the account to be deleted.")
	accountDeleteCmd.MarkFlagRequired("account")

	accountGetCmd.Flags().String("account", "", "The id of the account to be fetched.")
	accountGetCmd.MarkFlagRequired("account")

	accountListCmd.Flags().Int("page", 0, "The page of accounts to fetch, starting at 0.")
	accountListCmd.Flags().Int("per-page", 100, "The number of accounts to fetch per page.")
	accountListCmd.Flags().Bool("include-deleted", false, "Whether to include deleted accounts.")
	accountListCmd.Flags().Bool("table", false, "Whether to display the returned account list in a table or not")

	accountCmd.AddCommand(accountCreateCmd)
	accountCmd.AddCommand(accountProvisionCmd)
	accountCmd.AddCommand(accountDeleteCmd)
	accountCmd.AddCommand(accountGetCmd)
	accountCmd.AddCommand(accountListCmd)
}

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manipulate accounts managed by the genesis server.",
}

func printJSON(data interface{}) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	return encoder.Encode(data)
}

var accountCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an account.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		provider, _ := command.Flags().GetString("provider")
		serviceCatalogProductID, _ := command.Flags().GetString("service-catalog-product")
		provision, _ := command.Flags().GetBool("provision")

		request := &model.CreateAccountRequest{
			Provider:                provider,
			ServiceCatalogProductID: serviceCatalogProductID,
			Provision:               provision,
		}

		dryRun, _ := command.Flags().GetBool("dry-run")
		if dryRun {
			err := printJSON(request)
			if err != nil {
				return errors.Wrap(err, "failed to print API request")
			}

			return nil
		}

		account, err := client.CreateAccount(request)
		if err != nil {
			return errors.Wrap(err, "failed to create account")
		}

		if err = printJSON(account); err != nil {
			return errors.Wrap(err, "failed to print account response")
		}

		return nil
	},
}

var accountProvisionCmd = &cobra.Command{
	Use:   "provision",
	Short: "Provision/Reprovision an account's cloud resources.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)
		accountID, _ := command.Flags().GetString("account")

		var request *model.ProvisionAccountRequest = nil

		dryRun, _ := command.Flags().GetBool("dry-run")
		if dryRun {
			err := printJSON(request)
			if err != nil {
				return errors.Wrap(err, "failed to print API request")
			}

			return nil
		}

		account, err := client.ProvisionAccount(accountID, request)
		if err != nil {
			return errors.Wrap(err, "failed to provision account")
		}

		if err = printJSON(account); err != nil {
			return errors.Wrap(err, "failed to print account response")
		}

		return nil
	},
}

var accountDeleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an account.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		accountID, _ := command.Flags().GetString("account")

		err := client.DeleteAccount(accountID)
		if err != nil {
			return errors.Wrap(err, "failed to delete account")
		}

		return nil
	},
}

var accountGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a particular account.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		accountID, _ := command.Flags().GetString("account")
		account, err := client.GetAccount(accountID)
		if err != nil {
			return errors.Wrap(err, "failed to query account")
		}
		if account == nil {
			return nil
		}

		if err = printJSON(account); err != nil {
			return errors.Wrap(err, "failed to print account response")
		}

		return nil
	},
}

var accountListCmd = &cobra.Command{
	Use:   "list",
	Short: "List created accounts.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		page, _ := command.Flags().GetInt("page")
		perPage, _ := command.Flags().GetInt("per-page")
		includeDeleted, _ := command.Flags().GetBool("include-deleted")
		accounts, err := client.GetAccounts(&model.GetAccountsRequest{
			Page:           page,
			PerPage:        perPage,
			IncludeDeleted: includeDeleted,
		})
		if err != nil {
			return errors.Wrap(err, "failed to query accounts")
		}

		outputToTable, _ := command.Flags().GetBool("table")
		if outputToTable {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"ID", "STATE", "AWS ACCOUNT ID"})

			for _, account := range accounts {
				table.Append([]string{
					account.ID,
					account.State,
					account.ProviderMetadataAWS.AWSAccountID,
				})
			}
			table.Render()

			return nil
		}

		if err = printJSON(accounts); err != nil {
			return errors.Wrap(err, "failed to print account response")
		}

		return nil
	},
}
