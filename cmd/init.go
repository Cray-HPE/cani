/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package cmd

import (
	"github.com/Cray-HPE/cani/cmd/add"
	"github.com/Cray-HPE/cani/cmd/alpha"
	"github.com/Cray-HPE/cani/cmd/batch"
	"github.com/Cray-HPE/cani/cmd/classify"
	"github.com/Cray-HPE/cani/cmd/export"
	imprt "github.com/Cray-HPE/cani/cmd/import"
	initcmd "github.com/Cray-HPE/cani/cmd/init"
	"github.com/Cray-HPE/cani/cmd/remove"
	"github.com/Cray-HPE/cani/cmd/serve"
	"github.com/Cray-HPE/cani/cmd/show"
	"github.com/Cray-HPE/cani/cmd/update"
	"github.com/Cray-HPE/cani/internal/cli"
)

var (
	rootCmd *cli.Command
	cfgFile string
)

func Init() {
	// initialize the process-wide root command
	rootCmd = newRootTree()
}

// newRootTree assembles a complete command tree (root plus all verbs) and
// returns the root. Init uses it for the process-wide rootCmd; the batch runner
// uses it to build a fresh tree for each line it dispatches, so re-parsed flags
// never accumulate state across commands.
func newRootTree() *cli.Command {
	root := newRootCommand()

	// add the init command at the root level (not under alpha)
	root.AddCommand(initcmd.NewCommand())

	// at present, all other commands are under the alpha command since this is
	// still a work in progress
	alphaCmd := alpha.NewCommand()
	root.AddCommand(alphaCmd)
	alphaCmd.AddCommand(
		imprt.NewCommand(),
		add.NewCommand(),
		remove.NewCommand(),
		show.NewCommand(),
		serve.NewCommand(),
		export.NewCommand(),
		update.NewCommand(),
		classify.NewCommand(),
		batch.NewCommand(dispatchLine),
	)
	return root
}

// dispatchLine runs a single batch command line against a freshly built command
// tree so per-line flag parsing never shares state with other lines.
func dispatchLine(args []string) error {
	return newRootTree().ExecuteArgs(args)
}
