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
	"fmt"
	_ "image/png"
	"os"
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
	Command string              `json:"Command"`
	Flags   map[string]FlagInfo `json:"Flags"`
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
		Command: cmd.CommandPath(),
		Flags:   make(map[string]FlagInfo),
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

	// Create a filename that matches the command path
	fname := strings.ReplaceAll(c.Command, " ", "_")
	fileName := fmt.Sprintf("spec/%s_spec.sh", fname)
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

	log.Info().Msgf("Rendered %s", fileName)
	return nil
}
