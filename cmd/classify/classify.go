/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023-2024 Hewlett Packard Enterprise Development LP
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
package classify

import (
	"fmt"
	"log"
	"regexp"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/spf13/cobra"
)

var (
	autoFlag   bool
	filterFlag string
	autoScore  int
)

// NewCommand creates the classify command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "classify",
		Short: "Classify unclassified devices in the inventory",
		Long: `Scan the local inventory for devices that have no device type slug or
model and interactively assign a type from the hardware library.

Examples:
  # Interactively classify all unclassified devices
  cani alpha classify

  # Auto-accept suggestions with score >= 90
  cani alpha classify --auto

  # Only classify devices matching a name pattern
  cani alpha classify --filter "ncn-.*"`,
		RunE: runClassify,
	}

	cmd.Flags().BoolVar(&autoFlag, "auto", false, "Auto-accept top suggestion if score >= threshold")
	cmd.Flags().IntVar(&autoScore, "auto-score", 90, "Minimum score for --auto acceptance (0-100)")
	cmd.Flags().StringVar(&filterFlag, "filter", "", "Regex filter on device name")

	return cmd
}

// runClassify is the main entry point for the classify command.
func runClassify(cmd *cobra.Command, args []string) error {
	if err := datastores.SetDeviceStore(cmd, args); err != nil {
		return fmt.Errorf("failed to set device store: %w", err)
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("failed to load inventory: %w", err)
	}

	unclassified := devicetypes.FindUnclassifiedDevices(inv)
	if len(unclassified) == 0 {
		log.Println("All devices are classified — nothing to do.")
		return nil
	}

	// Apply name filter if provided.
	if filterFlag != "" {
		unclassified, err = filterDevices(unclassified, filterFlag)
		if err != nil {
			return err
		}
		if len(unclassified) == 0 {
			log.Println("No unclassified devices match the filter.")
			return nil
		}
	}

	log.Printf("Found %d unclassified devices", len(unclassified))

	opts := devicetypes.ClassifyOptions{
		NoColor: config.Cfg.NoColor,
	}

	classified := 0
	skippedCount := 0
	for i, ud := range unclassified {
		log.Printf("[%d/%d] %s", i+1, len(unclassified), ud.Name)

		slug, err := resolveSlug(ud, opts)
		if err != nil {
			return fmt.Errorf("classification error for %s: %w", ud.Name, err)
		}
		if slug == "" {
			skippedCount++
			continue
		}

		device, ok := inv.Devices[ud.ID]
		if !ok || device == nil {
			log.Printf("  ! device %s vanished from inventory", ud.Name)
			continue
		}

		if err := devicetypes.ApplyDeviceType(device, slug); err != nil {
			log.Printf("  ! %s: failed to apply type %q: %v", ud.Name, slug, err)
			continue
		}
		classified++
	}

	log.Printf("Classified %d devices, skipped %d", classified, skippedCount)

	if classified > 0 {
		if err := datastores.Datastore.Save(inv); err != nil {
			return fmt.Errorf("failed to save inventory: %w", err)
		}
		log.Println("Inventory saved.")
	}

	return nil
}

// resolveSlug determines the slug for a device, either via auto mode or
// interactive prompt.
func resolveSlug(ud devicetypes.UnclassifiedDevice, opts devicetypes.ClassifyOptions) (string, error) {
	if autoFlag {
		suggestions := devicetypes.SuggestTypes(ud, 1)
		if len(suggestions) > 0 && suggestions[0].Score >= autoScore {
			slug := suggestions[0].Slug
			log.Printf("  → auto-assigned: %s (score %d)", slug, suggestions[0].Score)
			return slug, nil
		}
		// Fall through to interactive if no high-confidence match.
		if !autoFlag {
			return "", nil
		}
	}
	return devicetypes.PromptForDeviceType(ud, opts)
}

// filterDevices returns only devices whose Name matches the regex pattern.
func filterDevices(devices []devicetypes.UnclassifiedDevice, pattern string) ([]devicetypes.UnclassifiedDevice, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid filter regex %q: %w", pattern, err)
	}
	var filtered []devicetypes.UnclassifiedDevice
	for _, d := range devices {
		if re.MatchString(d.Name) {
			filtered = append(filtered, d)
		}
	}
	return filtered, nil
}
