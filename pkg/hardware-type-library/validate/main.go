// MIT License
//
// (C) Copyright [2023] Hewlett Packard Enterprise Development LP
//
// Permission is hereby granted, free of charge, to any person obtaining a
// copy of this software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation
// the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the
// Software is furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included
// in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
// THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
// OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
// ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
// OTHER DEALINGS IN THE SOFTWARE.
package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/santhosh-tekuri/jsonschema"
	"gopkg.in/yaml.v3"
)

func unmarshalMultipleDocuments(in []byte) ([]interface{}, error) {
	r := bytes.NewReader(in)
	decoder := yaml.NewDecoder(r)

	var documents []interface{}
	for {
		var document interface{}
		if err := decoder.Decode(&document); err != nil {
			// Break out of loop when more yaml documents to process
			if err != io.EOF {
				return nil, err
			}

			break
		}

		documents = append(documents, document)
	}

	return documents, nil
}

func main() {
	compiler := jsonschema.NewCompiler()

	// Parse schemas
	schemaBaseDir := os.Args[1]
	schemaFiles, err := os.ReadDir(schemaBaseDir)
	if err != nil {
		panic(err)
	}
	for _, schemaFile := range schemaFiles {
		if schemaFile.IsDir() || !strings.HasSuffix(schemaFile.Name(), ".json") {
			continue
		}
		filePath := path.Join(schemaBaseDir, schemaFile.Name())

		fmt.Println("Reading schema:", filePath)

		fileRaw, err := ioutil.ReadFile(filePath)
		if err != nil {
			panic(err)
		}

		err = compiler.AddResource(path.Base(filePath), bytes.NewReader(fileRaw))
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("Compiling devicetype.json")
	schema := compiler.MustCompile("devicetype.json")
	_ = schema

	// Validate yaml files
	yamlBaseDir := os.Args[2]
	yamlFiles, err := os.ReadDir(yamlBaseDir)
	if err != nil {
		panic(err)
	}
	failed := false
	for _, yamlFile := range yamlFiles {
		if yamlFile.IsDir() || !(strings.HasSuffix(yamlFile.Name(), ".yaml") || strings.HasSuffix(yamlFile.Name(), ".yml")) {
			continue
		}
		yamlPath := path.Join(yamlBaseDir, yamlFile.Name())
		fmt.Println()
		fmt.Println("Validating:", yamlPath)

		yamlRaw, err := ioutil.ReadFile(yamlPath)
		if err != nil {
			fmt.Println("Failed to read file", yamlPath)
			fmt.Println("Error:", err)
			failed = true
			continue
		}

		documents, err := unmarshalMultipleDocuments(yamlRaw)
		if err != nil {
			fmt.Println("Failed to unmarshal file", yamlPath)
			fmt.Println("Error:", err)
			failed = true
			continue
		}
		for i, document := range documents {
			fmt.Println("Validating document", i, "of", len(documents), "in", yamlPath)

			err := schema.ValidateInterface(document)
			if err != nil {
				fmt.Println(err)
				failed = true
			}
		}
	}

	fmt.Println()
	fmt.Println("Results")

	if failed {
		fmt.Println("Validation Failed")
		os.Exit(1)
	} else {
		fmt.Println("Validation Passed")
	}

}
