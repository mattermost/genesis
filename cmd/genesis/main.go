// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

// Package main is the entry point to the Mattermost Genesis server and CLI.
package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "genesis",
	Short: "Genesis is a tool to provision, manage, and monitor Cloud Enterprise resources.",
	Run: func(cmd *cobra.Command, args []string) {
		serverCmd.RunE(cmd, args) //nolint
	},
	// SilenceErrors allows us to explicitly log the error returned from rootCmd below.
	SilenceErrors: true,
}

func init() {
	rootCmd.MarkFlagRequired("database") //nolint

	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(accountCmd)
	rootCmd.AddCommand(schemaCmd)
	rootCmd.AddCommand(webhookCmd)
	rootCmd.AddCommand(securityCmd)
	rootCmd.AddCommand(parentSubnetCmd)
	rootCmd.AddCommand(subnetCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logger.WithError(err).Error("command failed")
		os.Exit(1)
	}
}
