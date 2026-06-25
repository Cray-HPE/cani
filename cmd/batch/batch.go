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
package batch

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/util/store"
	"github.com/Cray-HPE/cani/pkg/datastores"
)

// NewCommand creates the "batch" command. dispatch runs a single parsed command
// line; the command layer injects it (building a fresh command tree per line)
// so this package needs no dependency on the root command package.
func NewCommand(dispatch func(args []string) error) *cli.Command {
	cmd := &cli.Command{
		Use:   "batch <file>",
		Short: "Run many cani commands from a file in a single process.",
		Long: `Run a file of cani commands in one process for speed.

Each line is one cani invocation, exactly as you would type it; an optional
leading "cani" or "bin/cani" is ignored. Blank lines, "#" comments, and any
non-cani lines (such as rm or make helpers) are skipped, so an existing setup
script can often be run as-is.

The inventory is loaded once at the start and written once at the end instead of
once per command, which makes large scripts dramatically faster.

Examples:
  cani alpha batch maple_inventory.sh
  cani alpha batch topology.cani --continue-on-error`,
		Args: func(cmd *cli.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("accepts 1 arg(s), received %d", len(args))
			}
			return nil
		},
		RunE: func(cmd *cli.Command, args []string) error {
			return runBatch(cmd, args, dispatch)
		},
	}

	cmd.Flags().Bool("continue-on-error", false, "keep running after a command fails instead of stopping")
	cmd.Flags().Bool("dry-run", false, "run every command but do not save the inventory")

	return cmd
}

// runBatch loads the inventory once, runs each command line against an in-memory
// session, then persists the result a single time.
func runBatch(cmd *cli.Command, args []string, dispatch func(args []string) error) error {
	lines, err := readLines(args[0])
	if err != nil {
		return err
	}

	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	inventory, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	session := datastores.BeginSession(inventory)
	defer datastores.EndSession()

	continueOnError, _ := cmd.Flags().GetBool("continue-on-error")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	executed, failed, stopErr := runLines(dispatch, lines, continueOnError)

	if dryRun {
		log.Printf("batch dry-run: %d command(s) executed, %d failed; inventory not saved", executed, failed)
		return stopErr
	}

	if err := flush(cmd, session); err != nil {
		return err
	}
	log.Printf("batch complete: %d command(s) executed, %d failed", executed, failed)
	return stopErr
}

// runLines dispatches each command line in order. It returns how many commands
// ran and failed plus the error that stopped the run (nil unless a command
// failed while continueOnError is false). Commands that succeeded before a stop
// remain applied to the session inventory, mirroring how the equivalent
// sequential script would have persisted its progress.
func runLines(dispatch func(args []string) error, lines []string, continueOnError bool) (executed, failed int, stopErr error) {
	for i, raw := range lines {
		tokens, ok := commandTokens(raw)
		if !ok {
			continue
		}
		if err := dispatch(tokens); err != nil {
			failed++
			log.Printf("batch: line %d failed: %v", i+1, err)
			if !continueOnError {
				return executed, failed, fmt.Errorf("batch stopped at line %d: %w", i+1, err)
			}
			continue
		}
		executed++
	}
	return executed, failed, nil
}

// flush ends the in-memory session and persists the accumulated inventory once
// through the configured disk store.
func flush(cmd *cli.Command, session *datastores.MemStore) error {
	datastores.EndSession()
	if err := store.Setup(cmd); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}
	if err := datastores.Datastore.Save(session.Inventory()); err != nil {
		return fmt.Errorf("failed to save inventory: %w", err)
	}
	return nil
}
