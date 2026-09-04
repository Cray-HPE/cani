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
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type moduleFile struct {
	Require []struct {
		Path     string
		Indirect bool
	}
}

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "check-go-deps:", err)
		os.Exit(1)
	}
}

func run(args []string, output io.Writer) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: checkdeps <go.mod> <allowlist>")
	}
	allowed, err := readAllowlist(args[1])
	if err != nil {
		return err
	}
	direct, err := readDirectRequirements(args[0])
	if err != nil {
		return err
	}
	var unsanctioned []string
	for _, dependency := range direct {
		if !allowed[dependency] {
			unsanctioned = append(unsanctioned, dependency)
		}
	}
	if len(unsanctioned) != 0 {
		return fmt.Errorf("unsanctioned direct dependency in go.mod:\n  + %s\n\nadd reviewed dependencies to %s; otherwise use the standard library",
			strings.Join(unsanctioned, "\n  + "), args[1])
	}
	_, err = fmt.Fprintf(output, "go.mod direct dependencies OK (%d sanctioned module(s) in use)\n", len(direct))
	return err
}

func readDirectRequirements(path string) ([]string, error) {
	command := exec.Command("go", "mod", "edit", "-json", path)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	data, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var manifest moduleFile
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("decoding go mod edit output: %w", err)
	}
	dependencies := make(map[string]bool)
	for _, requirement := range manifest.Require {
		if !requirement.Indirect {
			dependencies[requirement.Path] = true
		}
	}
	direct := make([]string, 0, len(dependencies))
	for dependency := range dependencies {
		direct = append(direct, dependency)
	}
	sort.Strings(direct)
	return direct, nil
}

func readAllowlist(path string) (map[string]bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading allowlist: %w", err)
	}
	defer file.Close()
	allowed := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line, _, _ := strings.Cut(scanner.Text(), "#")
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		if len(fields) != 1 {
			return nil, fmt.Errorf("invalid allowlist entry: %s", scanner.Text())
		}
		allowed[fields[0]] = true
	}
	return allowed, scanner.Err()
}
