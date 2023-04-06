package node

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	hsm_client "github.com/Cray-HPE/csminv/pkg/hsm-client"
	"github.com/Cray-HPE/csminv/pkg/sls"
	sls_client "github.com/Cray-HPE/csminv/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// AddPhysicalCmd represents the add node physical command
var AddPhysicalCmd = &cobra.Command{
	Use:   "physical",
	Short: "Add physical node to the inventory.",
	Long:  `Add physical node to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Add physical node")

		// Class of the node can be inferred from the SLS cabinet object.

		// Update Cable topology????

		// Validate cable topology using CANU (add the connection if it doesn't already exist????

		// Create SLS MgmtSwitchConnector for the node BMC
		// - Ensure the parent Management Switch exists in SLS, otherwise this is a failure
		// - Ensure the desired switch port does not already exist in SLS

		// Create SLS Node object for the node BMC
		// - Ensure that there isn't a node that already exists with the given xname
		// - Ensure the node is going to in cabinet that exists (if a liquid-cooled node ensure the chassis exists)

	},
}

// AddPhysicalCmd represents the add node physical command
var AddLogicalCmd = &cobra.Command{
	Use:   "logical",
	Short: "Add logical node to the inventory.",
	Long:  `Add logical node to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		//
		// Process configuration
		//

		// Initialize the global viper
		v := viper.GetViper()
		v.BindPFlags(cmd.Flags())
		if v.GetBool("simulation-environment") {
			v.Set("hsm-url", "http://localhost:8080/apis/smd/hsm")
			v.Set("sls-url", "http://localhost:8080/apis/sls")
			v.Set("bss-url", "http://localhost:8080/apis/bss/boot")
		}

		// Check to see if we have all require options
		role := v.Get("role")
		switch role {
		case "Compute":
			// Required flags
			// - Role
			// - NID
			// Optional Flags
			// - Alias
			if v.GetInt("nid") == -1 {
				log.Fatal("Missing required parameter --nid for Compute node")
			}
		case "Application":
			// Required flags
			// - Role
			// - SubRole
			// - Alias
			// Optional Fields
			// - NID
			if v.GetInt("sub-role") == -1 {
				log.Fatal("Missing required parameter --sub-role for Application node")
			}

			if len(v.GetStringSlice("alias")) == 0 {
				log.Fatal("Missing required parameter(s) --alias for Application node")
			}
		case "Management":
			// TODO
		}

		// Setup Context
		ctx := setupContext()

		// Retrieve API token
		// TODO align with what Jacob is doing
		token := os.Getenv("TOKEN")
		if token == "" {
			log.Fatal("Error environment variable TOKEN was not set")
		}

		//
		// Setup clients
		//

		// Setup HTTP client
		httpClient := retryablehttp.NewClient()
		httpClient.HTTPClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		// Setup SLS client
		slsURL := v.GetString("sls-url")
		slsClient := sls_client.NewSLSClient(slsURL, httpClient.StandardClient(), "").WithAPIToken(token)

		// Setup HSM client
		hsmURL := v.GetString("hsm-url")
		hsmClient := hsm_client.NewHSMClient(hsmURL, httpClient.StandardClient(), "").WithAPIToken(token)

		//
		// Perform changes
		//

		// Validate xname is valid
		nodeXnameStr := v.GetString("xname")
		nodeRaw := xnames.FromString(nodeXnameStr)
		if nodeRaw == nil {
			// TODO better error message
			log.Fatalf("Invalid node xname provided: %s", nodeXnameStr)
		}
		node, ok := nodeRaw.(xnames.Node)
		if !ok {
			log.Fatalf("Provided xname is not for a Node, but instead %s", nodeRaw.Type())
		}

		// NodeBMC: Node -> NodeBMC
		nodeBMC := node.Parent()
		log.Print("Node BMC: ", nodeBMC)

		// Cabinet xname: NodeBMC -> Chassis -> Cabinet
		cabinet := nodeBMC.Parent().Parent().Parent()
		log.Print("Cabinet: ", cabinet)

		// Retrieve HSM service values
		log.Print("Retrieving HSM Service Values")
		hsmServiceValues, err := hsmClient.GetServiceValues(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve service values from HSM: %v", err)
		}

		// Validate given node Role is valid
		roleValid := false
		for _, hsmRole := range hsmServiceValues.Role {
			if role == hsmRole {
				roleValid = true
				break
			}
		}

		if !roleValid {
			log.Fatalf("Invalid HSM Role provided: %s, Valid roles: %s", role, strings.Join(hsmServiceValues.Role, ", "))
		}

		// Check to see if the node has an entry in SLS
		// TODO what should be done if the logical entity is created before the physical data is added?
		// This most likely should be a failure
		log.Printf("Retrieving SLS Hardware object for %s", node.String())
		slsNode, err := slsClient.GetHardware(ctx, node.String())
		if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for node: %s", err)
		}
		log.Println(slsNode)

		// Retrieve SLS hardware
		log.Printf("Retrieving SLS Hardware object for %s", node.String())
		slsAllHardware, err := slsClient.GetAllHardware(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for node: %s", err)
		}

		switch role {
		case "Compute":
			nodeNid := v.GetInt("nid")

			// Check to see if the given NID is currently in use in SLS
			for _, hardware := range slsAllHardware {
				if hardware.TypeString != xnametypes.Node {
					continue
				}

				if hardware.Xname == slsNode.Xname {
					// Skip the node that we are adding
					continue
				}

				ep, err := sls.DecodeHardwareExtraProperties2[sls_common.ComptypeNode](hardware)
				if err != nil {
					log.Printf("Failed to decode extra properties for %s: %s", hardware.Xname, err)
					continue
				}
				// epRaw, err := sls.DecodeHardwareExtraProperties(hardware)
				// if err != nil {
				// 	log.Print("Failed to decode extra properties for %s: %s", hardware.Xname, err)
				// 	continue
				// }

				// ep, ok := epRaw.(sls_common.ComptypeNode)
				// if !ok {
				// 	log.Print("Node extra properties for %s have unexpected type %T", hardware.Xname, hardware)
				// 	continue
				// }

				if ep.NID == nodeNid {
					log.Fatalf("Found another node with the provided NID in SLS: %s", hardware.Xname)
				}

			}

			// Check to see if the given NID is currently in use in HSM
			hsmMatches, err := hsmClient.GetStateComponentsFilter(ctx, &hsm_client.StateComponentsSearchFilter{
				NID: []int{nodeNid},
			})
			if err != nil {
				log.Fatalf("Failed to query HSM State Components for NID %d: %s", nodeNid, err)
			}

			if len(hsmMatches.Components) > 0 {
				var xnames []string
				for _, component := range hsmMatches.Components {
					xnames = append(xnames, component.ID)
				}
				log.Fatalf("Found another node with the provided NID in HSM: %s", strings.Join(xnames, ","))
			}
		case "Application":
			// Check to see if the given SubRole is valid in HSM

			// Check to see if the given NID is currently in use in SLS
			// Check to see if the given NID is currently in use in HSM

			// Check to see if the given aliases are current in use in SLS
		case "Management":
			// TODO
		}
	},
}

func setupContext() context.Context {
	var cancel context.CancelFunc
	ctx, cancel := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		<-c

		// Cancel the context to cancel any in progress HTTP requests.
		cancel()
	}()

	return ctx
}

func init() {
	AddNodeCmd.AddCommand(AddPhysicalCmd, AddLogicalCmd)

	// This ensures the flags are displayed in teh order shown below
	AddLogicalCmd.Flags().SortFlags = false

	AddLogicalCmd.Flags().String("xname", "", "Node Xname") // Make positional?
	AddLogicalCmd.Flags().String("role", "", "Node HSM Role")
	AddLogicalCmd.Flags().String("sub-role", "", "Node HSM Sub Role")
	AddLogicalCmd.Flags().StringSlice("alias", []string{}, "Node Aliases. Required for Application nodes, not use for other node types")
	AddLogicalCmd.Flags().Int("nid", -1, "Node HSM Role")

	AddLogicalCmd.Flags().Bool("simulation-environment", false, "Advanced option: Target Hardware Simulation Environment")
	AddLogicalCmd.Flags().String("sls-url", "https://api-gw-service-nmn.local/apis/sls", "Advanced option: URL to System Layout Service (SLS)")
	AddLogicalCmd.Flags().String("bss-url", "https://api-gw-service-nmn.local/apis/bss", "Advanced option: URL to Boot Script Service (BSS)")
	AddLogicalCmd.Flags().String("hsm-url", "https://api-gw-service-nmn.local/apis/smd", "Advanced option: URL to Hardware State Manager (HSM)")

	AddLogicalCmd.MarkFlagRequired("xname")
	AddLogicalCmd.MarkFlagRequired("role")
}
