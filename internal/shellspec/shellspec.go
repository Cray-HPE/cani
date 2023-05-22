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

package shellspec

import (
	"bytes"
	"fmt"
	_ "image/png"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type FlagInfo struct {
	Name              string      `json:"name"`
	DataType          string      `json:"data_type"`
	Required          bool        `json:"required"`
	Persistent        bool        `json:"persistent"`
	MutuallyExclusive interface{} `json:"mutually_exclusive"`
	MutuallyRequired  interface{} `json:"mutually_required"`
}

type CommandInfo struct {
	CobraCmd *cobra.Command      `json:"-"`
	Command  string              `json:"Command"`
	Flags    map[string]FlagInfo `json:"Flags"`
}

// A struct to hold key-value pairs for the template
type CommandEntry struct {
	Command string
	Info    CommandInfo
}

func getFlagInfo(flag *pflag.Flag) FlagInfo {
	// FIXME: Update this function to populate the FlagInfo struct correctly
	return FlagInfo{
		Name:              flag.Name,
		DataType:          "string", // You may need to determine the data type of the flag
		Required:          false,    // You may need to add custom logic to determine if a flag is required
		Persistent:        false,    // You may need to add custom logic to determine if a flag is persistent
		MutuallyExclusive: nil,
		MutuallyRequired:  nil,
	}
}

func GetCommandInfo(cmd *cobra.Command) map[string]CommandInfo {
	commandMap := make(map[string]CommandInfo)

	cmdInfo := CommandInfo{
		CobraCmd: cmd,
		Command:  cmd.CommandPath(),
		Flags:    make(map[string]FlagInfo),
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		cmdInfo.Flags[flag.Name] = getFlagInfo(flag)
	})

	cmd.Root().PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		cmdInfo.Flags[flag.Name] = getFlagInfo(flag)
	})

	commandMap[cmd.CommandPath()] = cmdInfo

	for _, subCmd := range cmd.Commands() {
		subCommandMap := GetCommandInfo(subCmd)
		for k, v := range subCommandMap {
			commandMap[k] = v
		}
	}

	return commandMap
}

func joinStrings(sep string, s ...string) string {
	return strings.Join(s, sep)
}

func replaceString(input, old, new string) string {
	return strings.ReplaceAll(input, old, new)
}

func (c CommandInfo) GenerateSpecfile() error {
	// make a new template with some name
	tmpl := template.New("spec.sh.j2")

	tmpl.Funcs(template.FuncMap{
		"joinStrings":   joinStrings,
		"replaceString": replaceString,
	})

	// parse the template file
	tmpl.ParseFiles("testdata/templates/spec.sh.j2")

	// Craft a filename that matches the command path
	fname := strings.ReplaceAll(c.Command, " ", "_")
	fileName := fmt.Sprintf("spec/%s_spec.sh", fname)

	// make the directory if it does not exist
	err := os.MkdirAll(filepath.Dir(fileName), 0755)
	if err != nil {
		panic(err)
	}

	// Create the file
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Execute the template and write the output to the file
	err = tmpl.Execute(file, c)
	if err != nil {
		panic(err)
	}

	// Ensure all the data is written to disk
	err = file.Sync()
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Rendered spec file for '%s' to %s", c.Command, fileName)
	return nil
}

func (c CommandInfo) GenerateHelpFixtures() error {
	// capture the --help flag for each cobra subcommand into a bytes.Buffer
	buf := new(bytes.Buffer)
	// Set the output to the buffer
	c.CobraCmd.SetOut(buf)

	// Run the help command, capturing its output in buf
	err := c.CobraCmd.Help()
	if err != nil {
		panic(err)
	}
	output := buf.String()
	// helpMessage := fmt.Sprintf("  -h, --help                   help for %s", c.Command)
	// munged := insertHelpAlphabetically(output, helpMessage)
	// Craft a 'help' filename at a path that matches the command name
	fname := strings.ReplaceAll(c.Command, " ", "/")
	fileName := fmt.Sprintf("testdata/fixtures/%s/help", fname)

	// make the directory if it does not exist
	err = os.MkdirAll(filepath.Dir(fileName), 0755)
	if err != nil {
		panic(err)
	}

	// Create the file
	file, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write the string to the file
	_, err = file.WriteString(output)
	if err != nil {
		panic(err)
	}

	// Ensure all the data is written to disk
	err = file.Sync()
	if err != nil {
		panic(err)
	}

	log.Info().Msgf("Saved --help output for '%s' to %s", c.Command, fileName)
	return nil
}

// insertHelpAlphabetically inserts the help message into the help output
// TODO: Make this work properly
// By default, the Cobra library doesn't include the help command (--help) in the output of the Help() function.
// This is because the Help() function is intended to display the help message, so it's understood that --help has been triggered.
// The fixtures need the --help flag to match the output of the command, so we need to insert it into the help output.
func insertHelpAlphabetically(helpOutput string, helpMessage string) string {
	lines := strings.Split(helpOutput, "\n")

	var flagLines []string
	for _, line := range lines {
		if strings.HasPrefix(line, "  -") || strings.HasPrefix(line, "      --") {
			flagLines = append(flagLines, line)
		}
	}

	flagLines = append(flagLines, helpMessage)

	sort.Strings(flagLines)

	newHelpOutput := strings.Join(lines[:len(lines)-len(flagLines)], "\n") + "\n" + strings.Join(flagLines, "\n")
	return newHelpOutput
}
