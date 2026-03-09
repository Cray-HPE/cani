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
	imprt "github.com/Cray-HPE/cani/pkg/provider/nautobot/import"
	"github.com/spf13/cobra"
)

// Import delegates to the import sub-package.
func (p *Nautobot) Import(cmd *cobra.Command, args []string) (importedDevices []devicetypes.CaniDeviceType, err error) {
	return imprt.Import(cmd, args)
}

// importNautobot is the RunE handler for "cani import nautobot".
func (p *Nautobot) importNautobot(cmd *cobra.Command, args []string) error {
	clog.Info("Importing via Nautobot provider...")

	url, _ := cmd.Flags().GetString("url")
	token, _ := cmd.Flags().GetString("token")
	defaultLocation, _ := cmd.Flags().GetString("default-location")
	defaultRole, _ := cmd.Flags().GetString("default-role")
	defaultStatus, _ := cmd.Flags().GetString("default-status")
	merge, _ := cmd.Flags().GetBool("merge")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if url == "" {
		return fmt.Errorf("--url is required")
	}
	if token == "" {
		return fmt.Errorf("--token is required")
	}

	p.Options.URL = url
	p.Options.Token = token
	if p.Options.Export == nil {
		p.Options.Export = &NautobotExportOpts{}
	}
	p.Options.Export.Merge = merge
	p.Options.Export.DryRun = dryRun

	client, err := export.NewNautobotClient(url, token)
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

	if defaultLocation == "" {
		location, err := promptForSelection("location", func() ([]*export.CachedItem, error) {
			return cache.ListLocations()
		})
		if err != nil {
			return fmt.Errorf("failed to select location: %w", err)
		}
		defaultLocation = location
	}
	if p.Options.Import == nil {
		p.Options.Import = &NautobotImportOpts{}
	}
	p.Options.DefaultLocation = defaultLocation

	if defaultRole == "" {
		role, err := promptForSelection("role", func() ([]*export.CachedItem, error) {
			return cache.ListRoles()
		})
		if err != nil {
			return fmt.Errorf("failed to select role: %w", err)
		}
		defaultRole = role
	}
	p.Options.DefaultRole = defaultRole

	if defaultStatus == "" {
		status, err := promptForSelection("status", func() ([]*export.CachedItem, error) {
			return cache.ListStatuses()
		})
		if err != nil {
			return fmt.Errorf("failed to select status: %w", err)
		}
		defaultStatus = status
	}
	p.Options.DefaultStatus = defaultStatus

	clog.Warn("WARNING: API token will be stored in config file. Use --token to override at runtime.")

	clog.Detail("Nautobot provider initialized with:")
	clog.Detail("  URL: %s", url)
	clog.Detail("  Default Location: %s", defaultLocation)
	clog.Detail("  Default Role: %s", defaultRole)
	clog.Detail("  Default Status: %s", defaultStatus)
	clog.Detail("  Merge: %v", merge)
	clog.Detail("  Dry Run: %v", dryRun)

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
