package import_

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
)

// providerGetter returns the Hpcm singleton to store raw nodes.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearNodes()
	SetNodes(nodes []Node)
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	ClearNodes()
	SetNodes(nodes []Node)
}) {
	providerGetter = getter
}

// Import reads HPCM nodes from --node-json-file and/or --cm-config,
// merges them (deduplicating by name), and stores on the provider singleton.
// No transformation is done here; that happens in the transform step.
func Import(cmd *cobra.Command, args []string) error {
	p := providerGetter()

	jsonFile, _ := cmd.Flags().GetString("node-json-file")
	cmConfigFile, _ := cmd.Flags().GetString("cm-config")

	var allNodes []Node

	// Parse JSON nodes if --node-json-file is provided.
	if jsonFile != "" {
		nodes, err := importNodeJson(jsonFile)
		if err != nil {
			return err
		}
		allNodes = append(allNodes, nodes...)
	}

	// Parse cm.config nodes if --cm-config is provided.
	if cmConfigFile != "" {
		nodes, err := importCmConfig(cmConfigFile)
		if err != nil {
			return err
		}
		allNodes = append(allNodes, nodes...)
	}

	// Fall back to stdin if neither flag is provided.
	if jsonFile == "" && cmConfigFile == "" {
		nodes, err := importStdin()
		if err != nil {
			return err
		}
		allNodes = append(allNodes, nodes...)
	}

	allNodes = deduplicateNodes(allNodes)

	p.ClearNodes()
	p.SetNodes(allNodes)
	log.Printf("Stored %d raw nodes for transform", len(allNodes))
	return nil
}

// importNodeJson reads and parses nodes from a JSON file.
func importNodeJson(path string) ([]Node, error) {
	log.Printf("Reading nodes from %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading file %s: %w", path, err)
	}
	return ParseNodes(data)
}

// importCmConfig reads and parses nodes from an HPCM cm.config file.
func importCmConfig(path string) ([]Node, error) {
	log.Printf("Reading cm.config from %s", path)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cm.config %s: %w", path, err)
	}
	cfg, err := ParseCmConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parsing cm.config: %w", err)
	}
	return CmConfigToNodes(cfg), nil
}

// importStdin reads and parses nodes from stdin.
func importStdin() ([]Node, error) {
	log.Printf("Waiting for nodes from stdin...")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("reading stdin: %w", err)
	}
	return ParseNodes(data)
}

// deduplicateNodes merges nodes by Name, preferring the first occurrence.
// Logs a warning for each duplicate.
func deduplicateNodes(nodes []Node) []Node {
	seen := make(map[string]bool, len(nodes))
	out := make([]Node, 0, len(nodes))
	for _, n := range nodes {
		if seen[n.Name] {
			log.Printf("WARNING: duplicate node %q — keeping first occurrence", n.Name)
			continue
		}
		seen[n.Name] = true
		out = append(out, n)
	}
	return out
}

// Node matches the HPCM node JSON shape from the upstream hpcm-client.
// Contains only raw JSON fields; no inferred or computed data.
type Node struct {
	Name                 string                 `json:"name,omitempty"`
	Aliases              map[string]string      `json:"aliases,omitempty"`
	ID                   int64                  `json:"id,omitempty"`
	UUID                 string                 `json:"uuid,omitempty"`
	Etag                 string                 `json:"etag,omitempty"`
	CreationTime         time.Time              `json:"creationTime,omitempty"`
	ModificationTime     time.Time              `json:"modificationTime,omitempty"`
	DeletionTime         time.Time              `json:"deletionTime,omitempty"`
	Links                map[string]string      `json:"links,omitempty"`
	Network              *NetworkSettings       `json:"network,omitempty"`
	Image                *ImageSettings         `json:"image,omitempty"`
	Platform             *PlatformSettings      `json:"platform,omitempty"`
	Management           *ManagementSettings    `json:"management,omitempty"`
	Controller           *ControllerSettings    `json:"controller,omitempty"`
	Location             *LocationSettings      `json:"location,omitempty"`
	InternalName         string                 `json:"internalName,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	ImageTransport       string                 `json:"imageTransport,omitempty"`
	ImagePending         bool                   `json:"imagePending,omitempty"`
	TemplateName         string                 `json:"templateName,omitempty"`
	RootFs               string                 `json:"rootFs,omitempty"`
	OperationalStatus    int32                  `json:"operationalStatus,omitempty"`
	AdministrativeStatus int32                  `json:"administrativeStatus,omitempty"`
	Managed              bool                   `json:"managed,omitempty"`
	Monitoring           string                 `json:"monitoring,omitempty"`
	RootSlot             int32                  `json:"rootSlot,omitempty"`
	BiosBootMode         string                 `json:"biosBootMode,omitempty"`
	BootOrder            int32                  `json:"bootOrder,omitempty"`
	IscsiRoot            string                 `json:"iscsiRoot,omitempty"`
	Inventory            map[string]string      `json:"inventory,omitempty"`
	NodeController       string                 `json:"nodeController,omitempty"`
	Attributes           map[string]interface{} `json:"attributes,omitempty"`
}

// ParseNodes unmarshals JSON bytes into []Node.
func ParseNodes(data []byte) ([]Node, error) {
	var nodes []Node
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return nodes, nil
}
