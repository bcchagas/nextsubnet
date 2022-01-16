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
package cmd

import (
	"bufio"
	"fmt"
	"math"
	"net"
	"os"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/spf13/cobra"
)

// rootCmd represents the nextsubnet command
var rootCmd = &cobra.Command{
	Use:   "nextsubnet -n network -m mask [--ignore-list list | --ignore-file file]",
	Short: "Find the next subnet available for a network",
	Long:  `Find the next subnet available for a network.`,
	Example: `  # Find the next /24 subnet in the network 10.0.0.1/22 
  # that doesn't overlap any of the two existent subnets
  nextsubnet --network 10.0.0.1/22 --subnet-mask 24 --ignore-list 10.0.0.1/24,10.0.0.2/25

  # You can also pass in a file containing the subnets in use
  nextsubnet --network 10.0.0.1/22  --subnet-mask 24 --ignore-file subnets.txt`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if flags.ignoreList != "" && flags.ignoreFile != "" {
			return fmt.Errorf("--ignore-list and --ignore-file specified")
		}

		// check mask is lower than the network block
		if netMaskSize, _ := flags.network.Mask.Size(); netMaskSize >= flags.subnetMask {
			return fmt.Errorf("--subnet-mask %v must be greater than --network %v", flags.subnetMask, flags.network.String())
		}

		subnetsInUse := make([]*net.IPNet, 0)
		var err error

		// TODO return the first in case ignore-file and ignore-list is not provided
		// TODO read from ignore-file and process it
		if flags.ignoreFile != "" {
			subnetsInUse, err = ignoreFileParse(flags.ignoreFile)
			if err != nil {
				return err
			}
		}

		// TODO process ignore-list
		if flags.ignoreList != "" {
			subnetsInUse, err = ignoreListParse(flags.ignoreList)
			if err != nil {
				return err
			}
		}

		// TODO generate possible net values
		netMaskSize, _ := flags.network.Mask.Size()
		maskDiff := flags.subnetMask - netMaskSize
		subnetCapacity := math.Pow(2, float64(maskDiff))

		// maybe the size wont work for ipv6 (or ipv4 with lower networks)
		subnetCandidates := make([]*net.IPNet, int(subnetCapacity))
		for i := 0; i < len(subnetCandidates); i++ {

			subnetCandidates[i], err = cidr.Subnet(&flags.network, maskDiff, i)
			if err != nil {
				return err
			}

			// TODO when subnetsInUse contains two values that overlaps or a value that is not
			// in the range of the network it will run for
			// all subnetCandidates regardless and return a erro for each one of them. Better to
			// fail fast before
			err = cidr.VerifyNoOverlap(append(subnetsInUse, subnetCandidates[i]), &flags.network)
			if err != nil {
				//log.Println(err)
				continue
			}

			// Every candidate is presummably to return an error until a subtible subnet is found
			// When that is not the case, the flow will reach this point and break out of the loop
			// with the nextsubnet
			nextsubnet := subnetCandidates[i]
			fmt.Println(nextsubnet)
			break

		}
		return nil
	},
}

var flags struct {
	network    net.IPNet
	subnetMask int
	ignoreList string
	ignoreFile string
}

func init() {

	rootCmd.Flags().IPNetVarP(&flags.network, "network", "n", net.IPNet{},
		"(Required) Address of the network the subnet will be based of in CIDR notation e.g. 10.0.0.0/22")
	rootCmd.Flags().IntVarP(&flags.subnetMask, "subnet-mask", "m", 0,
		"(Required) Mask size of the subnet to be found e.g. 24")
	rootCmd.Flags().StringVar(&flags.ignoreList, "ignore-list", "",
		"List of subnets in CIDR notation separated by comma e.g. '10.0.0.0/24,10.0.0.1/24'")
	rootCmd.Flags().StringVar(&flags.ignoreFile, "ignore-file", "",
		"File containing the list of subnets in CIDR notation separated by line")

	rootCmd.MarkFlagRequired("subnet-mask")
	rootCmd.MarkFlagRequired("network")

}

// ignoreListParse receives a comma separated list of subnets and
// returns a slice of net.IPNet
func ignoreListParse(ignoreList string) ([]*net.IPNet, error) {
	var tmpIPNetSlice []*net.IPNet
	sliceOfStrings := strings.Split(ignoreList, ",")
	for _, v := range sliceOfStrings {
		_, tmpIPNet, err := net.ParseCIDR(v)
		if err != nil {
			return nil, err
		}
		tmpIPNetSlice = append(tmpIPNetSlice, tmpIPNet)
	}
	return tmpIPNetSlice, nil
}

// ignoreFileParse receives a path of file containing
// a list of subnets in CIDR format and returns a slice
// containing *net.IPNet
func ignoreFileParse(ignoreFile string) ([]*net.IPNet, error) {
	f, err := os.Open(flags.ignoreFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	tmpIPNetSlice := make([]*net.IPNet, 0)
	for scanner.Scan() {
		tmpString := strings.TrimSpace(scanner.Text())
		// Ignore empty lines in the file
		if len(tmpString) == 0 {
			continue
		}
		_, tmpIPNet, err := net.ParseCIDR(tmpString)
		if err != nil {
			return nil, err
		}
		tmpIPNetSlice = append(tmpIPNetSlice, tmpIPNet)
	}
	return tmpIPNetSlice, nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
