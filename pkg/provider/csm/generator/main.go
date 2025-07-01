/*
 *
 *  MIT License
 *
 *  (C) Copyright 2021-2023 Hewlett Packard Enterprise Development LP
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

package main

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"text/template"

	"github.com/Cray-HPE/cani/internal/provider/csm"
	"github.com/Cray-HPE/hms-xname/xnametypes"
)

type XnameField struct {
	Name          string
	LocationIndex int
}

type XnameTypeNode struct {
	Parent   *XnameTypeNode
	Children []*XnameTypeNode

	Entry  xnametypes.HMSCompRecognitionEntry
	Fields []XnameField
}

func main() {

	//
	// Build up XnameTypeNode
	//
	nodes := map[xnametypes.HMSType]*XnameTypeNode{}
	for _, entry := range xnametypes.GetHMSCompRecognitionTable() {
		if _, exists := nodes[entry.Type]; exists {
			panic(fmt.Errorf("Error: entry type already exists: %v", entry))
		}

		nodes[entry.Type] = &XnameTypeNode{
			Entry: entry,
		}
	}

	//
	// Create the child and parent relationships between XnameTypeNode
	//
	for _, node := range nodes {
		if node.Entry.ParentType == xnametypes.HMSTypeInvalid {
			continue
		}

		parentNode, parentExists := nodes[node.Entry.ParentType]
		if !parentExists {
			panic(fmt.Errorf("Error: parent type (%v) does not exist for type (%v) ", node.Entry.ParentType, node.Entry.Type))
		}

		// Update parent and child links
		node.Parent = parentNode
		parentNode.Children = append(parentNode.Children, node)
	}

	//
	// Sort the elements of the child slice by their HMSType
	// When templating this will generate the dot notation functions to get children
	// in a deterministic order.
	//
	for _, node := range nodes {
		sort.Slice(node.Children, func(i, j int) bool {
			return node.Children[i].Entry.Type < node.Children[j].Entry.Type
		})
	}

	//
	// Determine field names
	//
	for hmsType, node := range nodes {
		if typeConverter, exists := csm.GetXnameTypeConverters()[hmsType]; exists {
			// Only build the field data if we have a converter for it.

			// Get the field names
			fieldNames := getFields(node)

			// Get the hardware path location ordinal to xname ordinal mapping
			ordinalIndexMapping := typeConverter.GetOrdinalIndexMapping()

			// Get the number of expected ordinals in the xname
			_, expectedCount, err := xnametypes.GetHMSTypeFormatString(hmsType)
			if err != nil {
				panic(err)
			}

			// The number of field names needs to match the number of mappings
			if expectedCount != len(fieldNames) {
				fmt.Printf("Unexpected number of field names %d expected %d for %s\n", len(fieldNames), expectedCount, hmsType)
				os.Exit(1)
			}

			// The number of field names needs to match the number of mappings
			if expectedCount != len(ordinalIndexMapping) {
				fmt.Printf("Unexpected number of ordinal mappings %d expected %d for %s\n", len(ordinalIndexMapping), expectedCount, hmsType)
				os.Exit(1)
			}

			// Create the field list
			for i := 0; i < expectedCount; i++ {
				node.Fields = append(node.Fields, XnameField{
					Name:          fieldNames[i],
					LocationIndex: ordinalIndexMapping[i],
				})
			}
		}
	}

	//
	// Generate a list of HMSTypes based on the xname hierarchy.
	// IE the System HMSType is generated first.
	//
	root := nodes[xnametypes.System]
	xnameTypes := getTypeNames(root)

	xnameTypeNodes := []*XnameTypeNode{}
	for _, xnameType := range xnameTypes {
		// Lets filter out any types that haven't an existing type converter
		if _, exists := csm.GetXnameTypeConverters()[xnameType]; !exists {
			continue
		}

		xnameTypeNodes = append(xnameTypeNodes, nodes[xnameType])
	}

	//
	// Template
	//
	templateFile("./generator/types_generated.go.tpl", "./types_generated.go", xnameTypeNodes)

}

func getTypeNames(node *XnameTypeNode) []xnametypes.HMSType {
	types := []xnametypes.HMSType{node.Entry.Type}

	for _, child := range node.Children {
		types = append(types, getTypeNames(child)...)
	}

	return types
}

func getFields(node *XnameTypeNode) []string {
	if node == nil {
		return nil
	}

	if node.Entry.Type == xnametypes.System {
		return nil
	}

	return append(getFields(node.Parent), string(node.Entry.Type))
}

func templateFile(sourceFilePath, destFilePath string, xnameTypes []*XnameTypeNode) {
	fmt.Println("Templating", sourceFilePath)
	t, err := template.ParseFiles(sourceFilePath)
	if err != nil {
		panic(err)
	}
	f, err := os.Create(destFilePath)
	if err != nil {
		panic(err)
	}
	if err := t.Execute(f, xnameTypes); err != nil {
		panic(err)
	}

	fmt.Println("Formatting", destFilePath)
	cmd := exec.Command("go", "fmt", destFilePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout
	if err := cmd.Run(); err != nil {
		panic(err)
	}
}
