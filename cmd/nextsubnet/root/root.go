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
	"math"
	"net"
	"os"

	"github.com/apparentlymart/go-cidr/cidr"
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

		subnetsInUse, err := parseSubnet(fp)
		if err != nil {
			return err
		}

		nextSubnet, err := findNextSubnet(subnetsInUse)
		if err != nil {
			return err
		}

		fmt.Println(nextSubnet)

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

func parseSubnet(fp flagpole) ([]*net.IPNet, error) {
	// TODO return the first in case ignore-file and ignore-list is not provided
	if fp.ignoreFile != "" {
		return ns.IgnoreFileParse(fp.ignoreFile)
	}

	if fp.ignoreList != "" {
		return ns.IgnoreListParse(fp.ignoreList)
	}

	// When no --ignore-file or --ignore-list is passed, return an empty array
	return []*net.IPNet{}, nil
}

func findNextSubnet(subnetsInUse []*net.IPNet) (*net.IPNet, error) {
	// TODO generate possible net values
	netMaskSize, _ := fp.network.Mask.Size()
	maskDiff := fp.subnetMask - netMaskSize
	subnetCapacity := math.Pow(2, float64(maskDiff))

	for i := 0; i < int(subnetCapacity); i++ {

		subnetCandidate, err := cidr.Subnet(&fp.network, maskDiff, i)
		if err != nil {
			return nil, err
		}

		// TODO when subnetsInUse contains two values that overlaps or a value that is not
		// in the range of the network it will run for
		// all subnetCandidates regardless and return a erro for each one of them. Better to
		// fail fast before
		// Every candidate is presummably to return an error until a subtible subnet is found
		// When that is not the case, the flow will reach this point and break out of the loop
		// with the nextsubnet
		err = cidr.VerifyNoOverlap(append(subnetsInUse, subnetCandidate), &fp.network)
		if err == nil {
			return subnetCandidate, nil
		}
		continue
	}
	return nil, fmt.Errorf("in findNextSubnet: no subnet found")
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
