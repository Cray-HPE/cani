package nautobot

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
	import_ "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
	"github.com/spf13/cobra"
)

// Import implements the provider.Importer interface.
// It delegates to the import subpackage to fetch all entity types from
// the Nautobot API and stores the raw responses for later use by Transform().
func (p *Nautobot) Import(cmd *cobra.Command, args []string, inventory *devicetypes.Inventory) error {
	if err := import_.Import(cmd, args, inventory); err != nil {
		return fmt.Errorf("nautobot import failed: %w", err)
	}

	clog.Detail("  Fetched %d locations", len(p.rawLocations))
	clog.Detail("  Fetched %d racks", len(p.rawRacks))
	clog.Detail("  Fetched %d devices", len(p.rawDevices))
	clog.Detail("  Fetched %d device types", len(p.rawDeviceTypes))
	clog.Detail("  Fetched %d interfaces", len(p.rawInterfaces))
	clog.Detail("  Fetched %d modules", len(p.rawModules))
	clog.Detail("  Fetched %d module bays", len(p.rawModuleBays))
	clog.Detail("  Fetched %d cables", len(p.rawCables))
	clog.Detail("  Fetched %d inventory items", len(p.rawInventoryItems))

	return nil
}

// importNautobot is the RunE handler for "cani import nautobot".
// It creates the API client, tests the connection, and prompts for
// defaults if they were not supplied via CLI flags.
func (p *Nautobot) importNautobot(cmd *cobra.Command, args []string) error {
	clog.Info("Importing via Nautobot provider...")

	if err := p.loadOptionsFromConfig(); err != nil {
		return fmt.Errorf("failed to load nautobot config: %w", err)
	}
	p.applyFlagOverrides(cmd)

	if p.Options.URL == "" {
		return fmt.Errorf("nautobot URL not configured; set nautobot.url in the config file")
	}
	if p.Options.Token == "" {
		return fmt.Errorf("nautobot token not configured; set nautobot.token in the config file")
	}

	client, err := export.NewNautobotClient(p.Options.URL, p.Options.Token)
	if err != nil {
		return fmt.Errorf("failed to create Nautobot client: %w", err)
	}

	ctx := context.Background()
	if err := client.TestConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Nautobot: %w", err)
	}
	clog.Info("Successfully connected to Nautobot")

	cache := export.NewLookupCache(client)
	cache.SetContext(ctx)

	// Store on provider for use by Import().
	p.ctx = ctx
	p.client = client
	p.cache = cache

	return nil
}

// promptForSelection prompts the user to select from a list of available items.
func promptForSelection(itemType string, listFunc func() ([]*export.CachedItem, error)) (string, error) {
	items, err := listFunc()
	if err != nil {
		return "", err
	}
	if len(items) == 0 {
		return "", fmt.Errorf("no %ss found in Nautobot", itemType)
	}

	fmt.Printf("\nAvailable %ss:\n", itemType)
	for i, item := range items {
		fmt.Printf("  [%d] %s\n", i+1, item.Display)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("Select default %s (1-%d): ", itemType, len(items))
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		input = strings.TrimSpace(input)
		idx, err := strconv.Atoi(input)
		if err != nil || idx < 1 || idx > len(items) {
			fmt.Printf("Invalid selection. Please enter a number between 1 and %d.\n", len(items))
			continue
		}

		return items[idx-1].Name, nil
	}
}
