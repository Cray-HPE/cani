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
package blade

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"internal/shell"

	hsm_client "github.com/Cray-HPE/cani/pkg/hsm-client"
	X "github.com/Cray-HPE/cani/pkg/xname"
	"github.com/Cray-HPE/hms-xname/xnames"
	"github.com/antihax/optional"
	"github.com/rs/zerolog/log"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var (
	clientset *kubernetes.Clientset

	// getHmsDiscoveryJobName = "hms-discovery"
	getHmsDiscoveryJobName = "hms-discovery"
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
	xnameSlice             []string
	addBladeReqs           = []error{
		// prereqCrayCli(),
		// prereqDVS(),
		// prereqExistingCabinet(),
		// prereqSlingshotFabric(),
		// prereqSLS(),
		// prereqHSNLinkStatus(),
	}
)

// addBlade adds blades to the inventory.
//  1. check prerequisites
//  2. suspend HSM discovery job
//  3. for each blade to add:
//     a. determine if the destination chassis slot is populated (xXcCsS)
//     b. if the slot is populated, check if it is off (xXcCsS)
//     c. if the slot is on, power it off (xXcCsS)
//     d. if DVS is not on HSN, additional steps are needed to preserve xname to ip mapping
//     i.   get Node Maintenance Network ethernetInterfaces
//     ii.  delete the cray-dhcp-kea pod
//     iii. delete the ethernetInterfaces from HSM inventory
//     iv.  if a blade was already in the slot, POST the new ethernetInterfaces to HSM inventory
//     v.   restart the cray-dhcp-kea pod
//  4. rediscover the chassisbmc (xXcCbB)
//  5. verify that the chassisbmc is rediscovered (xXcCbB)
//  6. unsuspend the HSM discovery cron job
//  7. enable the chassis slot (xXcCbB)
//  8. power on the chassis slot (xXcCsS)
//  9. clear redfish events from nodeBMCs (xXcCsSbB)
//  10. enable nodes in the HSM database (xXcCsSbBnN)

// csmAddBlade takes an xname and adds it to the inventory.
func csmAddBlade(xname string) error {
	// a. determine if the destination chassis slot is populated
	populated, err := isChassisSlotPopulated(xname)
	if err != nil {
		return err
	}
	// b. if the slot is populated, check if it is off
	if populated {
		log.Info().Msgf("[%s] Blade slot is populated", xname)
		on, err := isChassisSlotOn(xname)
		if err != nil {
			return err
		}
		// c. if the slot is on, power it off
		if !on {
			log.Info().Msgf("[%s] Blade slot is on, powering it off", xname)
			err := powerOffChassisSlot(xname)
			if err != nil {
				return err
			}
		}

		fmt.Println("Install the blade")

		// To proceed, we must know if DVS is on the HSN or the NMN
		xname = xname + "b0n0" // FIXME: Needs node xname
		dvsOnHsn, err := isDvsOnHsn(xname)
		if err != nil {
			return err
		}

		// d. if DVS is not on HSN, additional steps are needed to preserve xname to ip mapping
		if !dvsOnHsn {
			log.Info().Msgf("[%s] DVS is not on HSN, proceeding with additional steps to preserve xname to IP mapping", xname)
			// //			i.   get Node Maintenance Network ethernetInterfaces
			// //	    ii.  delete the cray-dhcp-kea pod
			// //			iii. delete the ethernetInterfaces from HSM inventory
			// //      iv.  if a blade was already in the slot, POST the new ethernetInterfaces to HSM inventory
			// //			v.   restart the cray-dhcp-kea pod
			// err := preserveXnameToIpMapping(xn)
			// if err != nil {
			// 	return err
			// }
			// log.Info().Msgf("[%s] Preserved xn to IP mapping", xn)
		} else {
			log.Info().Msgf("[%s] DVS is on HSN. Nothing to do for IP mapping", xname)
		}

		// 4. rediscover the chassisbmc
		dcomplete, err := discoveryComplete(xname)
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Rediscovered chassisBMC", xname)

		// 5. verify that the chassisbmc is rediscovered
		if !dcomplete {
			log.Info().Msgf("[%s] Discovery is not complete, proceeding with additional steps to preserve xn to ip mapping", xname)
			return errors.New(fmt.Sprintf("[%s] Discovery is not complete, please try again later", xname))
		}

		err = rediscoverChassisSlot(xname)
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Blade slot rediscovered", xname)

		// 6. unsuspend the HSM discovery cron job
		err = enableHsmDiscoveryJob()
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] HSM discovery job enabled", xname)

		// 7. enable the chassis slot
		err = enableChassisSlot(xname)
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Blade slot enabled", xname)

		// 8. power on the chassis slot
		err = powerOnChassisSlot(xname)
		if err != nil {
			return err
		}
		err = waitUntilChassisSlotOn(xname)
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Blade slot powered on", xname)

		// 9. clear redfish events
		err = clearRedfishEvents()
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Redfish events cleared", xname)

		// 10. enable nodes in the HSM database
		err = enableNodeInHsm(xname)
		if err != nil {
			return err
		}
		log.Info().Msgf("[%s] Node enabled in HSM", xname)
	}

	return nil
}

func prereqCrayCli() error {
	log.Info().Msg("cray CLI is initialized")
	return nil
}

func prereqDVS() error {
	log.Info().Msg("DVS is on NMN or HMN")
	return nil
}

func prereqExistingCabinet() error {
	log.Info().Msg("Use existing cabinet")
	return nil
}

func prereqSlingshotFabric() error {
	log.Info().Msg("Slingshot fabric is configured with the desired topology")
	return nil
}

func prereqSLS() error {
	log.Info().Msg("SLS has the desired HSN configuration")
	return nil
}

func prereqHSNLinkStatus() error {
	log.Info().Msg("Check HSN and link status")
	return nil
}

// checkPrerequisites runs all prerequisite functions and errors if any fail.
func checkPrerequisites() error {
	var fail = false
	// Run each prerequisite function and collect errors
	for _, preq := range addBladeReqs {
		err := preq
		if err != nil {
			fmt.Printf("%s %+v\n", "NO", err)
			fail = true
			continue
		}
	}
	if fail {
		return errors.New("Prerequisites failed")
	}
	return nil
}

// suspendHsmDiscoveryJob suspends the hms-discovery cron job to disable it.
func suspendHsmDiscoveryJob() error {
	// patch the hms-discovery cron job to suspend it
	stdout, stderr, err := shell.Shell("kubectl", []string{"-n", "services", "patch", "cronjobs", getHmsDiscoveryJobName, "-p", "{\"spec\" : {\"suspend\" : true }}"})
	if err != nil {
		if debug {
			log.Debug().Msg(string(stderr))
		}
		return errors.New(fmt.Sprint("Failed to suspend hms-discovery cron job.\n", string(stderr)))
	}
	if debug {
		log.Debug().Msg(string(stdout))
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
	return nil
}

func getHsmDiscoveryJob() error {
	// get the hms-discovery cron job
	stdout, stderr, err := shell.Shell("kubectl", []string{"-n", "services", "get", "cronjobs", getHmsDiscoveryJobName})
	if err != nil {
		if debug {
			log.Debug().Msg(string(stderr))
		}
		return errors.New(fmt.Sprint("Failed to get hms-discovery cron job.", string(stderr)))
	}
	if debug {
		log.Debug().Msg(string(stdout))
	}
	// cronjob, err := clientset.BatchV1beta1().CronJobs(hmsNamespace).Get(
	// 	context.Background(),
	// 	getHmsDiscoveryJobName,
	// 	v1.GetOptions{})
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to get hms-discovery cron job.")
	// }
	return nil
}

// isChassisSlotPopulated determines if the destination chassis slot is populated.
func isChassisSlotPopulated(xname string) (bool, error) {
	// Get the statecomponents from HSM
	sc, resp, err := HSM.ComponentApi.DoComponentGet(context.Background(), xname)
	if debug {
		log.Debug().Msgf(fmt.Sprintf("StateComponent for %s %+v", xname, sc))
	}
	if err != nil {
		return false, errors.New(
			fmt.Sprintf(
				"Failed to get hsm state for destination chassis slot.\n%s %+v",
				err.Error(),
				resp))
	}

	// If the state is On or Off, return the slot is populated
	if *sc.State == hsm_client.ON_HmsState100 || *sc.State == hsm_client.OFF_HmsState100 {
		return true, nil
	}
	// If the state is Empty, return the slot is not populated
	if *sc.State == hsm_client.EMPTY_HmsState100 {
		return false, nil
	}

	return false, errors.New("Unknown state")
}

// isChassisSlotOff determines if the destination chassis slot is off.
func isChassisSlotOn(xname string) (bool, error) {
	stdout, stderr, err := shell.Shell("cray", []string{"capmc", "get_xname_status", "create", "--xnames", xname, "--format", "json"})
	if err != nil {
		if debug {
			log.Debug().Msgf(string(stderr))
		}
		return false, errors.New(fmt.Sprint("Failed to get xname status for destination chassis slot.", string(stderr), err.Error()))
	}
	if debug {
		log.Debug().Msgf(string(stdout))
	}
	// Unmarshal response into a map for simple parsing without defining structs
	var capInfo map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &capInfo)
	if err != nil {
		return false, err
	}

	if capInfo["e"].(float64) != 0 ||
		capInfo["err_msg"].(string) != "Success" ||
		capInfo["on"].([]string) != nil {
		return true, nil
	}

	if capInfo["off"].([]string) != nil {
		return false, nil
	}

	if debug {
		log.Debug().Msgf(fmt.Sprintf("%+v", capInfo))
	}
	return true, nil
}

func powerOffChassisSlot(xname string) error {
	// code to power off chassis slot
	stdout, stderr, err := shell.Shell("cray", []string{"capmc", "xname_off", "create", "--xnames", xname, "--recursive", "true"})
	if err != nil {
		if debug {
			log.Debug().Msgf(string(stderr))
		}
		return errors.New(fmt.Sprint("Failed to power off chassis slot.", string(stderr), err.Error()))
	}
	if debug {
		log.Debug().Msgf(string(stdout))
	}
	return nil
}

func powerOnChassisSlot(xname string) error {
	// code to power off chassis slot
	stdout, stderr, err := shell.Shell("cray", []string{"capmc", "xname_on", "create", "--xnames", xname, "--recursive", "true"})
	if err != nil {
		if debug {
			log.Debug().Msgf(string(stderr))
		}
		return errors.New(fmt.Sprint("Failed to power off chassis slot.", string(stderr), err.Error()))
	}
	if debug {
		log.Debug().Msgf(string(stdout))
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
	x := X.Xname(xname)
	// check a compute node's boot parameters
	stdout, stderr, err := shell.Shell("cray", []string{"bss", "bootparameters", "list", "--hosts", string(x), "--format", "json"})
	if err != nil {
		return false, errors.New(fmt.Sprintf("Could not determine if DVS is operating over the NMN or HMN\n%s", stderr))
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
	// // When DVS is operating over the NMN and a blade is being replaced, the mapping of node component name (xname) to node IP address must be preserved. Kea automatically adds entries to the HSM ethernetInterfaces table when DHCP lease is provided (about every 5 minutes)
	// _, stderr, err = shell.Shell("lnetctl", []string{"net", "show"})
	// if err != nil {
	// 	return false, errors.New(fmt.Sprintf("Failed to run 'lnetctl net show'\n%s", stderr))
	// }

	return false, nil
}

func getNodeMaintenanceEthernetInterfaces(xname string) ([]hsm_client.CompEthInterface100, error) {
	// stdout, _, err := shell.Shell("cat", []string{"testdata/fixtures/hsm/inventory_ethernetInterfaces.json"})
	// Get the ethernet interfaces from HSM
	eis, resp, err := HSM.ComponentEthernetInterfacesApi.DoCompEthInterfacesGetV2(context.Background(), nil)
	if err != nil {
		return []hsm_client.CompEthInterface100{}, errors.New(
			fmt.Sprintf("Failed to get node maintenance ethernet interfaces. %s. %+v",
				err.Error(),
				resp))
	}

	if debug {
		log.Debug().Msgf(fmt.Sprintf("ethernetInterfaces %+v", eis))
	}

	// Filter out the node maintenance network
	for _, ei := range eis {
		if ei.Description == "Node Maintenance Network" {
			log.Info().Msg(fmt.Sprintf("%s %s", ei.Description, ei.ID))
			eis = append(eis, ei)
		}
	}

	return eis, nil
}

func deleteK8sPod() error {
	pods, err := clientset.CoreV1().Pods("services").List(context.TODO(), v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cray-dhcp-kea",
	})
	if err != nil {
		return err
	}
	if debug {
		log.Debug().Msgf(fmt.Sprintf("Pods %+v", pods))
	}
	for _, i := range pods.Items {
		if debug {
			log.Debug().Msgf(fmt.Sprintf("%+v", i))
		}
		err = clientset.CoreV1().Pods(hmsNamespace).Delete(context.TODO(), i.Name, v1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}

// func getMac(xn string) (mac string, err error) {
// 	xname := xnames.FromString(xn)
// 	ei, err := HSM.GetEthernetInterfaces(context.Background(), xname.ParentInterface().String())
// 	if err != nil {
// 		return "", err
// 	}
// 	if debug {
// 		log.Debug().Msgf(fmt.Sprintf("EthernetInterfaces %+v", ei))
// 	}
// 	for _, iface := range ei {
// 		if iface.ComponentID == xname.ParentInterface().String() {
// 			mac = iface.MACAddress
// 			if debug {
// 				log.Debug().Msgf(fmt.Sprintf("MAC %+v", mac))
// 			}
// 		}
// 	}
// 	return mac, nil
// }

// func getComponentId(xn string) (cid string, err error) {
// 	xname := xnames.FromString(xn)
// 	ei, err := HSM.GetEthernetInterfaces(context.Background(), xname.ParentInterface().String())
// 	if err != nil {
// 		return "", err
// 	}
// 	if debug {
// 		log.Debug().Msgf(fmt.Sprintf("EthernetInterfaces %+v", ei))
// 	}
// 	for _, iface := range ei {
// 		if iface.ComponentID == xname.ParentInterface().String() {
// 			cid = iface.ComponentID
// 			if debug {
// 				log.Debug().Msgf(fmt.Sprintf("ComponentID %+v", cid))
// 			}
// 		}
// 	}
// 	return cid, nil
// }

// func getIp(xn string) (ips hsm_client.EthernetInterfaces, err error) {
// 	xname := xnames.FromString(xn)
// 	ei, err := HSM.GetEthernetInterfaces(context.Background(), xname.ParentInterface().String())
// 	if err != nil {
// 		return hsm_client.EthernetInterfaces{}, err
// 	}
// 	if debug {
// 		log.Debug().Msgf(fmt.Sprintf("EthernetInterfaces %+v", ei))
// 	}
// 	for _, iface := range ei {
// 		if iface.ComponentID == xname.ParentInterface().String() {
// 			ips = append(ips, iface)
// 			if debug {
// 				log.Debug().Msgf(fmt.Sprintf("IPs %+v", ips))
// 			}
// 		}
// 	}
// 	return ips, nil
// }

// // preserveXnameToIpMapping preserves the mapping of node component name (xname) to node IP address.
// func preserveXnameToIpMapping(xn string) error {
// 	replacement, err := isChassisSlotPopulated(xn)
// 	if err != nil {
// 		return err
// 	}
// 	if replacement {
// 		// i.   get Node Maintenance Network ethernetInterfaces
// 		eis, err := getNodeMaintenanceEthernetInterfaces(xn)
// 		if err != nil {
// 			return errors.New("Could not get node maintenance ethernet interfaces")
// 		}

// 		if len(eis) > 0 {
// 			log.Info().Msg("Since the blade was previously populated, existing ethernet interfaces will be deleted")
// 			for _, ei := range eis {
// 				// iii. delete the ethernetInterfaces from HSM inventory
// 				_, stderr, err := shell.Shell("cray", []string{"hsm", "inventory", "ethernetInterfaces", "delete", ei.ID})
// 				if err != nil {
// 					return errors.New(fmt.Sprint("Failed to delete ethernet interface.", string(stderr)))
// 				}
// 			}

// 			log.Info().Msg("Adding new ethernet interfaces")
// 			// ii.  delete the cray-dhcp-kea pod
// 			err := deleteK8sPod()
// 			if err != nil {
// 				return errors.New("Failed to delete cray-dhcp-kea pod")
// 			}
// 			mac, err := getMac(xn)
// 			if err != nil {
// 				return err
// 			}
// 			ips, err := getIp(xn)
// 			if err != nil {
// 				return err
// 			}
// 			cid, err := getComponentId(xn)
// 			if err != nil {
// 				return err
// 			}
// 			// curl -H "Authorization: Bearer ${TOKEN}" -L -X POST 'https://api-gw-service-nmn.local/apis/smd/hsm/v2/Inventory/EthernetInterfaces' -H 'Content-Type: application/json' --data-raw "{
// 			//     \"Description\": \"Node Maintenance Network\",
// 			//     \"MACAddress\": \"$MAC\",
// 			//     \"IPAddresses\": [
// 			//         {
// 			//             \"IPAddress\": \"$IP_ADDRESS\"
// 			//         }
// 			//     ],
// 			//     \"ComponentID\": \"$XNAME\"
// 			// }"
// 			data := hsm_client.CompEthInterface100{
// 				Description: "Node Maintenance Network",
// 				MACAddress:  mac,
// 				ComponentID: cid,
// 			}
// 			if len(ips) != 0 {
// 				ip := hsm_client.CompEthInterface100IpAddressMapping{
// 					IPAddress: "127.127.127.127"}
// 				data.IPAddresses = append(data.IPAddresses, ip)
// 			} else {
// 				for i, ip := range ips {
// 					data.IPAddresses[i] = ip.IPAddresses[i]
// 				}
// 			}
// 			// payloadBytes, err := json.Marshal(data)
// 			// if err != nil {
// 			// 	log.Fatal().Err(err)
// 			// }
// 			// body := bytes.NewReader(payloadBytes)

// 			// iv. if a blade was already in the slot, POST the new ethernetInterfaces to HSM inventory
// 			// req, err := http.NewRequest("POST", "https://api-gw-service-nmn.local/apis/smd/hsm/v2/Inventory/EthernetInterfaces", body)
// 			// if err != nil {
// 			// 	log.Fatal().Err(err)
// 			// }
// 			// req.Header.Set("Authorization", os.ExpandEnv("Bearer ${TOKEN}"))
// 			// req.Header.Set("Content-Type", "application/json")

// 			// resp, err := http.DefaultClient.Do(req)
// 			// if err != nil {
// 			// 	log.Fatal().Err(err)
// 			// }
// 			// defer resp.Body.Close()
// 			// v.   restart the cray-dhcp-kea pod
// 			err = deleteK8sPod()
// 			if err != nil {
// 				return errors.New("Failed to delete cray-dhcp-kea pod")
// 			}
// 		} else {
// 			log.Info().Msg("No ethernet interfaces found")
// 		}

// 	} else {
// 		log.Info().Msg("Nothing to do if this is a new blade")
// 	}

// 	return nil
// }

// enableHsmDiscoveryJob enables the hms-discovery cron job.
func enableHsmDiscoveryJob() error {
	// patch the hms-discovery cron job to suspend it
	_, stderr, err := shell.Shell("kubectl", []string{"-n", "services", "patch", "cronjobs", getHmsDiscoveryJobName, "-p", "{\"spec\" : {\"suspend\" : false }}"})
	if err != nil {
		return errors.New(fmt.Sprint("Failed to enable hms-discovery cron job.", string(stderr)))
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
	err = getHsmDiscoveryJob()
	if err != nil {
		return err
	}
	return nil
	// rediscoverChassisSlot()
	// complete, err := discoveryComplete(xname)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("Failed to determine if discovery is complete")
	// }
	// if !complete {
	// 	log.Fatal().Msg("Discovery is not complete")
	// }
	// return nil
}

// rediscoverChassisSlot enables the chassis slot.
func rediscoverChassisSlot(xname string) error {
	// the parent may not always be the bmc, but for development purposes, this works
	xn := xnames.FromString(xname)
	xname = xn.ParentInterface().String()
	// Rediscovering the ChassisBMC will update HSM to become aware of the newly populated slot and allow CAPMC to perform power actions on the slot.
	// Wrap the body in an interface using optional.NewInterface()
	body := optional.NewInterface(hsm_client.Discover100DiscoverInput{
		Xnames: []string{xname},
	})

	// Set the Body field of the options struct to the wrapped body
	opts := &hsm_client.DiscoverApiDoInventoryDiscoverPostOpts{
		Body: body,
	}
	// _, _, err := shell.Shell("cray", []string{"hsm", "inventory", "discover", "create", "--xnames", xname})
	_, resp, err := HSM.DiscoverApi.DoInventoryDiscoverPost(context.Background(), opts)
	if err != nil {
		return errors.New(fmt.Sprintf(
			"Failed to inform HSM about the newly-populated slot. %s %+v",
			err.Error(),
			resp.Status))
	}
	log.Info().Msgf("[%s] HSM discovery status: ", xname)
	return nil
}

func discoveryComplete(xname string) (bool, error) {
	// the parent may not always be the bmc, but for development purposes, this works
	xn := xnames.FromString(xname)
	xname = xn.ParentInterface().String()

	for {
		rfe, resp, err := HSM.RedfishEndpointApi.DoRedfishEndpointGet(context.Background(), xname)
		if err != nil {
			return false, errors.New(fmt.Sprintf("Failed to get RedfishEndpoints to check discovery status: %s. %+v",
				err.Error(),
				resp.Status))
		}
		log.Info().Msg("Discovery status: " + rfe.DiscoveryInfo.LastDiscoveryStatus)
		switch rfe.DiscoveryInfo.LastDiscoveryStatus {
		case "DiscoverOK":
			{
				log.Info().Msgf("[%s] Discovery is complete", xname)
				return true, nil
			}
		case "DiscoveryStarted":
			{
				// Wait for 5 seconds before checking again
				time.Sleep(5 * time.Second)
			}
		case "HTTPsGetFailed":
			{
				return false, errors.New("Discovery HTTPsGetFailed")
			}
		case "ChildVerificationFailed":
			{
				return false, errors.New("Discovery ChildVerificationFailed")
			}
		default:
			return false, errors.New(fmt.Sprintf("Unknown discovery status: %+v", rfe.DiscoveryInfo))
		}
	}
}

// enableChassisSlot powers on the chassis slot.
func enableChassisSlot(xname string) error {
	// _, _, err := shell.Shell("cray", []string{"hsm", "state", "components", "enabled", "update", "--enabled", "true", xname})
	body := hsm_client.Component100PatchEnabled{
		Enabled: true,
	}
	resp, err := HSM.ComponentApi.DoCompEnabledPatch(context.Background(), body, xname)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to enable chassis slot: %s. %+v", err.Error(), resp.Status))
	}
	return nil
}

// waitUntilChassisSlotOn powers on the chassis slot.
func waitUntilChassisSlotOn(xname string) error {
	// Wait for the chassis slot to power on
	for tick := range time.Tick(3 * time.Second) {
		// Prints UTC time and date
		log.Info().Msgf("waiting 3 minutes for slot to power on, %v", tick)
		on, err := isChassisSlotOn(xname)
		if err != nil {
			return err
		}
		if on {
			break
		}
	}
	return nil
}

// clearRedfishEvents clears redfish events.
// this is a bug in HSM that needs to be fixed and this basically works around it.
func clearRedfishEvents() error {
	// stdout, _, err := shell.Shell("cray", []string{"hsm", "inventory", "redfishEndpoints", "list", "--type", "NodeBMC", "--format", "json"})
	// stdout, _, err := shell.Shell("cat", []string{"testdata/fixtures/hsm/inventory_redfishEndpoints_list_nodebmc.json"})
	rfes, resp, err := HSM.RedfishEndpointApi.DoRedfishEndpointsGet(context.Background(), &hsm_client.RedfishEndpointApiDoRedfishEndpointsGetOpts{
		Type_: optional.NewString("NodeBMC"),
	})
	if err != nil {
		log.Fatal().Err(err).Msg(fmt.Sprintf("Failed to list redfish endpoints for node bmcs: %s", resp.Status))
	}

	// var ei map[string]interface{}
	for _, rfe := range rfes.RedfishEndpoints {
		bmc := rfe.ID
		nodeRaw := xnames.FromString(bmc)
		node, _ := nodeRaw.(xnames.NodeBMC)
		slot := node.ComputeModule
		// _, _, err := shell.Shell("/usr/share/doc/csm/scripts/operations/node_management/delete_bmc_subscriptions.py", []string{bmc})
		// if err != nil {
		// 	log.Fatal().Err(err).Msg("Failed to list redfish endpoints for node bmcs")
		// }
		log.Info().Msgf("[%s] Clearing redfish events in slot %d", bmc, slot)
	}
	return nil
}

// enableNodeInHsm enables the node in hsm.
func enableNodeInHsm(xname string) error {
	// enable the node in HSM
	// _, _, err := shell.Shell("cray", []string{"hsm", "state", "components", "bulkEnabled", "update", "--enabled", "true", "--component-ids", xname})
	resp, err := HSM.ComponentApi.DoCompBulkEnabledPatch(context.Background(), hsm_client.ComponentArrayPatchArrayEnabled{
		ComponentIDs: []string{xname},
	})
	if err != nil {
		return errors.New(
			fmt.Sprintf("Failed to enable node in HSM: %s. %+v",
				err.Error(),
				resp.Status))
	}
	// validate it is enabled
	// _, _, err = shell.Shell("cray", []string{"hsm", "state", "components", "query", "create", "--component-ids", xname})
	sc, resp, err := HSM.ComponentApi.DoComponentQueryGet(context.Background(), xname, nil)
	if err != nil {
		return errors.New(
			fmt.Sprintf("Failed to check if node is enabled in HSM: %s. %+v",
				err.Error(),
				resp.Status))
	}
	for _, component := range sc.Components {
		log.Info().Msgf("[%s] Node is enabled: %v", xname, component.Enabled)
	}
	return nil
}

// checkBosTemplate checks the bos template.
func checkBosTemplate(xname string) (string, error) {
	log.Info().Msg("Determine how the BOS Session template references compute hosts.")
	tname := StringPrompt("Enter the name of the BOS_TEMPLATE:")
	// stdout, _, err := shell.Shell("cat", []string{"testdata/fixtures/bos/sessiontemplate_describe.json"})
	stdout, _, err := shell.Shell("cray", []string{"bos", "sessiontemplates", "describe", tname, "--format", "json"})
	if err != nil {
		return tname, err
	}
	// Unmarshal response into a map for simple parsing without defining structs
	var templateDescribed map[string]interface{}
	err = json.Unmarshal([]byte(stdout), &templateDescribed)
	if err != nil {
		return tname, err
	}
	bs := templateDescribed["boot_sets"].(map[string]interface{})
	nl := bs["node_list"].(interface{})
	if nl == nil {
		log.Info().Msg("Boot sets not populated, skipping to booting nodes")
		return tname, nil
	}
	// log.Info().Msg("one or more boot sets within the BOS Session template reference nodes explicitly by xname")
	// addNodeToBos(templateDescribed, xname)
	return tname, nil
}

// addNodeToBos adds the node to bos.
func addNodeToBos(template map[string]interface{}, xname string) error {
	tname := StringPrompt("Enter the name of the BOS_TEMPLATE to add nodes to:")
	ttemplate := "/tmp/bostemplate.json"
	b, err := json.Marshal(template)
	if err != nil {
		return err
	}

	// open output file
	fo, err := os.Create(ttemplate)
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer fo.Close()

	// write the file
	if _, err := fo.Write(b); err != nil {
		return err
	}
	// stdout, _, err := shell.Shell("cat", []string{"testdata/fixtures/bos/sessiontemplate_describe.json"})
	_, _, err = shell.Shell("cray", []string{"bos", "sessiontemplate", "create", "--file", ttemplate, "--name", tname})
	if err != nil {
		return err
	}
	err = bootNode(tname, xname)
	if err != nil {
		return err
	}
	return nil
}

// bootNode boots the node.
func bootNode(tname string, xname string) error {
	_, _, err := shell.Shell("cray", []string{"bos", "session", "create", "--template-uuid", tname, "--operation", "reboot", "--limit", xname})
	if err != nil {
		return err
	}
	return nil
}

// checkFirmware checks the firmware.
func checkFirmware() {
	log.Info().Msg("Verify that the correct firmware versions are present for node BIOS, node controller (nC), NIC mezzanine card (NMC), GPUs, and so on")
	complete := YesNoPrompt("Are the versions correct?", false)
	if !complete {
		log.Error().Msg("Please update the firmware and try again")
		os.Exit(1)
	}
}

// checkCpsPods checks the cps pods.
func checkCpsPods() {
	// app.kubernetes.io/instance=cray-cps
	// app.kubernetes.io/name=cray-cps
	// etcd_cluster=cray-cps-etcd
	pods, err := clientset.CoreV1().Pods("services").List(context.TODO(), v1.ListOptions{
		LabelSelector: "app.kubernetes.io/name=cray-cps",
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
	// cray-cps pods should be running
}

// fixDvsMergeOne fixes a dvs merge one error
func fixDvsMergeOne(xname string) error {
	fmt.Println("SSH to each worker node running CPS/DVS, and run ensure that there are no recurring DVS merge errors shown")
	fmt.Println("Run 'dmesg -T | grep \"DVS: merge_one\"' on each worker")
	errPresent := YesNoPrompt("Were any of the merge_one errors present?", false)
	if errPresent {
		fmt.Println("The IP address of the node needs to be corrected. This will prevent the need to reload DVS")
		currentIP := StringPrompt("Enter the current IP address of the node:")
		desiredIP := StringPrompt("Enter the new IP address of the node:")

		// Determine the HSM EthernetInterface entry holding onto the desired IP address.
		// stdout, _, err := shell.Shell("cray", []string{"hsm", "inventory", "ethernetInterfaces", "list", "--ip-address", desiredIP, "--output", "json"})
		stdout, _, err := shell.Shell("cat", []string{"testdata/fixtures/hsm/inventory_redfishEndpoints_list_ip_address.json"})
		if err != nil {
			return err
		}
		// Unmarshal response into a map for simple parsing without defining structs
		var desiredIpInUse []map[string]interface{}
		err = json.Unmarshal([]byte(stdout), &desiredIpInUse)
		if err != nil {
			return err
		}
		for _, i := range desiredIpInUse {
			ips := i["IPAddresses"].([]interface{})
			for _, j := range ips {
				ip := j.(map[string]interface{})["IPAddress"].(string)
				if desiredIP == ip {
					id := i["ID"].(string)
					log.Info().Msg("IP address is already in use and must be removed from HSM")
					_, _, err := shell.Shell("cat", []string{"hsm", "inventory", "ethernetInterfaces", "delete", id})
					if err != nil {
						return err
					}

				}

			}
		}

		// Determine the HSM EthernetInterface entry holding onto the desired IP address.
		// stdout, _, err = shell.Shell("cray", []string{"hsm", "inventory", "ethernetInterfaces", "list", "--component-id", xname, "--ip-address", currentIP, "--output", "json"})
		stdout, _, err = shell.Shell("cat", []string{"testdata/fixtures/hsm/inventory_redfishEndpoints_list_ip_address.json"})
		if err != nil {
			return err
		}
		// Unmarshal response into a map for simple parsing without defining structs
		var currentIpInUse []map[string]interface{}
		err = json.Unmarshal([]byte(stdout), &currentIpInUse)
		if err != nil {
			return err
		}
		for _, i := range currentIpInUse {
			ips := i["IPAddresses"].([]interface{})
			for _, j := range ips {
				ip := j.(map[string]interface{})["IPAddress"].(string)
				if currentIP == ip {
					currentID := i["ID"].(string)
					log.Info().Msg("IP address is already in use and must be removed from HSM")
					_, _, err := shell.Shell("cat", []string{"hsm", "inventory", "ethernetInterfaces", "update", currentID, "--component-id", xname, "--ip-addresses--ip-address", desiredIP})
					if err != nil {
						return err
					}
				}
			}
		}
		fmt.Println("Reboot the node")
		rebooted := YesNoPrompt("Is the node rebooted", false)
		if !rebooted {
			fmt.Println("Please reboot the node and try again")
			os.Exit(1)
		}
	}
	return nil
}

func checkDvs() {
	fmt.Println("SSH to the node and check each DVS mount.")
	fmt.Println("Run 'mount | grep dvs | head -1'")
	dvsok := YesNoPrompt("Is DVS ok on each node?", false)
	if !dvsok {
		fmt.Println("Please try again")
		os.Exit(1)
	}
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
func checkFabricStatus() error {
	// Determine the pod name for the Slingshot fabric manager pod and check the status of the fabric.
	// app.kubernetes.io/instance=slingshot-fabric-manager
	pods, err := clientset.CoreV1().Pods("services").List(context.TODO(), v1.ListOptions{
		LabelSelector: "app.kubernetes.io/instance=slingshot-fabric-manager",
	})
	if err != nil {
		log.Fatal().Err(err)
	}
	for _, p := range pods.Items {
		_, _, err := shell.Shell("kubectl", []string{"exec", "-n", "services", "-it", p.Name, "--", "fmn_status"})
		if err != nil {
			return err
		} else {
			break
		}
	}
	return nil
}

// dedupeIps dedupes ips.
func dedupeIps(xname string) error {
	_, _, err := shell.Shell("nslookup", []string{xname})
	if err != nil {
		return err
	}
	return nil
}

// hostnameHasOneIp checks if hostname has one ip.
func hostnameHasOneIp() bool {
	// code to check hostname has one ip
	// ...
	return true
}

// reloadKea reloads kea.
func reloadKea() error {
	// Generated by curl-to-Go: https://mholt.github.io/curl-to-go

	// curl -s -k -H "Authorization: Bearer ${TOKEN}" -X POST -H "Content-Type: application/json" -d '{ "command": "config-reload",  "service": [ "dhcp4" ] }' https://api-gw-service-nmn.local/apis/dhcp-kea

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{},
	}
	client := &http.Client{Transport: tr}

	// anonymous struct for json payload
	data := struct {
		Command string   `json:"command"`
		Service []string `json:"service"`
	}{
		Command: "config-reload",
		Service: []string{"dhcp4"},
	}

	payloadBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api-gw-service-nmn.local/apis/dhcp-kea", body)
	if err != nil {
		// handle err
	}
	req.Header.Set("Authorization", os.ExpandEnv("Bearer ${TOKEN}"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// checkForDhcpLeases checks for dhcp leases.
func checkForDhcpLeases() error {
	// curl -H "Authorization: Bearer ${TOKEN}" -X POST -H "Content-Type: application/json" -d '{ "command": "lease4-get-all", "service": [ "dhcp4" ] }' https://api-gw-service-nmn.local/apis/dhcp-kea
	data := struct {
		Command string   `json:"command"`
		Service []string `json:"service"`
	}{
		Command: "leases-reclaim",
		Service: []string{"dhcp4"},
	}
	payloadBytes, err := json.Marshal(data)
	if err != nil {
		// handle err
	}
	body := bytes.NewReader(payloadBytes)

	req, err := http.NewRequest("POST", "https://api-gw-service-nmn.local/apis/dhcp-kea", body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", os.ExpandEnv("Bearer ${TOKEN}"))
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// If there are no DHCP leases, then there is a configuration error.

	return nil
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
