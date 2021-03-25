// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"os"

	"github.com/mattermost/genesis/model"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	subnetCmd.PersistentFlags().String("server", defaultLocalServerAPI, "The genesis server whose API will be queried.")
	subnetCmd.PersistentFlags().Bool("dry-run", false, "When set to true, only print the API request without sending it.")

	subnetGetCmd.Flags().String("subnet", "", "The subnet range to get from the subnet pool.")
	subnetGetCmd.MarkFlagRequired("subnet")

	subnetListCmd.Flags().Int("page", 0, "The page of subnets to fetch, starting at 0.")
	subnetListCmd.Flags().Int("per-page", 100, "The number of subnets to fetch per page.")
	subnetListCmd.Flags().Bool("free-subnets", false, "When set to true only available subnets are returned .")
	subnetListCmd.Flags().Bool("table", false, "Whether to display the returned subnet list in a table or not")

	subnetCmd.AddCommand(subnetListCmd)
	subnetCmd.AddCommand(subnetGetCmd)
}

var subnetCmd = &cobra.Command{
	Use:   "subnet",
	Short: "Manipulate subnets managed by the genesis server.",
}

var subnetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List subnets.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		page, _ := command.Flags().GetInt("page")
		perPage, _ := command.Flags().GetInt("per-page")
		free, _ := command.Flags().GetBool("free-subnets")
		subnets, err := client.GetSubnets(&model.GetSubnetsRequest{
			Page:    page,
			PerPage: perPage,
			Free:    free,
		})
		if err != nil {
			return errors.Wrap(err, "failed to query subnets")
		}

		outputToTable, _ := command.Flags().GetBool("table")
		if outputToTable {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"SUBNET", "CIDR", "ACCOUNT ID", "VPC ID", "PARENT SUBNET"})

			for _, subnet := range subnets {
				table.Append([]string{
					subnet.ID,
					subnet.CIDR,
					subnet.AccountID,
					subnet.VPCID,
					subnet.ParentSubnet,
				})
			}
			table.Render()

			return nil
		}

		err = printJSON(subnets)
		if err != nil {
			return errors.Wrap(err, "failed to print subnet response")
		}

		return nil
	},
}

var subnetGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a particular subnet.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		subnet, _ := command.Flags().GetString("subnet")
		sub, err := client.GetSubnet(subnet)
		if err != nil {
			return errors.Wrap(err, "failed to query subnet")
		}
		if sub == nil {
			return nil
		}

		err = printJSON(sub)
		if err != nil {
			return errors.Wrap(err, "failed to print subnet response")
		}

		return nil
	},
}
