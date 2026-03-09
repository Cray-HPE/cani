package commands

import (
	"fmt"
	"log"

	"github.com/Cray-HPE/cani/internal/util/resolve"
	"github.com/Cray-HPE/cani/pkg/datastores"
	"github.com/spf13/cobra"
)

// wrapWithDeviceUpdateHook wraps the "update device" subcommand so that
// CSM-specific flags (--nid, --alias) are applied after the core update.
func wrapWithDeviceUpdateHook(cmd *cobra.Command) {
	orig := cmd.RunE
	if orig == nil {
		return
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := orig(cmd, args); err != nil {
			return err
		}
		return applyCSMDeviceMeta(cmd, args)
	}
}

// applyCSMDeviceMeta reads the --nid and --alias flags and persists them
// as CSM provider metadata on the target device.
func applyCSMDeviceMeta(cmd *cobra.Command, args []string) error {
	nidChanged := cmd.Flags().Changed("nid")
	aliasChanged := cmd.Flags().Changed("alias")
	if !nidChanged && !aliasChanged {
		return nil
	}

	inv, err := datastores.Datastore.Load()
	if err != nil {
		return fmt.Errorf("csm: failed to reload inventory: %w", err)
	}

	id, err := resolve.Device(inv, args[0])
	if err != nil {
		return fmt.Errorf("csm: resolving device for provider meta: %w", err)
	}

	device := inv.Devices[id]

	if nidChanged {
		nid, _ := cmd.Flags().GetInt("nid")
		device.SetProviderMeta("csm", "nid", nid)
		log.Printf("Set CSM nid=%d on device %s", nid, id)
	}
	if aliasChanged {
		alias, _ := cmd.Flags().GetString("alias")
		device.SetProviderMeta("csm", "aliases", []string{alias})
		log.Printf("Set CSM alias=%q on device %s", alias, id)
	}

	if err := datastores.Datastore.Save(inv); err != nil {
		return fmt.Errorf("csm: failed to save provider metadata: %w", err)
	}

	return nil
}
