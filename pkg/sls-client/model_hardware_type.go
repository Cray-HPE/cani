/*
 * System Layout Service
 *
 * System Layout Service (SLS) holds information on the complete, designed system. SLS gets this information from an input file on the system. Besides information like what hardware should be present in a system, SLS also stores information about what network connections exist and what power connections exist. SLS details the physical locations of network hardware, compute nodes and cabinets. Further, it stores information about the network, such as which port on which switch should be connected to each compute node. The API allows updating this information as well.   Note that SLS is not responsible for verifying that the system is set up correctly. It only lets the Shasta system know what the system should be configured with. SLS does not store the details of the actual hardware like hardware identifiers. Instead it stores a generalized abstraction of the system that other services may use. SLS thus does not need to change as hardware within the system is replaced. Interaction with SLS is required if the system setup changes – for example, if system cabling is altered or during  installation, expansion, or reduction. SLS does not interact with the hardware.  Each object in SLS has the following basic properties: * Parent – Each object in SLS has a parent object except the system root (s0). * Children – Objects may have children. * xname – Every object has an xname – a unique identifier for that object. * Type – a hardware type like \"comptype_ncard\", \"comptype_cabinet\". * Class – kind of hardware like \"River\" or \"Mountain\" * TypeString – a human readable type like \"Cabinet\"  Some objects may have additional properties depending on their type. For example, additional properties for cabinets include \"Network\", \"IP6Prefix\", \"IP4Base\", \"MACprefix\" etc.   ## Resources  ### /hardware  Create hardware entries in SLS. This resource can be used when you add new components or expand your system. Interaction with this resource is not required if a component is removed or replaced.  ### /hardware/{xname}  Retrieve, update, or delete information about specific xnames.   ### /search/hardware  Uses HTTP query parameters to find hardware entries with matching properties. Returns a JSON list of xnames. If multiple query parameters are passed, any returned hardware must match all parameters.  For example, a query string of \"?parent=x0\" would return a list of all children of cabinet x0. A query string of \"?type=comptype_node\" would return a list of all compute nodes.  Valid query parameters are: xname, parent, class, type, power_connector, node_nics, networks, peers.   ### /search/networks  Uses HTTP query parameters to find network entries with matching properties.  ### /networks  Create new network objects or retrieve networks available in the system.  ### /networks/{network}  Retrieve, update, or delete information about specific networks.  ### /dumpstate  Dumps the current database state of the service. This may be useful when you are backing up the system or planning a reinstall of the system.  ### /loadstate  Upload and overwrite the current database with the contents of the posted data. The posted data should be a state dump from /dumpstate. This may be useful to restore the SLS database after you have reinstalled the system.   ## Workflows  ### Backup and Restore the SLS Database for Reinstallation  #### GET /dumpstate  Perform a dump of the current state of the SLS data. This should be done before reinstalling the system. The database dump is a JSON blob in an SLS-specific format.  #### POST /loadstate  Reimport the dump from /dumpstate and restore the SLS database after reinstall.      ### Expand System  #### POST /hardware  Add the new hardware objects.  #### GET /hardware/{xname}  Review hardware properties of the xname from the JSON array.  ### Remove Hardware  #### DELETE /hardware  Remove hardware from SLS  ### Modify Hardware Properties  #### PATCH /hardware  Modify hardware properties in SLS. Only additional properties can be modified. Basic properties like xname, parent, children, type, class, typestring cannot be modified.
 *
 * API version: 0.1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package sls_client

// HardwareType : The type of this piece of hardware.  This is an optional hint during upload; it will be ignored if it does not match the xname
type HardwareType string

// List of hardware_type
const (
	CDU_HardwareType                     HardwareType = "comptype_cdu"
	CDU_MGMT_SWITCH_HardwareType         HardwareType = "comptype_cdu_mgmt_switch"
	CAB_CDU_HardwareType                 HardwareType = "comptype_cab_cdu"
	CABINET_HardwareType                 HardwareType = "comptype_cabinet"
	CAB_PDU_CONTROLLER_HardwareType      HardwareType = "comptype_cab_pdu_controller"
	CAB_PDU_HardwareType                 HardwareType = "comptype_cab_pdu"
	CAB_PDU_NIC_HardwareType             HardwareType = "comptype_cab_pdu_nic"
	CAB_PDU_OUTLET_HardwareType          HardwareType = "comptype_cab_pdu_outlet"
	CAB_PDU_PWR_CONNECTOR_HardwareType   HardwareType = "comptype_cab_pdu_pwr_connector"
	CHASSIS_HardwareType                 HardwareType = "comptype_chassis"
	CHASSIS_BMC_HardwareType             HardwareType = "comptype_chassis_bmc"
	CMM_RECTIFIER_HardwareType           HardwareType = "comptype_cmm_rectifier"
	CMM_FPGA_HardwareType                HardwareType = "comptype_cmm_fpga"
	CEC_HardwareType                     HardwareType = "comptype_cec"
	COMPMOD_HardwareType                 HardwareType = "comptype_compmod"
	RTRMOD_HardwareType                  HardwareType = "comptype_rtrmod"
	NCARD_HardwareType                   HardwareType = "comptype_ncard"
	BMC_NIC_HardwareType                 HardwareType = "comptype_bmc_nic"
	NODE_ENCLOSURE_HardwareType          HardwareType = "comptype_node_enclosure"
	COMPMOD_POWER_CONNECTOR_HardwareType HardwareType = "comptype_compmod_power_connector"
	NODE_HardwareType                    HardwareType = "comptype_node"
	NODE_PROCESSOR_HardwareType          HardwareType = "comptype_node_processor"
	NODE_NIC_HardwareType                HardwareType = "comptype_node_nic"
	NODE_HSN_NIC_HardwareType            HardwareType = "comptype_node_hsn_nic"
	DIMM_HardwareType                    HardwareType = "comptype_dimm"
	NODE_ACCEL_HardwareType              HardwareType = "comptype_node_accel"
	NODE_FPGA_HardwareType               HardwareType = "comptype_node_fpga"
	HSN_ASIC_HardwareType                HardwareType = "comptype_hsn_asic"
	RTR_FPGA_HardwareType                HardwareType = "comptype_rtr_fpga"
	RTR_TOR_FPGA_HardwareType            HardwareType = "comptype_rtr_tor_fpga"
	RTR_BMC_HardwareType                 HardwareType = "comptype_rtr_bmc"
	RTR_BMC_NIC_HardwareType             HardwareType = "comptype_rtr_bmc_nic"
	HSN_BOARD_HardwareType               HardwareType = "comptype_hsn_board"
	HSN_LINK_HardwareType                HardwareType = "comptype_hsn_link"
	HSN_CONNECTOR_HardwareType           HardwareType = "comptype_hsn_connector"
	HSN_CONNECTOR_PORT_HardwareType      HardwareType = "comptype_hsn_connector_port"
	MGMT_SWITCH_HardwareType             HardwareType = "comptype_mgmt_switch"
	MGMT_SWITCH_CONNECTOR_HardwareType   HardwareType = "comptype_mgmt_switch_connector"
	HL_SWITCH_HardwareType               HardwareType = "comptype_hl_switch"
)
