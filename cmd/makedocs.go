/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
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
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

// MakeDocsCmd represents the makedocs command
var MakeDocsCmd = &cobra.Command{
	Use:          "makedocs",
	Short:        "Generate markdown docs to ./docs/",
	Long:         `Generate markdown docs to ./docs/`,
	Args:         cobra.NoArgs,
	SilenceUsage: true, // Errors are more important than the usage
	RunE:         makeDocs,
}

// makeDocs generates mardown docs for all cobra commands
func makeDocs(cmd *cobra.Command, args []string) error {
	err := os.MkdirAll("docs/commands", fs.FileMode(0755))
	if err != nil {
		return err
	}
	err = doc.GenMarkdownTree(RootCmd, "docs/commands")
	if err != nil {
		return err
	}
	return nil
}
