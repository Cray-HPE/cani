/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
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

	"github.com/Cray-HPE/cani/pkg/devicetypes/schema"
)

func main() {
	args := os.Args[1:]
	check := false
	if len(args) > 0 && args[0] == "--check" {
		check = true
		args = args[1:]
	}
	if len(args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s [--check] OLD_SCHEMA NEW_SCHEMA\n", os.Args[0])
		os.Exit(2)
	}

	oldData, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading old schema: %v\n", err)
		os.Exit(1)
	}
	newData, err := os.ReadFile(args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading new schema: %v\n", err)
		os.Exit(1)
	}

	report, err := schema.Compare(oldData, newData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "comparing schemas: %v\n", err)
		os.Exit(1)
	}
	fmt.Print(report.Markdown())
	if check {
		if err := schema.ValidateEvolution(oldData, newData); err != nil {
			fmt.Fprintf(os.Stderr, "invalid schema evolution: %v\n", err)
			os.Exit(1)
		}
	}
}
