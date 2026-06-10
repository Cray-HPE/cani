package import_

import (
	"fmt"
	"log"
	"os"

	"github.com/Cray-HPE/cani/pkg/provider/csm/client"
	"github.com/spf13/cobra"
)

// providerGetter returns the Csm singleton to store raw SLS/SMD data.
// Set by the parent package's init() to break the import cycle.
var providerGetter func() interface {
	ClearRawData()
	SetSls(sls *SlsDumpstate)
	SetSmd(smd *SmdComponentList)
}

// SetProviderGetter allows the parent package to provide singleton access.
func SetProviderGetter(getter func() interface {
	ClearRawData()
	SetSls(sls *SlsDumpstate)
	SetSmd(smd *SmdComponentList)
}) {
	providerGetter = getter
}

// Import reads SLS and SMD/HSM data from either JSON files or the live
// CSM API, parses them, and stores the raw data on the provider singleton.
// No transformation is done here; that happens in the transform step.
func Import(cmd *cobra.Command, args []string) error {
	p := providerGetter()
	p.ClearRawData()

	useAPI, err := shouldUseAPI(cmd)
	if err != nil {
		return err
	}

	if useAPI {
		return importFromAPI(cmd, p)
	}
	return importFromFiles(cmd, p)
}

// shouldUseAPI returns true when the user wants a live API import
// (--use-simulator or explicit --csm-api-host).
func shouldUseAPI(cmd *cobra.Command) (bool, error) {
	sim, _ := cmd.Flags().GetBool("use-simulator")
	if sim {
		return true, nil
	}
	return cmd.Flags().Changed("csm-api-host"), nil
}

// ShouldUseAPI is an exported wrapper around shouldUseAPI.
func ShouldUseAPI(cmd *cobra.Command) (bool, error) {
	return shouldUseAPI(cmd)
}

// importFromAPI fetches SLS and SMD data from the live CSM APIs.
func importFromAPI(cmd *cobra.Command, p interface {
	SetSls(sls *SlsDumpstate)
	SetSmd(smd *SmdComponentList)
},
) error {
	c, err := buildClient(cmd)
	if err != nil {
		return fmt.Errorf("building CSM client: %w", err)
	}

	// Fetch SLS dumpstate
	slsURL := c.BaseURLSLS + "/dumpstate"
	log.Printf("Fetching SLS dumpstate from %s", slsURL)
	slsData, err := c.Get(slsURL)
	if err != nil {
		return fmt.Errorf("fetching SLS dumpstate: %w", err)
	}
	sls, err := ParseSlsDumpstate(slsData)
	if err != nil {
		return err
	}
	p.SetSls(sls)
	log.Printf("Imported SLS: %d hardware entries, %d networks",
		len(sls.Hardware), len(sls.Networks))

	// Fetch SMD state components
	smdURL := c.BaseURLHSM + "/State/Components"
	log.Printf("Fetching SMD state components from %s", smdURL)
	smdData, err := c.Get(smdURL)
	if err != nil {
		return fmt.Errorf("fetching SMD state components: %w", err)
	}
	smd, err := ParseSmdComponents(smdData)
	if err != nil {
		return err
	}
	p.SetSmd(smd)
	log.Printf("Imported SMD: %d components", len(smd.Components))

	return nil
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

// importFromFiles reads SLS and SMD data from local JSON files.
func importFromFiles(cmd *cobra.Command, p interface {
	SetSls(sls *SlsDumpstate)
	SetSmd(smd *SmdComponentList)
},
) error {
	slsFile, _ := cmd.Flags().GetString("sls-file")
	smdFile, _ := cmd.Flags().GetString("smd-file")

	if slsFile == "" {
		return fmt.Errorf("--sls-file is required for file-based CSM import")
	}

	// Parse SLS dumpstate
	sls, err := readAndParseSls(slsFile)
	if err != nil {
		return err
	}
	p.SetSls(sls)
	log.Printf("Imported SLS: %d hardware entries, %d networks",
		len(sls.Hardware), len(sls.Networks))

	// Parse SMD components (optional — system may not have HSM data)
	if smdFile != "" {
		smd, err := readAndParseSmd(smdFile)
		if err != nil {
			return err
		}
		p.SetSmd(smd)
		log.Printf("Imported SMD: %d components", len(smd.Components))
	} else {
		log.Println("No --smd-file provided; skipping SMD import")
	}

	return nil
}

func readAndParseSls(path string) (*SlsDumpstate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading SLS file %s: %w", path, err)
	}
	return ParseSlsDumpstate(data)
}

func readAndParseSmd(path string) (*SmdComponentList, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading SMD file %s: %w", path, err)
	}
	return ParseSmdComponents(data)
}
