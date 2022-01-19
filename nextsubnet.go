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
	"net"
	"os"
	"strings"
)

// ignoreListParse receives a comma separated list of subnets and
// returns a slice of net.IPNet
func IgnoreListParse(ignoreList string) ([]*net.IPNet, error) {
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

// ignoreFileParse receives file path containing a list of
// subnets in CIDR format and returns a slice of *net.IPNet
func IgnoreFileParse(ignoreFile string) ([]*net.IPNet, error) {
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
