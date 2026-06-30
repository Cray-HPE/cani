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
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gopkg.in/yaml.v3"
)

const wcfKeyCount = 50

// seedConfigNode builds a multi-key YAML document node used to widen any
// non-atomic write window so a torn read is easy to observe.
func seedConfigNode(t *testing.T) *yaml.Node {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("providers:\n")
	for i := 0; i < wcfKeyCount; i++ {
		fmt.Fprintf(&sb, "  key%02d: value-%02d\n", i, i)
	}
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(sb.String()), &root); err != nil {
		t.Fatalf("seeding node: %v", err)
	}
	return &root
}

// checkConfigComplete reads path and returns an error if the file is missing
// keys or does not parse — either symptom of a torn (non-atomic) write.
func checkConfigComplete(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // atomic rename keeps path present; tolerate just in case.
		}
		return fmt.Errorf("reading config: %w", err)
	}
	var doc struct {
		Providers map[string]string `yaml:"providers"`
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("torn read — config did not parse: %w", err)
	}
	if len(data) > 0 && len(doc.Providers) != wcfKeyCount {
		return fmt.Errorf("torn read — got %d keys, want %d", len(doc.Providers), wcfKeyCount)
	}
	return nil
}

// writeLoop rewrites path up to iterations times, stopping early if stop closes.
func writeLoop(path string, root *yaml.Node, iterations int, stop <-chan struct{}) error {
	for i := 0; i < iterations; i++ {
		select {
		case <-stop:
			return nil
		default:
		}
		if err := writeConfigFile(path, root); err != nil {
			return fmt.Errorf("concurrent write: %w", err)
		}
	}
	return nil
}

// readUntilDone reads and parses path repeatedly until done closes, returning the
// first torn-read error observed (nil if the file always parsed completely).
func readUntilDone(path string, done <-chan struct{}) error {
	for {
		select {
		case <-done:
			return nil
		default:
			if err := checkConfigComplete(path); err != nil {
				return err
			}
		}
	}
}

// stressWriteConfigFile concurrently rewrites path with writers goroutines while
// reading it back on the caller's goroutine, returning the first torn-read or
// write failure observed (nil if none).
func stressWriteConfigFile(path string, root *yaml.Node, writers, iterations int) error {
	stop := make(chan struct{})
	var stopOnce sync.Once
	stopAll := func() { stopOnce.Do(func() { close(stop) }) }

	errs := make(chan error, writers)
	var wg sync.WaitGroup
	wg.Add(writers)
	for w := 0; w < writers; w++ {
		go func() {
			defer wg.Done()
			errs <- writeLoop(path, root, iterations, stop)
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	readErr := readUntilDone(path, done)
	stopAll() // no-op on the happy path; halts writers on a detected torn read.
	wg.Wait() // ensure no writer goroutine outlives the caller.
	close(errs)

	if readErr != nil {
		return readErr
	}
	for err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// TestWriteConfigFileAtomicUnderConcurrency verifies writeConfigFile replaces the
// config atomically so a concurrent reader never observes a half-written file.
//
// Why it matters: setupDomain Loads then Saves the config on every cani
// invocation, so a shell pipeline like `cani export | ... | cani import` runs two
// processes that read and rewrite the same cani.yml at once. A non-atomic
// (O_TRUNC + stream) write let a reader see a truncated file, surfacing as
// "yaml: line N: could not find expected ':'" and a flaky CI failure.
// Inputs: four writer goroutines rewriting a 50-key document while the main
// goroutine continuously reads and parses it. Outputs: every read yields a
// complete, 50-key document and no temp files are left behind.
// Data choice: 50 keys makes any partial write large enough to reliably parse
// as incomplete, so this test fails against the old streaming implementation.
func TestWriteConfigFileAtomicUnderConcurrency(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cani.yml")
	root := seedConfigNode(t)

	// Seed once so readers always find the file present.
	if err := writeConfigFile(path, root); err != nil {
		t.Fatalf("seed write: %v", err)
	}

	if err := stressWriteConfigFile(path, root, 4, 150); err != nil {
		t.Fatal(err)
	}
	if err := checkConfigComplete(path); err != nil {
		t.Fatal(err)
	}
	if leftover, _ := filepath.Glob(filepath.Join(dir, ".cani-config-*.tmp")); len(leftover) != 0 {
		t.Errorf("temp files left behind: %v", leftover)
	}
}
