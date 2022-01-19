/*
Copyright © 2022 Bruno Chagas <bcchagas@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package root

import (
	"fmt"
	"log"
	"net"
	"os"

	ns "github.com/bcchagas/nextsubnet"
	"github.com/spf13/cobra"
)

type flagpole struct {
	network    net.IPNet
	subnetMask int
	ignoreList string
	ignoreFile string
}

var fp flagpole = flagpole{}

var rootCmd = &cobra.Command{
	Use:   "nextsubnet -n network -m mask [--ignore-list list | --ignore-file file]",
	Short: "Find the next subnet available for a network",
	Long:  "Find the next subnet available for a network",
	Example: `  # Find the next /24 subnet in the network 10.0.0.1/22 
  # that doesn't overlap any of the two existent subnets
  nextsubnet --network 10.0.0.1/22 --subnet-mask 25 --ignore-list 10.0.0.0/24,10.0.1.128/25

  # You can also pass in a file containing the subnets in use
  nextsubnet --network 10.0.0.1/22  --subnet-mask 24 --ignore-file subnets.txt`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if err := validateFlags(fp); err != nil {
			return err
		}

		subnet, err := ns.NextSubnet{
			SubnetMask:      fp.subnetMask,
			Network:         fp.network,
			SubnetsStr:      fp.ignoreList,
			SubnetsFilePath: fp.ignoreFile,
		}.FindNextSubnet()

		// TODO create a custom error to check if it's a not found error
		// and clean up the next session that checks if subnet is != nil
		if err != nil {
			return err
		}

		if subnet != nil {
			fmt.Println(subnet)
		} else {
			log.Println("Not found")
			os.Exit(1)
		}

		return nil
	},
}

func init() {

	rootCmd.Flags().IPNetVarP(&fp.network, "network", "n", net.IPNet{},
		"(Required) Address of the network the subnet will be based of in CIDR notation e.g. 10.0.0.0/22")
	rootCmd.Flags().IntVarP(&fp.subnetMask, "subnet-mask", "m", 0,
		"(Required) Mask size of the subnet to be found e.g. 24")
	rootCmd.Flags().StringVar(&fp.ignoreList, "ignore-list", "",
		"List of subnets in CIDR notation separated by comma e.g. '10.0.0.0/24,10.0.1.128/25'")
	rootCmd.Flags().StringVar(&fp.ignoreFile, "ignore-file", "",
		"File containing the list of subnets in CIDR notation separated by line")

	rootCmd.MarkFlagRequired("subnet-mask")
	rootCmd.MarkFlagRequired("network")

}

// TODO this validation shoud be in the nextsubnet package
func validateFlags(fp flagpole) error {
	if fp.ignoreList != "" && fp.ignoreFile != "" {
		return fmt.Errorf("--ignore-list and --ignore-file specified")
	}

	// check mask is lower than the network block
	if netMaskSize, _ := fp.network.Mask.Size(); netMaskSize >= fp.subnetMask {
		return fmt.Errorf(
			"--subnet-mask %v must be greater than --network %v",
			fp.subnetMask,
			fp.network.String(),
		)
	}
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
