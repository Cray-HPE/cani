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
