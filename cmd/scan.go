/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cmd

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/stmcginnis/gofish"
	"golang.org/x/sync/semaphore"
)

// scanCmd represents the scan command
var scanCmd = &cobra.Command{
	Use:   "scan HOST",
	Short: "Scan a host or network to detect hardware.",
	Long:  `Scan a host or network to detect hardware.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		getRedfishData()
	},
}

var (
	username string
	password string
	insecure bool
	port     int
	host     string
	cidr     string
	portscan bool
)

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.PersistentFlags().StringVarP(&host, "host", "H", "", "Host or IP of the BMC")
	scanCmd.MarkPersistentFlagRequired("host")

	scanCmd.PersistentFlags().StringVarP(&username, "username", "U", "admin", "Username for the BMC")
	scanCmd.MarkPersistentFlagRequired("username")

	scanCmd.PersistentFlags().StringVarP(&password, "password", "P", "", "Password for the BMC")
	scanCmd.MarkPersistentFlagRequired("password")

	scanCmd.PersistentFlags().IntVarP(&port, "port", "p", 443, "Port number for the BMC")
	scanCmd.MarkPersistentFlagRequired("port")

	scanCmd.MarkFlagsRequiredTogether("host", "username", "password")

	scanCmd.PersistentFlags().BoolVarP(&insecure, "insecure", "k", true, "Do not enforce certificate validation")
	// scanCmd.PersistentFlags().StringVarP(&cidr, "cidr", "C", "", "CIDR to scan for BMCs")
	// scanCmd.MarkFlagsMutuallyExclusive("host", "cidr")
	// scanCmd.PersistentFlags().BoolVarP(&portscan, "portscan", "s", false, "Scan targets for common BMC ports")

}

var (
	rfclient gofish.ClientConfig
)

type PortScan struct {
	// ip to scan
	ip string
	// threshold that will limit the number of go routines that will be running at any given time
	lock *semaphore.Weighted
}

// Ulimit gets the max number of open files for the current user
func Ulimit() int64 {
	stdout, _, err := shell("ulimit", []string{"-n"})
	if err != nil {
		panic(err)
	}

	s := strings.TrimSpace(string(stdout))

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		panic(err)
	}

	return i
}

// ScanPort scans a single port on a given ip address
func ScanPort(ip string, port int, timeout time.Duration) {
	target := fmt.Sprintf("%s:%d", ip, port)
	// error when an operation didn’t complete within the allotted timeout value
	// This may happen for various reasons including a DROP firewall rule where the target doesn’t explicitly tell you a port isn’t open from the scanner’s perspective.
	// Check for this error since we really don’t want to wait around forever to get nothing
	conn, err := net.DialTimeout("tcp", target, timeout)

	if err != nil {
		// account for an error that will occur when trying to connect to a port, but too many connections (files) are already open and need to schedule the execution to happen again
		if strings.Contains(err.Error(), "too many open files") {
			time.Sleep(timeout)
			ScanPort(ip, port, timeout)
			// account for an error that will occur when trying to connect to a port, but the port is closed
		} else {
			fmt.Println(port, "closed")
		}
		return
	}

	conn.Close()
	fmt.Println(port, "open")
}

// Start will start scanning ports on a given ip address
func (ps *PortScan) Start(f, l int, timeout time.Duration) {
	wg := sync.WaitGroup{}
	defer wg.Wait()

	for port := f; port <= l; port++ {
		ps.lock.Acquire(context.TODO(), 1)
		wg.Add(1)
		go func(port int) {
			defer ps.lock.Release(1)
			defer wg.Done()
			ScanPort(ps.ip, port, timeout)
		}(port)
	}
}

// scanPorts scans a range of ports
func scanPorts() {
	if host != "" {
		ps := &PortScan{
			ip:   host,
			lock: semaphore.NewWeighted(Ulimit()),
		}
		ps.Start(1, 65535, 500*time.Millisecond)
	} else if cidr != "" {
		_, ipnet, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(err)
		}

		for ip := ipnet.IP; ipnet.Contains(ip); {
			ps := &PortScan{
				ip:   ip.String(),
				lock: semaphore.NewWeighted(Ulimit()),
			}
			ps.Start(1, 65535, 500*time.Millisecond)
		}
	}

}

// getRedfishData gets the Redfish data from the BMC
func getRedfishData() {
	endpoint := fmt.Sprintf("https://%s:%d", host, port)
	log.Print(endpoint)
	rfclient.Endpoint = endpoint
	rfclient.Username = username
	rfclient.Password = password
	rfclient.Insecure = insecure
	// if debug {
	// 	rfclient.DumpWriter = os.Stdout
	// }

	c, err := gofish.Connect(rfclient)
	// c, err := gofish.ConnectDefault(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	service := c.Service
	chassis, err := service.Chassis()
	if err != nil {
		log.Fatalln(err)
	}
	// b := []byte{}

	for _, chass := range chassis {
		fmt.Printf("Manufacturer/model: %#v %#v\n", chass.Manufacturer, chass.Model)
		fmt.Printf("Serial/part: %#v %#v\n\n", chass.SerialNumber, chass.PartNumber)
		// err := chass.UnmarshalJSON(b)
		// if err != nil {
		// 	panic(err)
		// }
		// fmt.Printf("%+v\n", b)
	}
}
