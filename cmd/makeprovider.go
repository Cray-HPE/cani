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
	"fmt"
	"regexp"

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
	// 	pkg := args[0]
	// 	dir := args[1]

	// 	pdir := filepath.Join(dir, pkg)
	// 	err := os.MkdirAll(pdir, os.ModePerm)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// Use reflection to analyze the methods of the InventoryProvider interface
	// 	inventoryProviderType := reflect.TypeOf((*provider.InventoryProvider)(nil)).Elem()

	// 	// for each method, generate a file
	// 	for i := 0; i < inventoryProviderType.NumMethod(); i++ {
	// 		method := inventoryProviderType.Method(i)
	// 		fileContent := generateStub(method, pkg)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		// prepend the package name
	// 		fileContent = fmt.Sprintf("package %s\n%s", strings.ToLower(pkg), fileContent)

	// 		// convert to a useful filename
	// 		methodFilename := toSnakeCase(method.Name)
	// 		fileName := fmt.Sprintf("%s.go", methodFilename)
	// 		filePath := filepath.Join(pdir, fileName)
	// 		log.Printf("Generating %+v", filePath)
	// 		if err := os.WriteFile(filePath, []byte(fileContent), 0644); err != nil {
	// 			return err
	// 		}

	// 		// After code is genreated, fix format and imports
	// 		err := formatGoFile(filePath)
	// 		if err != nil {
	// 			return err
	// 		}
	// 	}

	// 	// the init.go has methods for generating provider commands needed by the cmd package
	// 	err = initStub(pdir, pkg)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	// the providername.go has the types and the New() method
	// 	err = pkgStub(pdir, pkg)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return nil
	// }

	// // generateStub generates code to statisfy an interface constraint method
	// func generateStub(method reflect.Method, pkg string) string {
	// 	var params []string
	// 	var returns []string

	// 	// Process each parameter
	// 	for i := 0; i < method.Type.NumIn(); i++ {
	// 		paramType := method.Type.In(i)
	// 		paramName := fmt.Sprintf("param%d", i) // Simple parameter names
	// 		params = append(params, paramName+" "+paramType.String())
	// 	}

	// 	// Process return values
	// 	for i := 0; i < method.Type.NumOut(); i++ {
	// 		returnType := method.Type.Out(i)
	// 		returns = append(returns, returnType.String())
	// 	}

	// 	// recieverVar, _ := utf8.DecodeRuneInString(pkg)
	// 	// firstLetter := strings.ToLower(string((recieverVar)))

	// 	// Construct the function signature
	// 	signature := fmt.Sprintf("// %s implements the %s method of the InventoryProvider interface\nfunc %s(%s)",
	// 		method.Name,
	// 		method.Name,
	// 		method.Name,
	// 		strings.Join(params, ", "))

	// 	if len(returns) > 0 {
	// 		signature += " (" + strings.Join(returns, ", ") + ")"
	// 	}

	// 	returnsModified := []string{}
	// 	// stupid, but works for immediate need
	// 	for _, r := range returns {
	// 		switch r {
	// 		case "error":
	// 			returnsModified = append(returnsModified, `nil`)
	// 		case "string":
	// 			returnsModified = append(returnsModified, `""`)
	// 		case "[]string":
	// 			returnsModified = append(returnsModified, `[]string{}`)
	// 		case "*cobra.Command":
	// 			returnsModified = append(returnsModified, `&cobra.Command{}`)
	// 		case "interface":
	// 			returnsModified = append(returnsModified, `map[string]interface{}{}`)
	// 		case "interface{}":
	// 			returnsModified = append(returnsModified, `interface{}{}sdf`)
	// 		case "map[uuid.UUID]provider.HardwareValidationResult":
	// 			returnsModified = append(returnsModified, `map[uuid.UUID]provider.HardwareValidationResult{}`)
	// 		case "[]provider.FieldMetadata":
	// 			returnsModified = append(returnsModified, `[]provider.FieldMetadata{}`)
	// 		case "provider.HardwareRecommendations":
	// 			returnsModified = append(returnsModified, `provider.HardwareRecommendations{}`)
	// 		case "provider.SetFieldsResult":
	// 			returnsModified = append(returnsModified, `provider.SetFieldsResult{}`)
	// 		default:
	// 			returnsModified = append(returnsModified, r)
	// 		}
	// 	}
	// 	// Add a basic function body
	// 	body := fmt.Sprintf(`{
	// log.Printf("%s not yet implemented")

	// return %s
	// }`, method.Name, strings.Join(returnsModified, ", "))

	// 	return signature + " " + body
	// }

	// // isCommandAvailable checks if a command is available in the PATH
	// func isCommandAvailable(name string) bool {
	// 	_, err := exec.LookPath(name)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	return err == nil
	// }

	// // formatGoFile formats a go file using gofmt and goimports
	// func formatGoFile(filePath string) error {
	// 	if isCommandAvailable("gofmt") {
	// 		// Run gofmt
	// 		gofmtCmd := exec.Command("gofmt", "-w", filePath)
	// 		if err := gofmtCmd.Run(); err != nil {
	// 			return fmt.Errorf("gofmt failed: %s", err)
	// 		}
	// 	} else {
	// 		return fmt.Errorf("gofmt not available")
	// 	}

	// 	// Run goimports
	// 	if isCommandAvailable("goimports") {
	// 		goimportsCmd := exec.Command("goimports", "-w", filePath)
	// 		if err := goimportsCmd.Run(); err != nil {
	// 			return fmt.Errorf("goimports failed: %s", err)
	// 		}
	// 	} else {
	// 		return fmt.Errorf("goimports not available")
	// 	}

	// 	return nil
	// }

	// // initStub creates a file with the init command
	// func initStub(dir, pkg string) error {
	// 	newProviderCmd := `
	// 	// NewProviderCmd returns the appropriate command to the cmd layer
	// 	func NewProviderCmd(caniCmd *cobra.Command) (providerCmd *cobra.Command, err error) {
	// 		// first, choose the right command
	// 		switch caniCmd.Name() {
	// 		case "init":
	// 			providerCmd, err = NewSessionInitCommand(caniCmd)
	// 		case "cabinet":
	// 			switch caniCmd.Parent().Name() {
	// 			case "add":
	// 				providerCmd, err = NewAddCabinetCommand(caniCmd)
	// 			case "update":
	// 				providerCmd, err = NewUpdateCabinetCommand(caniCmd)
	// 			case "list":
	// 				providerCmd, err = NewListCabinetCommand(caniCmd)
	// 			}
	// 		case "blade":
	// 			switch caniCmd.Parent().Name() {
	// 			case "add":
	// 				providerCmd, err = NewAddBladeCommand(caniCmd)
	// 			case "update":
	// 				providerCmd, err = NewUpdateBladeCommand(caniCmd)
	// 			case "list":
	// 				providerCmd, err = NewListBladeCommand(caniCmd)
	// 			}
	// 		case "node":
	// 			// check for add/update variants
	// 			switch caniCmd.Parent().Name() {
	// 			case "add":
	// 				providerCmd, err = NewAddNodeCommand(caniCmd)
	// 			case "update":
	// 				providerCmd, err = NewUpdateNodeCommand(caniCmd)
	// 			case "list":
	// 				providerCmd, err = NewListNodeCommand(caniCmd)
	// 			}
	// 		case "export":
	// 			providerCmd, err = NewExportCommand(caniCmd)
	// 		case "import":
	// 			providerCmd, err = NewImportCommand(caniCmd)
	// 		default:
	// 			err = fmt.Errorf("Command not implemented by provider: %s %s", caniCmd.Parent().Name(), caniCmd.Name())
	// 		}
	// 		if err != nil {
	// 			return providerCmd, err
	// 		}

	// 		return providerCmd, nil
	// 	}
	// `

	// 	iStub := filepath.Join(dir, "init.go")
	// 	log.Printf("Generating %+v", iStub)
	// 	f, err := os.Create(iStub)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	defer f.Close()

	// 	// Use reflection to analyze the methods of the ProviderCommands interface
	// 	providerCommands := reflect.TypeOf((*provider.ProviderCommands)(nil)).Elem()
	// 	content := []string{}
	// 	for i := 0; i < providerCommands.NumMethod(); i++ {
	// 		fileContent := ""
	// 		method := providerCommands.Method(i)
	// 		if method.Name != "NewProviderCmd" {
	// 			fileContent = generateStub(method, pkg)
	// 		} else {
	// 			fileContent = newProviderCmd
	// 		}
	// 		if err != nil {
	// 			return err
	// 		}
	// 		content = append(content, fileContent)
	// 	}

	// 	// this is what gets written
	// 	payload := fmt.Sprintf(`package %s

	// %s`,
	// 		strings.ToLower(pkg),
	// 		strings.Join(content, "\n"))

	// 	if _, err := io.WriteString(f, payload); err != nil {
	// 		return err
	// 	}

	// 	err = formatGoFile(iStub)
	// 	if err != nil {
	// 		return err
	// 	}

	return fmt.Errorf("not yet implemented")
}

// // pkgStub creates a file with the provider name for the provider struct
// func pkgStub(dir, pkgname string) error {
// 	//  create a go file with the provider name for the provider struct
// 	pStub := filepath.Join(dir, pkgname+".go")
// 	log.Printf("Generating %+v", pStub)
// 	f, err := os.Create(pStub)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	// this is what gets written
// 	payload := fmt.Sprintf(`package %s

// type %s struct {}

// // options for %s
// type %sOpts struct {}

// // New returns a new %s and is needed to instantiate the provider from the cmd package
// func New() *%s {

// return &%s{}
// }`, pkgname, strings.Title(pkgname), pkgname, strings.Title(pkgname), strings.Title(pkgname), strings.Title(pkgname), strings.Title(pkgname))

// 	if _, err := io.WriteString(f, payload); err != nil {
// 		return err
// 	}

// 	err = formatGoFile(pStub)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func toSnakeCase(str string) string {
// 	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
// 	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
// 	return strings.ToLower(snake)
// }
