// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package main

import (
	"net/url"
	"os"

	"github.com/mattermost/genesis/model"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	parentSubnetCmd.PersistentFlags().String("server", defaultLocalServerAPI, "The genesis server whose API will be queried.")
	parentSubnetCmd.PersistentFlags().Bool("dry-run", false, "When set to true, only print the API request without sending it.")

	parentSubnetAddCmd.Flags().String("cidr", "", "The subnet that will be added in the parent subnet pool.")
	parentSubnetAddCmd.Flags().Int("split-range", 24, "The range that the passed subnet range will be split into.")
	parentSubnetAddCmd.MarkFlagRequired("cidr") //nolint

	parentSubnetGetCmd.Flags().String("subnet", "", "The subnet id to get from the parent subnets.")
	parentSubnetGetCmd.MarkFlagRequired("subnet") //nolint

	parentSubnetListCmd.Flags().Int("page", 0, "The page of subnets to fetch, starting at 0.")
	parentSubnetListCmd.Flags().Int("per-page", 100, "The number of parent subnets to fetch per page.")
	parentSubnetListCmd.Flags().Bool("table", false, "Whether to display the returned parent subnet list in a table or not")

	parentSubnetCmd.AddCommand(parentSubnetAddCmd)
	parentSubnetCmd.AddCommand(parentSubnetListCmd)
	parentSubnetCmd.AddCommand(parentSubnetGetCmd)
}

var parentSubnetCmd = &cobra.Command{
	Use:   "parent-subnet",
	Short: "Manipulate parent subnets managed by the genesis server.",
}

var parentSubnetAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a parent subnet range.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		if _, err := url.Parse(serverAddress); err != nil {
			return errors.Wrap(err, "provided server address not a valid address")
		}

		client := model.NewClient(serverAddress)

		cidr, _ := command.Flags().GetString("cidr")
		splitRange, _ := command.Flags().GetInt("split-range")
		request := &model.AddParentSubnetRequest{
			CIDR:       cidr,
			SplitRange: splitRange,
		}

		dryRun, _ := command.Flags().GetBool("dry-run")
		if dryRun {
			if err := printJSON(request); err != nil {
				return errors.Wrap(err, "failed to print API request")
			}

			return nil
		}

		subnet, err := client.AddParentSubnet(request)
		if err != nil {
			return errors.Wrap(err, "failed to add parent subnet")
		}

		if err = printJSON(subnet); err != nil {
			return errors.Wrap(err, "failed to print parent subnet response")
		}

		return nil
	},
}

var parentSubnetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List parent subnets.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		if _, err := url.Parse(serverAddress); err != nil {
			return errors.Wrap(err, "provided server address not a valid address")
		}

		client := model.NewClient(serverAddress)

		page, _ := command.Flags().GetInt("page")
		perPage, _ := command.Flags().GetInt("per-page")
		parentSubnets, err := client.GetParentSubnets(&model.GetParentSubnetsRequest{
			Page:    page,
			PerPage: perPage,
		})
		if err != nil {
			return errors.Wrap(err, "failed to query parent subnets")
		}

		outputToTable, _ := command.Flags().GetBool("table")
		if outputToTable {
			table := tablewriter.NewWriter(os.Stdout)
			table.SetAlignment(tablewriter.ALIGN_LEFT)
			table.SetHeader([]string{"PARENT SUBNET", "CIDR"})

			for _, subnet := range parentSubnets {
				table.Append([]string{
					subnet.ID,
					subnet.CIDR,
				})
			}
			table.Render()

			return nil
		}

		if err = printJSON(parentSubnets); err != nil {
			return errors.Wrap(err, "failed to print parent subnet response")
		}

		return nil
	},
}

var parentSubnetGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a particular parent subnet.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		if _, err := url.Parse(serverAddress); err != nil {
			return errors.Wrap(err, "provided server address not a valid address")
		}

		client := model.NewClient(serverAddress)

		subnet, _ := command.Flags().GetString("subnet")
		parentSubnet, err := client.GetParentSubnet(subnet)
		if err != nil {
			return errors.Wrap(err, "failed to query parent subnet")
		}
		if parentSubnet == nil {
			return nil
		}

		if err = printJSON(parentSubnet); err != nil {
			return errors.Wrap(err, "failed to print parent subnet response")
		}

		return nil
	},
}
