/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	x "github.com/Cray-HPE/hms-xname/xnames"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// addBladeCmd represents the cabinet add command
var addBladeCmd = &cobra.Command{
	Use:   "blade",
	Short: "Add blades to the inventory.",
	Long:  `Add blades to the inventory.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		clearRedfishEvents()
		// addBlade(args)
	},
}

var (
	kubeconfig string
	clientset  *kubernetes.Clientset
	// getHmsDiscoveryJobName = "hms-discovery"
	getHmsDiscoveryJobName = "hello"
	hmsNamespace           = "services"
	liquidCooledDocsFlag   = "liquid-cooled-docs"
	airCooledDocsFlag      = "air-cooled-docs"
	addBladeStatusFlag     = "status"
	liquidCooledFlag       = "liquid-cooled"
	airCooledFlag          = "air-cooled"
	addBladeStatus         bool
	liquidCooled           bool
	liquidCooledDocs       bool
	addLiquidBladeLink     = "https://github.com/Cray-HPE/docs-csm/blob/main/operations/node_management/Adding_a_Liquid-cooled_blade_to_a_System.md"
	airCooled              bool
	airCooledDocs          bool
	addAirBladeLink        = "https://github.com/Cray-HPE/docs-csm/blob/main/operations/node_management/Add_a_Standard_Rack_Node.md"
	xnames                 []string
	addBladeReqs           = []string{
		"cray CLI is initialized",
		"DVS is on NMN or HMN",
		"Use existing cabinet",
		"Slingshot fabric is configured with the desired topology",
		"SLS has the desired HSN configuration",
		"Check HSN and link status",
	}
)

func init() {
	addCmd.AddCommand(addBladeCmd)

	addBladeCmd.PersistentFlags().BoolVarP(&addBladeStatus, addBladeStatusFlag, "s", false, "Check status of the system to see if it is ready to add a blade.")
	addBladeCmd.PersistentFlags().BoolVarP(&liquidCooledDocs, liquidCooledDocsFlag, "L", false, "Show docs for adding a liquid-cooled blade.")
	addBladeCmd.PersistentFlags().BoolVarP(&airCooledDocs, airCooledDocsFlag, "A", false, "Show docs for adding an air-cooled blade.")
	addBladeCmd.PersistentFlags().BoolVarP(&liquidCooled, liquidCooledFlag, "l", false, "Add liquid-cooled blades to the inventory.")
	addBladeCmd.PersistentFlags().BoolVarP(&airCooled, airCooledFlag, "a", false, "Add air-cooled blades to the inventory.")
	addBladeCmd.PersistentFlags().StringSliceVarP(&xnames, "xnames", "x", []string{}, "Comma-separate xnames of the blade to add.")

	// addBladeCmd.PersistentFlags().StringVarP(&kubeconfig, "kubeconfig", "k", "", "Path to Kubernetes config file")
	// // create the config object from kubeconfig

	// kconfig, err := clientcmd.BuildConfigFromFlags(fmt.Sprintf(
	// 	"https://%s:%s",
	// 	os.Getenv("KUBERNETES_SERVICE_HOST"),
	// 	os.Getenv("KUBERNETES_SERVICE_PORT")),
	// 	kubeconfig)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to use kubernetes config.")
	// }

	// // create clientset (set of muliple clients) for each Group
	// clientset, err = kubernetes.NewForConfig(kconfig)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to create kubernetes clientset.")
	// }

}

// checkAddBladeStatus checks the status of the system to see if it is ready to add a blade.
func checkAddBladeStatus() {
	for _, req := range addBladeReqs {
		fmt.Printf("%s%s\n", padRight(no, " ", 12), req)
	}
}

// suspendHmsDiscoveryJob suspends the hms-discovery cron job to disable it.
func suspendHmsDiscoveryJob() {
	// patch the hms-discovery cron job to suspend it
	_, _, err := shell("kubectl", []string{"-n", "services", "patch", "cronjobs", getHmsDiscoveryJobName, "-p", "{\"spec\" : {\"suspend\" : true }}"})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to suspend hms-discovery cron job.")
	}
	// TODO: Use the actual k8s library for this
	// cronjob, err := clientset.BatchV1beta1().CronJobs(hmsNamespace).Patch(
	// 	context.TODO(),
	// 	getHmsDiscoveryJobName,
	// 	types.MergePatchType,
	// 	[]byte(`{"spec":{"suspend":true}}`),
	// 	v1.PatchOptions{})
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to suspend hms-discovery cron job.")
	// }
}

func getHmsDiscoveryJob() {
	// get the hms-discovery cron job
	_, _, err := shell("kubectl", []string{"-n", "services", "get", "cronjobs", getHmsDiscoveryJobName})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to suspend hms-discovery cron job.")
	}
	// cronjob, err := clientset.BatchV1beta1().CronJobs(hmsNamespace).Get(
	// 	context.Background(),
	// 	getHmsDiscoveryJobName,
	// 	v1.GetOptions{})
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to get hms-discovery cron job.")
	// }
}

// isChassisSlotPopulated determines if the destination chassis slot is populated.
func isChassisSlotPopulated(xname string) (bool, error) {
	// code to check if chassis slot is populated
	// ...
	stdout, _, err := shell("cray", []string{"hsm", "state", "components", "describe", xname, "--format", "json"})
	if err != nil {
		return false, err
	}

	// Unmarshal response into a map for simple parsing without defining structs
	var hsmInfo map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &hsmInfo)
	if err != nil {
		return false, err
	}
	if hsmInfo["State"] == "Populated" {
		return false, errors.New("Chassis slot is populated")
	}
	return true, nil
}

// isChassisSlotOff determines if the destination chassis slot is off.
func isChassisSlotOff(xname string) (bool, error) {
	// code to check if chassis slot is off
	// ...
	stdout, _, err := shell("cray", []string{"capmc", "get_xname_status", "create", "--xnames", xname, "--format", "json"})
	if err != nil {
		return false, err
	}
	// Unmarshal response into a map for simple parsing without defining structs
	var capInfo map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &capInfo)
	if err != nil {
		return false, err
	}

	if capInfo["e"].(int) != 0 ||
		capInfo["err_msg"].(string) != "Success" ||
		capInfo["off"].([]string) != nil {
		return false, errors.New("Slot is not off")
	}

	return true, nil
}

func powerOffChassisSlot(xname string) error {
	// code to power off chassis slot
	_, _, err := shell("cray", []string{"capmc", "xname_off", "create", "--xnames", xname, "--recursive", "true"})
	if err != nil {
		return err
	}
	return nil
}

// drain and fill coolant during a swap
// install the blade

// getApiGwToken gets the API gateway token.
func getApiGwToken() (*core.Secret, error) {
	secret, err := clientset.CoreV1().Secrets(hmsNamespace).Get(context.TODO(), "", v1.GetOptions{})
	if err != nil {
		return secret, err
	}
	return secret, nil
}

// isDvsOnHsn determines if DVS is operating over the NMN or HMN.
func isDvsOnHsn(xname string) (bool, error) {
	// When DVS is operating over the NMN and a blade is being replaced, the mapping of node component name (xname) to node IP address must be preserved. Kea automatically adds entries to the HSM ethernetInterfaces table when DHCP lease is provided (about every 5 minutes)
	_, stderr, err := shell("lnetctl", []string{"net", "show"})
	if err != nil {
		// if lnetctl fails, check a compute node's boot parameters
		fmt.Println(stderr)
		log.Warn().Msg("Failed to run 'lnetctl net show'")
		stdout, stderr, err := shell("cray", []string{"bss", "bootparameters", "list", "--hosts", xname, "--format", "json"})
		if err != nil {
			// if lnetctl fails, check a compute node's boot parameters
			fmt.Println(stderr)
			log.Fatal().Msg("Could not determine if DVS is operating over the NMN or HMN")
		}
		// Unmarshal response into a map for simple parsing without defining structs
		var bootparameters []map[string]interface{}
		err = json.Unmarshal([]byte(stdout), &bootparameters)
		if err != nil {
			return false, err
		}

		params := fmt.Sprintf(bootparameters[0]["params"].(string))
		if strings.Contains(params, "hsn0") ||
			strings.Contains(params, "hsn1") {
			return true, nil
		}
	}

	return false, nil
}

func getNodeMaintenanceEthernetInterfaces() ([]string, error) {
	// code to get node maint ethernet interface
	// ...
	stdout, _, err := shell("cat", []string{"testdata/fixtures/hsm/inventory_ethernetInterfaces.json"})
	// stdout, _, err := shell("cray", []string{"hsm", "inventory", "ethernetInterfaces", "--format", "json"})
	if err != nil {
		return nil, err
	}

	// Unmarshal response into a map for simple parsing without defining structs
	var ethInterfaces []map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &ethInterfaces)
	if err != nil {
		return nil, err
	}

	var ei []string
	for _, ethInterface := range ethInterfaces {
		if ethInterface["Description"].(string) == "Node Maintenance Network" {
			ei = append(ei, ethInterface["Name"].(string))
		}
	}
	return ei, nil
}

func deleteK8sPod() {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cray-dhcp-kea",
	})
	if err != nil {
		log.Fatal().Err(err)
	}
	for _, i := range pods.Items {
		err = clientset.CoreV1().Pods(hmsNamespace).Delete(context.TODO(), i.Name, v1.DeleteOptions{})
		if err != nil {
			log.Fatal().Err(err)
		}
	}
}

// preserveXnameToIpMapping preserves the mapping of node component name (xname) to node IP address.
func preserveXnameToIpMapping() error {
	// Get Node Maintenance Network ethernet interfaces
	ei, err := getNodeMaintenanceEthernetInterfaces()
	if err != nil {
		log.Fatal().Err(err).Msg("Could not get node maintenance ethernet interfaces")
	}
	// delete cray-dhcp-kea pods
	deleteK8sPod()
	// Remove them from HSM
	for _, einterface := range ei {
		_, _, err := shell("cray", []string{"hsm", "inventory", "ethernetInterfaces", "delete", einterface})
		if err != nil {
			return err
		}
	}
	ok := YesNoPrompt("Is this node replacing an existing one", false)
	if ok {
		fmt.Println("Values from the previous blade must be recorded")
		mac := StringPrompt("MAC address of the previous blade")
		ipstring := StringPrompt("Comma-separtated IP addresses of the previous blade")
		ips := strings.Split(ipstring, ",")
		xname := StringPrompt("Xname of the previous blade")
		// curl -H "Authorization: Bearer ${TOKEN}" -L -X POST 'https://api-gw-service-nmn.local/apis/smd/hsm/v2/Inventory/EthernetInterfaces' -H 'Content-Type: application/json' --data-raw "{
		//     \"Description\": \"Node Maintenance Network\",
		//     \"MACAddress\": \"$MAC\",
		//     \"IPAddresses\": [
		//         {
		//             \"IPAddress\": \"$IP_ADDRESS\"
		//         }
		//     ],
		//     \"ComponentID\": \"$XNAME\"
		// }"
		data := Payload{
			Description: "Node Maintenance Network",
			MACAddress:  mac,
			IPAddresses: ips,
			ComponentID: xname,
		}
		payloadBytes, err := json.Marshal(data)
		if err != nil {
			log.Fatal().Err(err)
		}
		body := bytes.NewReader(payloadBytes)

		req, err := http.NewRequest("POST", "https://api-gw-service-nmn.local/apis/smd/hsm/v2/Inventory/EthernetInterfaces", body)
		if err != nil {
			log.Fatal().Err(err)
		}
		req.Header.Set("Authorization", os.ExpandEnv("Bearer ${TOKEN}"))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Fatal().Err(err)
		}
		defer resp.Body.Close()

	} else {
		fmt.Println("Nothing to do if this is a new blade")
	}

	// Kea may must be restarted after the above changes are posted
	deleteK8sPod()

	return nil
}

type Payload struct {
	Description string   `json:"Description"`
	MACAddress  string   `json:"MACAddress"`
	IPAddresses []string `json:"IPAddresses"`
	ComponentID string   `json:"ComponentID"`
}

// enableHmsDiscoveryJob enables the hms-discovery cron job.
func enableHmsDiscoveryJob(xname string) {
	rediscoverChassisSlot()
	complete, err := discoveryComplete(xname)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to determine if discovery is complete")
	}
	if !complete {
		log.Fatal().Msg("Discovery is not complete")
	}
	getHmsDiscoveryJob()
}

// rediscoverChassisSlot enables the chassis slot.
func rediscoverChassisSlot() {
	chassisbmc := StringPrompt("Enter the ChassisBMC Xname")
	// Rediscover the ChassisBMC
	// Rediscovering the ChassisBMC will update HSM to become aware of the newly populated slot and allow CAPMC to perform power actions on the slot.
	_, _, err := shell("cray", []string{"hsm", "inventory", "discover", "create", "--xnames", chassisbmc})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to inform HSM about the newly-populated slot.")
	}
}

func discoveryComplete(xname string) (bool, error) {
	stdout, _, err := shell("cray", []string{"hsm", "inventory", "redfishEndpoints", "describe", xname, "--format", "json"})
	if err != nil {
		return false, err
	}
	// Unmarshal response into a map for simple parsing without defining structs
	var re []map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &re)
	if err != nil {
		return false, err
	}

	for _, e := range re {
		discovery := e["DiscoveryInfo"].(map[string]interface{})
		switch discovery["LastDiscoveryStatus"].(string) {
		case "DiscoverOK":
			{
				fmt.Println("Discovery is complete")
				return true, nil
			}
		case "DiscoveryStarted":
			{
				fmt.Println("Discovery is still in progress")
				return false, nil
			}
		case "HTTPsGetFailed":
			{
				return false, errors.New("HTTPsGetFailed")
			}
		case "ChildVerificationFailed":
			{
				return false, errors.New("ChildVerificationFailed")
			}
		default:
			return false, errors.New("Unknown discovery status")
		}
	}
	return false, nil
}

// enableChassisSlot powers on the chassis slot.
func enableChassisSlot(xname string) error {
	_, _, err := shell("cray", []string{"hsm", "state", "components", "enabled", "update", "--enabled", "true", xname})
	if err != nil {
		return err
	}
	return nil
}

// powerOnChassisSlot powers on the chassis slot.
func powerOnChassisSlot(xname string) error {
	_, _, err := shell("cray", []string{"capmc", "xname_on", "create", "--xnames", xname, "--recursive", "true"})
	if err != nil {
		return err
	}
	// Wait for the chassis slot to power on
	for tick := range time.Tick(3 * time.Second) {

		// Prints UTC time and date
		fmt.Println("waiting 3 minutes for slot to power on", tick)
	}
	return nil
}

// clearRedfishEvents clears redfish events.
// this is a bug in HSM that needs to be fixed and this basically works around it.
func clearRedfishEvents() error {
	// stdout, _, err := shell("cray", []string{"hsm", "inventory", "redfishEndpoints", "list", "--type", "NodeBMC", "--format", "json"})
	stdout, _, err := shell("cat", []string{"testdata/fixtures/hsm/inventory_redfishEndpoints_list_nodebmc.json"})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to list redfish endpoints for node bmcs")
	}

	// Unmarshal response into a map for simple parsing without defining structs
	var reBmc map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &reBmc)
	if err != nil {
		return err
	}

	// var ei map[string]interface{}
	for _, re := range reBmc["RedfishEndpoints"].([]interface{}) {
		bmc := re.(map[string]interface{})["ID"].(string)
		nodeRaw := x.FromString(bmc)
		node, _ := nodeRaw.(x.NodeBMC)
		slot := node.ComputeModule
		// _, _, err := shell("/usr/share/doc/csm/scripts/operations/node_management/delete_bmc_subscriptions.py", []string{bmc})
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("Failed to list redfish endpoints for node bmcs")
		// }
		fmt.Printf("Clearing redfish events for %s in slot %d\n", bmc, slot)
	}
	return nil
}

// enableNodeInHsm enables the node in hsm.
func enableNodeInHsm() {
	// code to enable node in hsm
	// ...
}

// checkBosTemplate checks the bos template.
func checkBosTemplate() {
	// code to check bos template
	// ...
}

// addNodeToBos adds the node to bos.
func addNodeToBos() {
	// code to add node to bos
	// ...
}

// createBosSession creates the bos session.
func createBosSession() {
	// code to create bos session
	// ...
}

// bootNode boots the node.
func bootNode() {
	// code to boot node
	// ...
}

// checkFirmware checks the firmware.
func checkFirmware() {
	// code to check firmware
	// ...
}

// checkDvs checks the dvs.
func checkDvs() {
	// code to check dvs
	// ...
}

// checkCpsPods checks the cps pods.
func checkCpsPods() {
	// code to check cps pods
	// ...
}

// fixDvsMergeOne fixes a dvs merge one error
func fixDvsMergeOne() {
	// code to fix dvs merge one
	// ...
}

// getEthernetInterface gets interfaces from hsm
func getEthernetInterface(ip string) {
	// code to get ethernet interface
	// ...
}

// updateEthernetInterface updates interfaces in hsm
func updateEthernetInterface() {
	// code to update ethernet interface
	// ...
}

// checkFabricStatus checks the fabric status.
func checkFabricStatus() {
	// code to check fabric status
	// ...
}

// dedupeIps dedupes ips.
func dedupeIps() {
	// code to dedupe ips
	// ...
}

// hostnameHasOneIp checks if hostname has one ip.
func hostnameHasOneIp() bool {
	// code to check hostname has one ip
	// ...
	return true
}

// reloadKea reloads kea.
func reloadKea() {
	// code to reload kea
	// ...
}

// checkForDhcpLeases checks for dhcp leases.
func checkForDhcpLeases() {
	// code to check for dhcp leases
	// ...
}

// validateDns validates dns.
func validateDns() {
	// code to validate dns
	// ...
}

// validateSsh validates ssh.
func validateSsh() {
	// code to validate ssh
	// ...
}

// addBlade adds blades to the inventory.
func addBlade(args []string) {
	if addBladeStatus {
		checkAddBladeStatus()
		return
	}

	getHmsDiscoveryJob()
	suspendHmsDiscoveryJob()
	for _, xname := range xnames {
		fmt.Println("checking", xname)
		// check if the slot is populated
		populated, err := isChassisSlotPopulated(xname)
		if err != nil {
			log.Error().Err(err).Msg(err.Error())
			os.Exit(1)
		}
		// if the slot is populated, check if it is off
		if populated {
			off, err := isChassisSlotOff(xname)
			if err != nil {
				log.Error().Err(err).Msg(err.Error())
				os.Exit(1)
			}
			// if the slot is on, power it off
			if !off {
				powerOffChassisSlot(xname)
			}
			dvsOnHsn, err := isDvsOnHsn(xname)
			if err != nil {
				log.Error().Err(err).Msg(err.Error())
				os.Exit(1)
			}
			// If DVS is not on HSN, additional steps are needed to preserve xname to ip mapping
			if !dvsOnHsn {
				fmt.Println("DVS is not on HSN, proceeding with additional steps to preserve xname to ip mapping")
				err := preserveXnameToIpMapping()
				if err != nil {
					log.Error().Err(err).Msg(err.Error())
					os.Exit(1)
				}
			}
			dcomplete, err := discoveryComplete(xname)
			if err != nil {
				log.Error().Err(err).Msg(err.Error())
				os.Exit(1)
			}
			if !dcomplete {
				fmt.Println("Discovery is not complete, proceeding with additional steps to preserve xname to ip mapping")
				err := preserveXnameToIpMapping()
				if err != nil {
					log.Error().Err(err).Msg(err.Error())
					os.Exit(1)
				}
			}
			rediscoverChassisSlot()
			enableHmsDiscoveryJob(xname)
			enableChassisSlot(xname)
			powerOnChassisSlot(xname)
			rediscoverChassisSlot()
		}
	}
}

// YesNoPrompt asks yes/no questions using the label.
func YesNoPrompt(label string, def bool) bool {
	choices := "Y/n"
	if !def {
		choices = "y/N"
	}

	r := bufio.NewReader(os.Stdin)
	var s string

	for {
		fmt.Fprintf(os.Stderr, "%s (%s) ", label, choices)
		s, _ = r.ReadString('\n')
		s = strings.TrimSpace(s)
		if s == "" {
			return def
		}
		s = strings.ToLower(s)
		if s == "y" || s == "yes" {
			return true
		}
		if s == "n" || s == "no" {
			return false
		}
	}
}

// StringPrompt asks for a string value using the label
func StringPrompt(label string) string {
	var s string
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprint(os.Stderr, label+" ")
		s, _ = r.ReadString('\n')
		if s != "" {
			break
		}
	}
	return strings.TrimSpace(s)
}
