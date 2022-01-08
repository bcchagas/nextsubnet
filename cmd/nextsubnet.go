/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// nextsubnetCmd represents the nextsubnet command
var nextsubnetCmd = &cobra.Command{
	Use:   "nextsubnet -n network -m mask [--ignore-list list | --ignore-file file]",
	Short: "Find the next subnet available for a network",
	Long:  `Find the next subnet available for a network.`,
	Example: "  # Find the next /24 subnet in the network 10.0.0.1/22 that doesn't overlap any of the two existent subnets\n" +
		"  bccli nextsubnet --network 10.0.0.1/22 --mask 24 --ignore-list \"10.0.0.1/24,10.0.0.2/25\"\n\n" +
		"  # You can also pass in a file containing the subnets in use\n" +
		"  bccli nextsubnet --network 10.0.0.1/22  --mask 24 --ignore-file subnets.txt\n",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(nextsubnetFlags.network)
		fmt.Println(nextsubnetFlags.mask)
		fmt.Println(nextsubnetFlags.ignoreList)
		fmt.Println(nextsubnetFlags.IgnoreFile)
	},
}

var nextsubnetFlags struct {
	network    string
	mask       int
	ignoreList string
	IgnoreFile string
}

func init() {
	rootCmd.AddCommand(nextsubnetCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// nextsubnetCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	nextsubnetCmd.Flags().StringVarP(&nextsubnetFlags.network, "network", "n", "", "(Required) Address of the network the subnet will be based of in CIDR notation e.g. 10.0.0.0/24")
	nextsubnetCmd.MarkFlagRequired("network")

	nextsubnetCmd.Flags().IntVarP(&nextsubnetFlags.mask, "mask", "m", 0, "(Required) Mask of the subnet to be found e.g. 24 for a '/24' subnet")
	nextsubnetCmd.MarkFlagRequired("mask")

	nextsubnetCmd.Flags().StringVar(&nextsubnetFlags.ignoreList, "ignore-list", "", "List of subnets in CIDR notation separated by comma e.g. '10.0.0.0/24,10.0.0.1/24'")
	nextsubnetCmd.Flags().StringVar(&nextsubnetFlags.IgnoreFile, "ignore-file", "", "File containing the list of subnets in CIDR notation separated by line")

}
