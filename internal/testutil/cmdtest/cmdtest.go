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

// Package cmdtest runs CRUD commands in-process against an in-memory
// datastore. It exists so cmd/add, cmd/remove, cmd/update and cmd/show can be
// exercised through their real RunE hooks without touching disk, reading
// os.Args, or registering a provider.
package cmdtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/provider"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

// FakeStore is an in-memory DeviceStore that records how the command layer
// used it, so tests can assert on save counts (e.g. dry-run must not save).
type FakeStore struct {
	Inventory *devicetypes.Inventory
	Loads     int
	Saves     int
	LoadErr   error
	SaveErr   error
}

// Load returns a fresh copy of the inventory and counts the call.
//
// The copy is produced by a JSON round-trip so the fake matches JSONStore.Load:
// commands mutate a detached object, derived reverse indices are rebuilt from
// the forward FKs rather than trusted, and anything the schema cannot serialize
// is lost here exactly as it would be on disk.
func (s *FakeStore) Load() (*devicetypes.Inventory, error) {
	s.Loads++
	if s.LoadErr != nil {
		return nil, s.LoadErr
	}

	encoded, err := json.Marshal(s.Inventory)
	if err != nil {
		return nil, fmt.Errorf("cmdtest: marshalling inventory: %w", err)
	}
	fresh := devicetypes.NewInventory()
	if err := json.Unmarshal(encoded, fresh); err != nil {
		return nil, fmt.Errorf("cmdtest: unmarshalling inventory: %w", err)
	}

	fresh.VerifyParentChildRelationships()
	return fresh, nil
}

// Save records the inventory the command produced and counts the call.
func (s *FakeStore) Save(inventory *devicetypes.Inventory) error {
	s.Saves++
	if s.SaveErr != nil {
		return s.SaveErr
	}
	s.Inventory = inventory
	return nil
}

// Harness holds a synthetic command tree and the store backing it.
type Harness struct {
	Root   *cli.Command
	Parent *cli.Command
	Target *cli.Command
	Store  *FakeStore
	Out    *bytes.Buffer
	Err    *bytes.Buffer
}

// New builds root -> parent -> target and installs an in-memory store.
//
// The synthetic root supplies the persistent "datastore" flag that
// internal/util/store.Setup looks up via cmd.Root(). Callers pass the real
// parent (e.g. add.NewCommand()) so the target inherits the real persistent
// flags, and a nil inventory yields an empty one.
//
// Tests using this harness must not call t.Parallel: datastore selection is
// process-global.
func New(t *testing.T, parent, target *cli.Command, inventory *devicetypes.Inventory) *Harness {
	t.Helper()

	if inventory == nil {
		inventory = devicetypes.NewInventory()
	}

	root := &cli.Command{Use: "cani"}
	root.PersistentFlags().String("datastore", "json", "datastore type")
	root.AddCommand(parent)
	parent.AddCommand(target)

	// (*cli.Command).mergeInheritedFlags is unexported and only runs inside
	// Execute, which reads os.Args; replicate it for the tree just built.
	target.Flags().AddFlagSet(parent.PersistentFlags())
	target.Flags().AddFlagSet(root.PersistentFlags())

	harness := &Harness{
		Root:   root,
		Parent: parent,
		Target: target,
		Store:  &FakeStore{Inventory: inventory},
		Out:    &bytes.Buffer{},
		Err:    &bytes.Buffer{},
	}
	root.SetOut(harness.Out)
	root.SetErr(harness.Err)

	datastores.SetDeviceStoreForTest(t, harness.Store)

	return harness
}

// Run sets the given flags then executes the target's Args, PreRunE and RunE
// hooks in the same order as (*cli.Command).runPipeline.
//
// Each flag maps to a slice so repeatable flags (--tag, --metadata, --set) can
// be set more than once, which is how providers receive generic CLI input.
//
// Two pipeline steps are absent. PersistentPreRunE is skipped deliberately: the
// real root uses it to load configuration and provider state, which is exactly
// what these tests isolate. validateFlagGroups cannot be called at all because
// it is unexported, so flag-group constraints are not enforced here.
func (h *Harness) Run(t *testing.T, flags map[string][]string, args ...string) error {
	t.Helper()

	for name, values := range flags {
		for _, value := range values {
			if err := h.Target.Flags().Set(name, value); err != nil {
				t.Fatalf("set --%s=%q: %v", name, value, err)
			}
		}
	}

	if h.Target.Args != nil {
		if err := h.Target.Args(h.Target, args); err != nil {
			return err
		}
	}
	if h.Target.PreRunE != nil {
		if err := h.Target.PreRunE(h.Target, args); err != nil {
			return err
		}
	}
	return h.Target.RunE(h.Target, args)
}

// Inventory returns the inventory currently held by the store.
func (h *Harness) Inventory() *devicetypes.Inventory {
	return h.Store.Inventory
}

// RequireNoProviders fails the test when any provider has self-registered.
//
// This proves a CRUD package has not grown a concrete provider import, since
// importing one runs its init() and registers it. It is necessary but not
// sufficient: branching on a provider *name* needs no import, and behaviour
// that varies with registry contents would still pass here. Those are covered
// by TestCRUDHasNoProviderNameLiterals and
// TestCRUDResultIsInvariantAcrossRegisteredProviders respectively.
func RequireNoProviders(t *testing.T) {
	t.Helper()
	if registered := provider.GetProviders(); len(registered) != 0 {
		names := make([]string, 0, len(registered))
		for name := range registered {
			names = append(names, name)
		}
		sort.Strings(names)
		t.Fatalf("CRUD must not depend on providers, but these are registered: %v", names)
	}
}
