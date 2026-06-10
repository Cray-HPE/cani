package export

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Cray-HPE/cani/pkg/devicetypes"
	"github.com/Cray-HPE/cani/pkg/provider/csm/client"
	import_ "github.com/Cray-HPE/cani/pkg/provider/csm/import"
	"github.com/spf13/cobra"
)

// Export reconciles the local CANI inventory with the live CSM system.
// When --commit is set it pushes changes; otherwise it previews them.
// When no API target is specified, it writes CSV to stdout.
func Export(cmd *cobra.Command, inventory devicetypes.Inventory) error {
	format, _ := cmd.Flags().GetString("format")

	switch format {
	case "sls-json":
		return exportSLSJSON(cmd, inventory)
	case "csv", "":
		return exportCSVOrAPI(cmd, inventory)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportSLSJSON handles the sls-json format by reconciling and dumping JSON.
func exportSLSJSON(cmd *cobra.Command, inventory devicetypes.Inventory) error {
	c, err := buildClient(cmd)
	if err != nil {
		return fmt.Errorf("building CSM client: %w", err)
	}

	ignoreValidation, _ := cmd.Flags().GetBool("ignore-validation")

	state, err := fetchCurrentState(c)
	if err != nil {
		return err
	}

	expected := buildExpectedHardware(inventory, state.Hardware)

	if !ignoreValidation {
		if err := validateSLSHardware(expected, inventory); err != nil {
			return err
		}
	}

	return writeSLSJSON(os.Stdout, expected)
}

// exportCSVOrAPI either writes CSV to stdout or does API reconcile.
func exportCSVOrAPI(cmd *cobra.Command, inventory devicetypes.Inventory) error {
	useAPI, err := shouldUseAPI(cmd)
	if err != nil {
		return err
	}

	if !useAPI {
		return exportCSV(cmd, inventory)
	}

	c, err := buildClient(cmd)
	if err != nil {
		return fmt.Errorf("building CSM client: %w", err)
	}

	commit, _ := cmd.Flags().GetBool("commit")
	dryrun, _ := cmd.Flags().GetBool("dryrun")
	return reconcile(c, inventory, commit, dryrun)
}

// exportCSV writes the inventory as CSV to stdout, filtered by
// --headers, --type, and --all flags.
func exportCSV(cmd *cobra.Command, inventory devicetypes.Inventory) error {
	headersStr, _ := cmd.Flags().GetString("headers")
	typesStr, _ := cmd.Flags().GetString("type")
	allTypes, _ := cmd.Flags().GetBool("all")

	headers := splitCSV(headersStr)
	var types []string
	if !allTypes {
		types = splitCSV(typesStr)
	}

	return ExportCSV(os.Stdout, inventory, headers, types)
}

// splitCSV splits a comma-separated string and trims whitespace.
func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// reconcile compares the local inventory to the live SLS state and
// applies changes when commit is true. When dryrun is true, changes
// are previewed but never applied, even if commit is also true.
func reconcile(c *client.Client, inventory devicetypes.Inventory, commit, dryrun bool) error {
	if dryrun {
		log.Println("Performing dryrun no changes will be applied to the system!")
	}

	state, err := fetchCurrentState(c)
	if err != nil {
		return err
	}

	expected := buildExpectedHardware(inventory, state.Hardware)

	// Enrich new cabinet entries with network metadata before diffing
	// so the PUT payload includes Networks in ExtraProperties.
	enrichCabinetNetworks(expected, state.Networks, inventory)

	changes := diffHardware(expected, state.Hardware)

	log.Printf("Reconcile: %d devices, %d to add, %d to update",
		len(inventory.Devices), len(changes.Added), len(changes.Changed))

	var stats reconcileStats
	if commit && !dryrun {
		if err := applyChanges(c, changes, &stats); err != nil {
			return err
		}
		if err := reconcileNetworks(c, state.Networks, changes, inventory, &stats); err != nil {
			return fmt.Errorf("failed to reconcile network changes: %w", err)
		}
	} else if !dryrun {
		log.Println("Dry-run mode: no changes pushed (pass --commit to apply)")
	}

	printSummary(os.Stdout, inventory, stats)

	if dryrun {
		log.Println("Dryrun enabled, no changes performed!")
	}

	return nil
}

// fetchCurrentState retrieves the current SLS dumpstate.
func fetchCurrentState(c *client.Client) (*import_.SlsDumpstate, error) {
	slsURL := c.BaseURLSLS + "/dumpstate"
	log.Printf("Fetching current SLS state from %s", slsURL)

	data, err := c.Get(slsURL)
	if err != nil {
		return nil, fmt.Errorf("fetching SLS dumpstate: %w", err)
	}

	state, err := import_.ParseSlsDumpstate(data)
	if err != nil {
		return nil, fmt.Errorf("parsing SLS dumpstate: %w", err)
	}

	return state, nil
}

// applyChanges pushes added and changed hardware to SLS via PUT.
func applyChanges(c *client.Client, changes hardwareChanges, stats *reconcileStats) error {
	for _, hw := range changes.Added {
		if err := putHardware(c, hw); err != nil {
			return err
		}
		stats.PutCount++
	}

	for _, hw := range changes.Changed {
		if err := putHardware(c, hw); err != nil {
			return err
		}
		stats.PutCount++
	}

	for _, hw := range changes.Removed {
		if err := deleteHardware(c, hw.Xname); err != nil {
			return err
		}
		stats.DeleteCount++
	}

	return nil
}

// putHardware sends a PUT request to SLS for a single hardware entry.
func putHardware(c *client.Client, hw import_.SlsHardware) error {
	url := c.BaseURLSLS + "/hardware/" + hw.Xname
	body, err := marshalHardware(hw)
	if err != nil {
		return err
	}

	log.Printf("PUT %s", url)
	if _, err := c.Put(url, body); err != nil {
		return fmt.Errorf("PUT %s: %w", url, err)
	}
	return nil
}

// deleteHardware sends a DELETE request to SLS for a single xname.
func deleteHardware(c *client.Client, xname string) error {
	url := c.BaseURLSLS + "/hardware/" + xname
	log.Printf("DELETE %s", url)
	if _, err := c.Delete(url); err != nil {
		return fmt.Errorf("DELETE %s: %w", url, err)
	}
	return nil
}

// shouldUseAPI returns true when the user wants a live API export.
// It checks for -S, explicit --csm-api-host, or --commit.
func shouldUseAPI(cmd *cobra.Command) (bool, error) {
	sim, _ := cmd.Flags().GetBool("use-simulator")
	if sim {
		return true, nil
	}
	if cmd.Flags().Changed("csm-api-host") {
		return true, nil
	}
	commit, _ := cmd.Flags().GetBool("commit")
	return commit, nil
}

// buildClient creates an authenticated client.Client from CLI flags.
func buildClient(cmd *cobra.Command) (*client.Client, error) {
	sim, _ := cmd.Flags().GetBool("use-simulator")
	insecure, _ := cmd.Flags().GetBool("insecure")
	host, _ := cmd.Flags().GetString("csm-api-host")
	user, _ := cmd.Flags().GetString("csm-keycloak-username")
	pass, _ := cmd.Flags().GetString("csm-keycloak-password")
	pass = client.ResolveSecret(pass, "csm-keycloak-password", "CANI_CSM_KEYCLOAK_PASSWORD")
	slsURL, _ := cmd.Flags().GetString("csm-url-sls")
	hsmURL, _ := cmd.Flags().GetString("csm-url-hsm")
	caCert, _ := cmd.Flags().GetString("csm-ca-cert")
	k8sPods, _ := cmd.Flags().GetString("csm-k8s-pods-cidr")
	k8sSvcs, _ := cmd.Flags().GetString("csm-k8s-services-cidr")
	kubeconfig, _ := cmd.Flags().GetString("csm-kube-config")
	secretName, _ := cmd.Flags().GetString("csm-secret-name")
	clientID, _ := cmd.Flags().GetString("csm-client-id")
	clientSecret, _ := cmd.Flags().GetString("csm-client-secret")
	clientSecret = client.ResolveSecret(clientSecret, "csm-client-secret", "CANI_CSM_CLIENT_SECRET")

	opts := client.Options{
		ProviderHost:       host,
		InsecureSkipVerify: insecure,
		CaCertPath:         caCert,
		TokenUsername:      user,
		TokenPassword:      pass,
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		BaseURLSLS:         slsURL,
		BaseURLHSM:         hsmURL,
		K8sPodsCidr:        k8sPods,
		K8sServicesCidr:    k8sSvcs,
		KubeConfig:         kubeconfig,
		SecretName:         secretName,
		UseSimulation:      sim,
	}
	return client.NewClient(opts)
}
