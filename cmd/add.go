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
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	client "github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"

	"github.com/Cray-HPE/cani/cmd/blade"
	"github.com/Cray-HPE/cani/cmd/cabinet"
	"github.com/Cray-HPE/cani/cmd/hsn"
	"github.com/Cray-HPE/cani/cmd/node"
	"github.com/Cray-HPE/cani/cmd/pdu"
	sw "github.com/Cray-HPE/cani/cmd/switch"
)

// addCmd represents the switch add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add assets to the inventory.",
	Long:  `Add assets to the inventory.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		if simulation {
			blade.AddBladeCmd.SetArgs([]string{"-S"})
		}
	},
}

func init() {
	addCmd.AddCommand(blade.AddBladeCmd)
	addCmd.AddCommand(cabinet.AddCabinetCmd)
	addCmd.AddCommand(hsn.AddHsnCmd)
	addCmd.AddCommand(node.AddNodeCmd)
	addCmd.AddCommand(pdu.AddPduCmd)
	addCmd.AddCommand(sw.AddSwitchCmd)
}

// CreateNewContainer creates a container from an image
func CreateNewContainer(image string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		fmt.Println("Unable to create container client")
		panic(err)
	}

	cont, err := cli.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: image,
		},
		&container.HostConfig{},
		&network.NetworkingConfig{},
		&v1.Platform{},
		image)
	if err != nil {
		panic(err)
	}

	cli.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	fmt.Printf("Container %s is started\n", cont.ID)
	cli.ContainerRemove(context.Background(), cont.ID, types.ContainerRemoveOptions{})
	return cont.ID, nil
}

// StopContainer stops a running container
func StopContainer(containerID string) error {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	err = cli.ContainerStop(context.Background(), containerID, container.StopOptions{})
	if err != nil {
		panic(err)
	}
	return err
}

// ListContainers lists all running containers
func ListContainers() error {
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	if len(containers) > 0 {
		for _, container := range containers {
			fmt.Printf("Container ID: %s", container.ID)
		}
	} else {
		fmt.Println("There are no containers running")
	}
	return nil
}
