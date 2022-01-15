/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"bufio"
	"fmt"
//	"log"
	"math"
	"net"
	"os"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
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
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {

		if flags.ignoreList != "" && flags.ignoreFile != "" {
			return fmt.Errorf("--ignore-list and --ignore-file specified")
		}

		// check mask is lower than the network block
		if netMaskSize, _ := flags.network.Mask.Size(); netMaskSize >= flags.mask {
			return fmt.Errorf("--mask %v must be greater than --network %v", flags.mask, flags.network.String())
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
		maskDiff := flags.mask - netMaskSize //flags.mask is not intuitive that it represents the subnets
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
	mask       int
	ignoreList string
	ignoreFile string
}

func init() {

	rootCmd.AddCommand(nextsubnetCmd)

	nextsubnetCmd.Flags().IPNetVarP(&flags.network, "network", "n", net.IPNet{},
		"(Required) Address of the network the subnet will be based of in CIDR notation e.g. 10.0.0.0/22")
	nextsubnetCmd.Flags().IntVarP(&flags.mask, "mask", "m", 0,
		"(Required) Mask size of the subnet to be found e.g. 24")
	nextsubnetCmd.Flags().StringVar(&flags.ignoreList, "ignore-list", "",
		"List of subnets in CIDR notation separated by comma e.g. '10.0.0.0/24,10.0.0.1/24'")
	nextsubnetCmd.Flags().StringVar(&flags.ignoreFile, "ignore-file", "",
		"File containing the list of subnets in CIDR notation separated by line")

	nextsubnetCmd.MarkFlagRequired("mask")
	nextsubnetCmd.MarkFlagRequired("network")

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
