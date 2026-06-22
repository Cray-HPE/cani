package import_

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/commands"
	"github.com/Cray-HPE/cani/pkg/visual"
)

// providerGetter returns the Redfish singleton to store raw roots.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearRoots()
	SetRoots(roots []ServiceRoot)
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	ClearRoots()
	SetRoots(roots []ServiceRoot)
}) {
	providerGetter = getter
}

// Import reads Redfish ServiceRoot JSON from --root file or stdin,
// parses it (single object or array), deduplicates by UUID, and stores
// on the provider singleton. No transformation is done here.
func Import(cmd *cli.Command, args []string, inventory *devicetypes.Inventory) error {
	data, err := readInput(commands.RootFlag)
	if err != nil {
		return err
	}

	roots, err := ParseServiceRoots(data)
	if err != nil {
		return err
	}

	if len(roots) == 0 {
		log.Println("No valid ServiceRoot records found")
		return nil
	}

	// Show step-through output if step mode is enabled (before dedup).
	if config.Cfg != nil && config.Cfg.StepMode {
		opts := visual.ETLOptions{NoColor: config.Cfg.NoColor}
		for i, root := range roots {
			raw := formatRawRoot(root)
			parsed := formatParsedRoot(root)
			id := formatIdentifier(root)
			if err := visual.PromptRecordStepRaw(i+1, len(roots), raw, parsed, id, opts); err != nil {
				return fmt.Errorf("step interrupted: %w", err)
			}
		}
	}

	totalParsed := len(roots)
	roots = deduplicateRoots(roots)

	// Store on the singleton for the transform phase.
	prov := providerGetter()
	prov.ClearRoots()
	prov.SetRoots(roots)

	if dupes := totalParsed - len(roots); dupes > 0 {
		log.Printf("Parsed %d Redfish ServiceRoot(s) from %s (%d duplicate(s) removed)",
			totalParsed, sourceLabel(commands.RootFlag), dupes)
	} else {
		log.Printf("Parsed %d Redfish ServiceRoot(s) from %s",
			totalParsed, sourceLabel(commands.RootFlag))
	}
	return nil
}

// readInput reads from the file at path, or from stdin if path is empty.
func readInput(path string) ([]byte, error) {
	if path != "" {
		log.Printf("Reading Redfish ServiceRoot from %s", path)
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading file %s: %w", path, err)
		}
		return data, nil
	}

	log.Println("Reading Redfish ServiceRoot from stdin...")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return data, nil
}

// deduplicateRoots keeps the first ServiceRoot for each unique endpoint.
// The key combines UUID with the BMC FQDN (or hostname) so that different
// physical servers sharing a UUID are preserved as separate records.
func deduplicateRoots(roots []ServiceRoot) []ServiceRoot {
	seen := make(map[string]bool, len(roots))
	out := make([]ServiceRoot, 0, len(roots))
	for _, r := range roots {
		key := deduplicationKey(r)
		if seen[key] {
			log.Printf("WARNING: duplicate ServiceRoot %q — keeping first occurrence", key)
			continue
		}
		seen[key] = true
		out = append(out, r)
	}
	return out
}

// deduplicationKey builds a composite key that uniquely identifies a physical
// endpoint. It combines UUID with BMC FQDN/hostname so two servers that
// happen to share a UUID but have distinct BMCs are not collapsed.
func deduplicationKey(r ServiceRoot) string {
	id := r.UUID
	if id == "" {
		id = r.Product
	}
	if fqdn := r.ManagerFQDN(); fqdn != "" {
		return id + "|" + fqdn
	}
	if host := r.ManagerHostName(); host != "" {
		return id + "|" + host
	}
	return id
}

// formatRawRoot creates a compact string of the raw ServiceRoot data.
func formatRawRoot(r ServiceRoot) string {
	return fmt.Sprintf("Product=%s Vendor=%s UUID=%s RedfishVersion=%s",
		r.Product, r.Vendor, r.UUID, r.RedfishVersion)
}

// formatParsedRoot creates a descriptive string of the parsed record.
func formatParsedRoot(r ServiceRoot) string {
	mgr := r.ManagerType()
	if mgr != "" {
		return fmt.Sprintf("server: %s (%s %s)", r.Product, mgr, r.ManagerFirmwareVersion())
	}
	return fmt.Sprintf("server: %s", r.Product)
}

// formatIdentifier returns a likely-unique identifier for the record.
// Prefers FQDN, falls back to hostname, then UUID.
func formatIdentifier(r ServiceRoot) string {
	if fqdn := r.ManagerFQDN(); fqdn != "" {
		return fqdn
	}
	if host := r.ManagerHostName(); host != "" {
		return host
	}
	return r.UUID
}

// sourceLabel returns a display label for the input source.
func sourceLabel(path string) string {
	if path != "" {
		return path
	}
	return "stdin"
}
