package main

import (
	"fmt"
	"os"

	"github.com/Cray-HPE/cani/cmd"
	_ "github.com/Cray-HPE/cani/pkg/devicetypes"
	_ "github.com/Cray-HPE/cani/pkg/provider/csm"
	_ "github.com/Cray-HPE/cani/pkg/provider/example"
	_ "github.com/Cray-HPE/cani/pkg/provider/nautobot"
	_ "github.com/Cray-HPE/cani/pkg/provider/ochami"
	_ "github.com/Cray-HPE/cani/pkg/provider/redfish"

	"github.com/Cray-HPE/cani/internal/cli"
)

func main() {
	dir := "docs/commands"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	if err := os.RemoveAll(dir); err != nil {
		fmt.Fprintf(os.Stderr, "error removing directory %s: %v\n", dir, err)
		os.Exit(1)
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory %s: %v\n", dir, err)
		os.Exit(1)
	}

	cmd.Init()
	root := cmd.RootCommand()

	if err := cli.GenMarkdownTree(root, dir); err != nil {
		fmt.Fprintf(os.Stderr, "error generating docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("command docs generated in %s\n", dir)
}
