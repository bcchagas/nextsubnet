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

// Package nextsubnet contains functions to help find the next subnet
// available in a network
package nextsubnet

import (
	"bufio"
	"math"
	"net"
	"os"
	"strings"

	"github.com/apparentlymart/go-cidr/cidr"
)

type NextSubnet struct {
	SubnetMask      int
	Network         net.IPNet
	SubnetsIPNet    []*net.IPNet
	SubnetsStr      string
	SubnetsFilePath string
}

func (ns NextSubnet) Find() (*net.IPNet, error) {

	for i := 0; i < int(ns.subnetCapacity()); i++ {

		subnetCandidate, err := cidr.Subnet(&ns.Network, ns.subnetNewBits(), i)
		if err != nil {
			return nil, err
		}

		// TODO when subnetsInUse contains two values that overlaps or a value that is not
		// in the range of the network it will run for
		// all subnetCandidates regardless and return a erro for each one of them. Better to
		// fail fast before
		// Every candidate will presummably return an error until a subtible subnet is found
		// When that is not the case, the flow will reach this point and break out of the loop
		// with the nextsubnet
		subnetsInUse, err := ns.getSubnets()
		if err != nil {
			return nil, err
		}

		err = cidr.VerifyNoOverlap(append(subnetsInUse, subnetCandidate), &ns.Network)
		if err == nil {
			return subnetCandidate, nil
		}

		continue
	}

	// a subnet was not found
	return nil, nil
}

func (ns NextSubnet) subnetCapacity() float64 {
	return math.Pow(2, float64(ns.subnetNewBits()))
}

func (ns NextSubnet) subnetNewBits() int {
	netMaskSize, _ := ns.Network.Mask.Size()
	return ns.SubnetMask - netMaskSize
}

func (ns NextSubnet) getSubnets() ([]*net.IPNet, error) {

	if ns.SubnetsIPNet != nil {
		return ns.SubnetsIPNet, nil
	}

	// TODO return the first in case ignore-file and ignore-list is not provided
	if ns.SubnetsFilePath != "" {
		return ignoreFileParse(ns.SubnetsFilePath)
	}

	if ns.SubnetsStr != "" {
		return ignoreListParse(ns.SubnetsStr)
	}

	// When no --ignore-file or --ignore-list is passed, return an empty array
	return []*net.IPNet{}, nil
}

// TODO Rename ignoreFileParse to something that makes sense
// for go clients users

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

// TODO Rename ignoreFileParse to something that makes sense
// for go clients users

// ignoreFileParse receives file path containing a list of
// subnets in CIDR format and returns a slice of *net.IPNet
func ignoreFileParse(ignoreFile string) ([]*net.IPNet, error) {
	f, err := os.Open(ignoreFile)
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
