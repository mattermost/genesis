// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package main

import (
	"net/url"

	"github.com/mattermost/genesis/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	securityCmd.PersistentFlags().String("server", defaultLocalServerAPI, "The genesis server whose API will be queried.")

	securityAccountCmd.PersistentFlags().String("account", "", "The id of the account.")
	securityAccountCmd.MarkPersistentFlagRequired("account") //nolint

	securityCmd.AddCommand(securityAccountCmd)
	securityAccountCmd.AddCommand(securityAccountLockAPICmd)
	securityAccountCmd.AddCommand(securityAccountUnlockAPICmd)
}

var securityCmd = &cobra.Command{
	Use:   "security",
	Short: "Manage security locks for different genesis resources.",
}

var securityAccountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage security locks for account resources.",
}

var securityAccountLockAPICmd = &cobra.Command{
	Use:   "api-lock",
	Short: "Lock API changes on a given account",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		if _, err := url.Parse(serverAddress); err != nil {
			return errors.Wrap(err, "provided server address not a valid address")
		}

		client := model.NewClient(serverAddress)

		accountID, _ := command.Flags().GetString("account")
		err := client.LockAPIForAccount(accountID)
		if err != nil {
			return errors.Wrap(err, "failed to lock account API")
		}

		return nil
	},
}

var securityAccountUnlockAPICmd = &cobra.Command{
	Use:   "api-unlock",
	Short: "Unlock API changes on a given account",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		if _, err := url.Parse(serverAddress); err != nil {
			return errors.Wrap(err, "provided server address not a valid address")
		}

		client := model.NewClient(serverAddress)

		accountID, _ := command.Flags().GetString("account")
		err := client.UnlockAPIForAccount(accountID)
		if err != nil {
			return errors.Wrap(err, "failed to unlock account API")
		}

		return nil
	},
}
