package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Cray-HPE/cani/pkg/devicetypes/schema"
)

// genschema writes the generated JSON Schema for the cani inventory datastore
// to disk. By default it writes to pkg/devicetypes/schema/inventory.schema.json
// (relative to the repository root); an alternate path may be passed as the
// first argument. It mirrors the tools/gendocs generator pattern.
func main() {
	out := "pkg/devicetypes/schema/inventory.schema.json"
	if len(os.Args) > 1 {
		out = os.Args[1]
	}

	data, err := schema.Generate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating schema: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "error creating directory %s: %v\n", filepath.Dir(out), err)
		os.Exit(1)
	}

	if err := os.WriteFile(out, data, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing schema to %s: %v\n", out, err)
		os.Exit(1)
	}

	fmt.Printf("schema written to %s\n", out)
}
