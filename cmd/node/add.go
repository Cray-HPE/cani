package node

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/Cray-HPE/csminv/internal/csminv"
	hsm_client "github.com/Cray-HPE/csminv/pkg/hsm-client"
	"github.com/Cray-HPE/csminv/pkg/sls"
	base "github.com/Cray-HPE/hms-base/v2"
	sls_client "github.com/Cray-HPE/hms-sls/v2/pkg/sls-client"
	sls_common "github.com/Cray-HPE/hms-sls/v2/pkg/sls-common"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/Cray-HPE/hms-xname/xnametypes"

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
		// TODO How should idempotency be implemented? How should we be able to rerun this command if it partially failed?

		// Initialize the global viper
		v := viper.GetViper()
		v.BindPFlags(cmd.Flags())
		if v.GetBool("simulation-environment") {
			v.Set("hsm-url", "http://localhost:8080/apis/smd/hsm")
			v.Set("sls-url", "http://localhost:8080/apis/sls")
			v.Set("bss-url", "http://localhost:8080/apis/bss/boot")
		}
		v.BindEnv("token", "TOKEN")

		ci := csminv.New(v)

		//
		// Perform changes
		//

		// Validate xname is valid
		nodeXnameStr := v.GetString("xname")
		node := xnames.FromStringToStruct[xnames.Node](nodeXnameStr)
		if node == nil {
			// TODO better error message
			log.Fatalf("Invalid node xname provided: %s", nodeXnameStr)
		}

		// NodeBMC: Node -> NodeBMC
		nodeBMC := node.Parent()

		// Chassis xname: NodeBMC -> ComputeModule -> Chassis
		chassis := nodeBMC.Parent().Parent()

		// Cabinet Xname: Chassis -> Cabinet
		cabinet := chassis.Parent()

		// Check to see if the cabinet the node is being added to exists
		slsCabinet, err := ci.SLSClient.GetHardware(ci.Ctx, cabinet.String())
		if errors.Is(err, sls_client.ErrHardwareNotFound) {
			log.Fatalf("Cabinet containing node does not exist in SLS (%s)", cabinet.String())
		} else if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for cabinet (%s): %s", cabinet.String(), err)
		}

		ci.ValidateChassisXname(slsCabinet, chassis)

		// Class of the node can be inferred from the SLS cabinet object.
		// TODO
		// slsCabinet.Children

		// Update Cable topology????
		// TODO

		// Validate cable topology using CANU (add the connection if it doesn't already exist????
		// TODO

		// Check to see if the node object already exists in SLS
		slsExistingNode, err := ci.SLSClient.GetHardware(ci.Ctx, node.String())
		if errors.Is(err, sls_client.ErrHardwareNotFound) {
			// The node does not exist in SLS, this is an no op
			log.Printf("Node (%s) does not exist in SLS Hardware", node.String())
		} else if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for node (%s): %s", node.String(), err)
		} else {
			// A Node Object exist in SLS. This is a potentially unexpected state.
			// To detect if a the physical node was removed from the system, but the logical node remains the HSM component state could be used to determine if it was removed
			// Or add a extra property to SLS like "Empty"/CSMInvState, or there is no HSM component associated with the node
			log.Printf("Node (%s) does exists in SLS Hardware", node.String())
		}

		// Check to see if the node is present in HSM
		hsmNode, err := ci.HSMClient.GetStateComponent(ci.Ctx, node.String())
		if errors.Is(err, hsm_client.ErrNotFound) {
			// This means that the node has not been discovered by HSM. This can occur when the hardware was removed from the system, and the logical entry in
			// SLS remained.
			log.Print("Node does not exist in HSM State Components")
		} else if err != nil {
			log.Fatalf("Failed to retrieve HSM State Component for node (%s): %s", node.String(), err)
		} else {
			log.Printf("Node (%s) exists in HSM State Components with state (%s)", node.String(), hsmNode.State)
			if hsmNode.State != base.StateEmpty.String() {
				// This means that the user is trying to add a physical node to logical node that already has a node populated
				log.Fatal("Please remove existing physical node before attempting to add a physical node in the same place", node.String(), hsmNode.State)
			} else {
				// If the state is empty that means some one has disabled the redfish endpoint. Which is ok I think.
				// TODO consider if this scenario should be allowed.
			}
		}

		if slsCabinet.Class == sls_common.ClassRiver {
			//
			// Create SLS MgmtSwitchConnector for the node BMC. Require for River Nodes
			//

			// TODO ncn-m001 may be allowed to be added without a MgmtSwitchConnector

			switchPortXnameStr := v.GetString("switchport")
			if switchPortXnameStr == "" {
				log.Fatalf("Argument --switchport is required for adding river nodes")
			}

			switchPort := xnames.FromStringToStruct[xnames.MgmtSwitchConnector](switchPortXnameStr)
			if switchPort == nil {
				// TODO better error message
				log.Fatalf("Invalid MgmtSwitchConnector xname provided: %s", switchPortXnameStr)
			}

			// Ensure the parent Management Switch exists in SLS, otherwise this is a failure
			mgmtSwitch := switchPort.Parent()

			slsMgmtSwitch, err := ci.SLSClient.GetHardware(ci.Ctx, mgmtSwitch.String())
			if err != nil {
				log.Fatalf("Failed to retrieve SLS Hardware for MgmtSwitch (%s): %s", mgmtSwitch.String(), err)
			}

			slsMgmtSwitchEP, err := sls.DecodeHardwareExtraProperties2[sls_common.ComptypeMgmtSwitch](slsMgmtSwitch)
			if err != nil {
				log.Fatalf("Failed to decode MgmtSwitch hardware extra properties (%s): %s", mgmtSwitch.String(), err)
			}

			log.Printf("MgmtSwitch %s (%s) brand is %s", mgmtSwitch.String(), strings.Join(slsMgmtSwitchEP.Aliases, ", "), slsMgmtSwitchEP.Brand)

			// Ensure the desired switch port does not already exist in SLS
			slsExistingSwitchport, err := ci.SLSClient.GetHardware(ci.Ctx, switchPort.String())
			if errors.Is(err, sls_client.ErrHardwareNotFound) {
				// The switch port is not in use
				log.Printf("Switch port (%s) is currently not in use", switchPort.String())
			} else if err != nil {
				log.Fatalf("Failed to retrieve SLS Hardware for MgmtSwitchConnector (%s): %s", switchPort.String(), err)
			} else {
				// The switch port exists
				slsExistingSwitchportEP, err := sls.DecodeHardwareExtraProperties2[sls_common.ComptypeMgmtSwitchConnector](slsExistingSwitchport)
				if err != nil {
					log.Printf("Failed to decode extra properties for switch port (%s)", switchPort.String())
				}

				log.Fatalf("Error switch port (%s) is currently in use by (%s)", switchPort.String(), strings.Join(slsExistingSwitchportEP.NodeNics, ", "))
			}

			slsSwitchPort := sls_common.NewGenericHardware(switchPort.String(), sls_common.ClassRiver, sls_common.ComptypeMgmtSwitchConnector{
				VendorName: csminv.GetMgmtSwitchConnectorVendorName(*slsMgmtSwitchEP, *switchPort),
				NodeNics:   []string{nodeBMC.String()},
			})

			// Push the new switch port in to SLS
			log.Print("To be created switch port:", slsSwitchPort)
			err = ci.SLSClient.PutHardware(ci.Ctx, slsSwitchPort)
			if err != nil {
				log.Fatalf("Failed to put SLS Hardware for MgmtSwitchConnector (%s): %s", switchPort.String(), err)
			}
		}

		if slsExistingNode.Xname == "" {
			// Create SLS Node object, and MgmtSwitchConnector for the node BMC
			// - Ensure that there isn't a node that already exists with the given xname. What about if there is an logical
			//   SLS entry still left?
			// - Ensure the node is going to in cabinet that exists (if a liquid-cooled node ensure the chassis exists)

			slsNode := sls_common.NewGenericHardware(node.String(), slsCabinet.Class, sls_common.ComptypeNode{
				Role: "Unknown",
			})

			// Push the new node in to SLS
			log.Print("To be created node:", slsNode)
			err = ci.SLSClient.PutHardware(ci.Ctx, slsNode)
			if err != nil {
				log.Fatalf("Failed to put SLS Hardware for MgmtSwitchConnector (%s): %s", node.String(), err)
			}

		} else {
			// If we are here it means that we are reattaching a physical node to a logical node. I don't believe this means that anything needs to
			// actually happen at this point. The data in SLS should be good.
		}

		// At this point the node should be discoverable ny HSM
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
		v.BindEnv("token", "TOKEN")

		// Check to see if we have all require options
		role := v.GetString("role")
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

		ci := csminv.New(v)
		ctx := ci.Ctx

		//
		// Perform changes
		//

		// Validate xname is valid
		nodeXnameStr := v.GetString("xname")
		node := xnames.FromStringToStruct[xnames.Node](nodeXnameStr)
		if node == nil {
			// TODO better error message
			log.Fatalf("Invalid node xname provided: %s", nodeXnameStr)
		}

		// NodeBMC: Node -> NodeBMC
		nodeBMC := node.Parent()

		// Chassis xname: NodeBMC -> ComputeModule -> Chassis
		chassis := nodeBMC.Parent().Parent()

		// Cabinet Xname: Chassis -> Cabinet
		cabinet := chassis.Parent()

		// Check to see if the cabinet the node is being added to exists
		slsCabinet, err := ci.SLSClient.GetHardware(ctx, cabinet.String())
		if errors.Is(err, sls_client.ErrHardwareNotFound) {
			log.Fatalf("Cabinet containing node does not exist in SLS (%s)", cabinet.String())
		} else if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for cabinet (%s): %s", cabinet.String(), err)
		}

		ci.ValidateChassisXname(slsCabinet, chassis)

		// Retrieve HSM service values
		log.Print("Retrieving HSM Service Values")
		hsmServiceValues, err := ci.HSMClient.GetServiceValues(ctx)
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
		// This most likely should be a failure, if the role of the node is not Unknown
		log.Printf("Retrieving SLS Hardware object for %s", node.String())
		slsNode, err := ci.SLSClient.GetHardware(ctx, node.String())
		if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for node: %s", err)
		}
		log.Println(slsNode)

		slsNodeEP, err := sls.DecodeHardwareExtraProperties2[sls_common.ComptypeNode](slsNode)
		if err != nil {
			log.Fatalf("Failed to decode extra properties for node: %s", err)
		}

		// Enforce that we are only supposed to add a logical entity to hardware that is unknown.
		// If the desire is change an existing node logical data, either we need to delete and re-add the logical data
		// or add an update logical command
		if slsNodeEP.Role != "Unknown" {
			log.Fatalf("Adding logical node information to an already known node to the system: ", slsNode)
		}

		// Retrieve SLS hardware
		log.Print("Retrieving All SLS Hardware", node.String())
		slsAllHardware, err := ci.SLSClient.GetAllHardware(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve SLS Hardware for node: %s", err)
		}

		switch role {
		case "Compute":
			nodeNid := v.GetInt("nid")

			// Check to see if the given NID is currently in use in SLS
			// TODO replace this with a hardware search API call
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

				if ep.NID == nodeNid {
					log.Fatalf("Found another node with the provided NID in SLS: %s", hardware.Xname)
				}

			}

			// Check to see if the given NID is currently in use in HSM
			hsmMatches, err := ci.HSMClient.GetStateComponentsFilter(ctx, &hsm_client.StateComponentsSearchFilter{
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

			// Update the node's extra
			slsNodeEP.Role = role
			slsNodeEP.SubRole = ""
			slsNodeEP.NID = nodeNid
			slsNodeEP.Aliases = []string{fmt.Sprintf("nid%06d", nodeNid)}
		case "Application":
			// Check to see if the given SubRole is valid in HSM

			// Check to see if the given NID is currently in use in SLS
			// Check to see if the given NID is currently in use in HSM

			// Check to see if the given aliases are current in use in SLS
		case "Management":
			// TODO
		}

		// Patch SLS with updated logical information
		updatedSLSNode := sls_common.NewGenericHardware(node.String(), slsNode.Class, slsNodeEP)
		err = ci.SLSClient.PutHardware(ctx, updatedSLSNode)
		if err != nil {
			log.Fatal("Failed to update SLS with updated extra properties", err)
		}

		// TODO force set Role/SubRole/NID information in HSM if a state component exists, such as if the node has been already discovered from an unknown state
	},
}

func init() {
	AddNodeCmd.AddCommand(AddPhysicalCmd, AddLogicalCmd)

	//
	// Add physical
	//

	// This ensures the flags are displayed in teh order shown below
	AddPhysicalCmd.Flags().SortFlags = false

	AddPhysicalCmd.Flags().String("xname", "", "Node Xname (geolocation information)") // Make positional?
	AddPhysicalCmd.Flags().String("switchport", "", "BMC switch port (required for river)")

	// TODO make persistent flags
	AddPhysicalCmd.Flags().Bool("simulation-environment", false, "Advanced option: Target Hardware Simulation Environment")
	AddPhysicalCmd.Flags().String("sls-url", "https://api-gw-service-nmn.local/apis/sls", "Advanced option: URL to System Layout Service (SLS)")
	AddPhysicalCmd.Flags().String("bss-url", "https://api-gw-service-nmn.local/apis/bss", "Advanced option: URL to Boot Script Service (BSS)")
	AddPhysicalCmd.Flags().String("hsm-url", "https://api-gw-service-nmn.local/apis/smd", "Advanced option: URL to Hardware State Manager (HSM)")
	AddPhysicalCmd.Flags().String("token", "", "API Gateway Token")

	//
	// Add Logical
	//

	// This ensures the flags are displayed in the order shown below
	AddLogicalCmd.Flags().SortFlags = false

	AddLogicalCmd.Flags().String("xname", "", "Node Xname") // Make positional?
	AddLogicalCmd.Flags().String("role", "", "Node HSM Role")
	AddLogicalCmd.Flags().String("sub-role", "", "Node HSM Sub Role")
	AddLogicalCmd.Flags().StringSlice("alias", []string{}, "Node Aliases. Required for Application nodes, not use for other node types")
	AddLogicalCmd.Flags().Int("nid", -1, "Node HSM Role")

	// TODO make persistent flags
	AddLogicalCmd.Flags().Bool("simulation-environment", false, "Advanced option: Target Hardware Simulation Environment")
	AddLogicalCmd.Flags().String("sls-url", "https://api-gw-service-nmn.local/apis/sls", "Advanced option: URL to System Layout Service (SLS)")
	AddLogicalCmd.Flags().String("bss-url", "https://api-gw-service-nmn.local/apis/bss", "Advanced option: URL to Boot Script Service (BSS)")
	AddLogicalCmd.Flags().String("hsm-url", "https://api-gw-service-nmn.local/apis/smd", "Advanced option: URL to Hardware State Manager (HSM)")
	AddLogicalCmd.Flags().String("token", "", "API Gateway Token")

	AddLogicalCmd.MarkFlagRequired("xname")
	AddLogicalCmd.MarkFlagRequired("role")
}
