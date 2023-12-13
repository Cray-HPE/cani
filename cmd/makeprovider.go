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
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/Cray-HPE/cani/internal/inventory"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// MakeProviderCmd represents the makeprovider command
var MakeProviderCmd = &cobra.Command{
	Use:    "makeprovider PKG_NAME DIR",
	Short:  "Generate provider package stubs to internal/provider",
	Long:   `Generate provider package stubs to internal/provider`,
	Args:   cobra.MinimumNArgs(2),
	RunE:   makeProvider,
	Hidden: true, // mostly for dev and plugin authors
}

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")
)

// makeProvider generates go files that conform to the interface constraint for a provider
func makeProvider(cmd *cobra.Command, args []string) error {
	provider := args[0]
	dir := args[1]

	// make a path with the dir and the provider name
	pdir := filepath.Join(dir, provider)
	err := os.MkdirAll(pdir, fs.FileMode(0755))
	if err != nil {
		return err
	}

	// generate the stubs
	err = generateStubs(pdir)
	if err != nil {
		return err
	}
	return nil
}

// generateStubs generates code to statisfy the interface constraints
func generateStubs(dir string) (err error) {
	// get the absolute path
	path, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	// the package name is the provider name, which is the dir where it will end up
	pkgname := filepath.Base(path)

	//  create a go file with the provider name for the provider struct
	err = pkgStub(dir, pkgname)
	if err != nil {
		return err
	}

	//  create a go file for init()
	err = initStub(dir, pkgname)
	if err != nil {
		return err
	}

	// make a fake type that fulfills the interface
	s := Stub{}
	t := reflect.TypeOf(&s).Elem()
	for i := 0; i < t.NumMethod(); i++ {
		method := t.Method(i)

		// convert to a useful filename
		ml := toSnakeCase(method.Name)
		fname := fmt.Sprintf("%s%s", ml, ".go")
		filename := filepath.Join(path, fname)

		log.Info().Msgf("Generating stubs: %+v", filename)

		// make the file
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer f.Close()

		// write a file with the package name
		payload := fmt.Sprintf(`package %s
`, pkgname)
		if _, err := io.WriteString(f, payload); err != nil {
			return err
		}
	}
	return nil
}

func initStub(dir, pkgname string) error {
	iStub := filepath.Join(dir, "init.go")
	f, err := os.Create(iStub)
	if err != nil {
		return err
	}
	defer f.Close()

	// this is what gets written
	payload := fmt.Sprintf(`package %s

import "github.com/spf13/cobra"

func NewSessionInitCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	// Session init flags
	// cmd.Flags().BoolP("myflag", "m", false, "My flag")

	return cmd, nil
}

func NewAddCabinetCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}

func UpdateAddCabinetCommand(caniCmd *cobra.Command) error {
	return nil
}

func NewAddNodeCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}

func NewUpdateNodeCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}

func UpdateUpdateNodeCommand(caniCmd *cobra.Command) error {

	return nil
}

func NewExportCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}

func NewImportCommand() (cmd *cobra.Command, err error) {
	cmd = &cobra.Command{}

	return cmd, nil
}
		`, pkgname)
	if _, err := io.WriteString(f, payload); err != nil {
		return err
	}
	return nil
}

func pkgStub(dir, pkgname string) error {
	//  create a go file with the provider name for the provider struct
	pStub := filepath.Join(dir, pkgname+".go")
	f, err := os.Create(pStub)
	if err != nil {
		return err
	}
	defer f.Close()

	// this is what gets written
	payload := fmt.Sprintf(`package %s
	
	type %s struct {}
	
	// options for %s
	type %sOpts struct {}
		`, pkgname, strings.Title(pkgname), pkgname, strings.Title(pkgname))
	if _, err := io.WriteString(f, payload); err != nil {
		return err
	}
	return nil
}

func toSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

type Stub struct{}

func (s Stub) ValidateExternal(cmd *cobra.Command, args []string) error {
	return nil
}
func (s Stub) ValidateInternal(cmd *cobra.Command, args []string, datastore inventory.Datastore, enableRequiredDataChecks bool) (map[uuid.UUID]provider.HardwareValidationResult, error) {
	return map[uuid.UUID]provider.HardwareValidationResult{}, nil
}
func (s Stub) ImportInit(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	return nil
}
func (s Stub) Import(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	return nil
}
func (s Stub) Export(cmd *cobra.Command, args []string, datastore inventory.Datastore) error {
	return nil
}
func (s Stub) Reconcile(cmd *cobra.Command, args []string, datastore inventory.Datastore, dryrun bool, ignoreExternalValidation bool) error {
	return nil
}
func (s Stub) RecommendHardware(inv inventory.Inventory, cmd *cobra.Command, args []string, auto bool) (recommended provider.HardwareRecommendations, err error) {
	return recommended, nil
}
func (s Stub) SetProviderOptions(cmd *cobra.Command, args []string) error {
	return nil
}
func (s Stub) GetProviderOptions() (interface{}, error) {
	opts := map[string]interface{}{}
	return opts, nil
}
func (s Stub) BuildHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string, recommendations provider.HardwareRecommendations) error {
	return nil
}
func (s Stub) NewHardwareMetadata(hw *inventory.Hardware, cmd *cobra.Command, args []string) error {
	return nil
}
func (s Stub) GetFields(hw *inventory.Hardware, fieldNames []string) (values []string, err error) {
	return values, nil
}
func (s Stub) SetFields(hw *inventory.Hardware, values map[string]string) (result provider.SetFieldsResult, err error) {
	return result, nil
}
func (s Stub) GetFieldMetadata() ([]provider.FieldMetadata, error) {
	return []provider.FieldMetadata{}, nil
}
func (s Stub) ListCabinetMetadataColumns() (columns []string) {
	return nil
}
func (s Stub) ListCabinetMetadataRow(inventory.Hardware) (values []string, err error) {
	return values, nil
}
func (s Stub) PrintHardware(hw *inventory.Hardware) {
}
