/*
 * Hardware State Manager API
 *
 * The Hardware State Manager (HSM) inventories, monitors, and manages hardware, and tracks the logical and dynamic component states, such as roles, NIDs, and other basic metadata needed to provide most common administrative and operational functions. HSM is the single source of truth for the state of the system. It contains the component state and information on Redfish endpoints for communicating with components via Redfish. It also allows administrators to create partitions and groups for other uses. ## Resources ### /State/Components HMS components are created during inventory discovery and provide a higher-level representation of the component, including state, NID, role (i.e. compute/service), subtype, and so on. Unlike ComponentEndpoints, however, they are not strictly linked to the parent RedfishEndpoint, and are not automatically deleted when the RedfishEndpoints are (though they can be deleted via a separate call). This is because these components can also represent abstract components, such as removed components (e.g. which would remain, but have their states changed to \"Empty\" upon removal). ### /Defaults/NodeMaps  This resource allows a mapping file (NodeMaps) to be uploaded that maps node xnames to Node IDs, and optionally, to roles and subroles. These mappings are used when discovering nodes for the first time. These mappings should be uploaded prior to discovery and should contain mappings for each valid node xname in the system, whether populated or not. Nodemap is a JSON file that contains the xname of the node, node ID, and optionally role and subrole. Role can be Compute, Application, Storage, Management etc. The NodeMaps collection can be uploaded to HSM automatically at install time by specifying it as a JSON file. As a result, the endpoints are then automatically discovered by REDS, and inventory discovery is performed by HSM. The desired NID numbers will be set as soon as the nodes are created using the NodeMaps collection.  It is recommended that Nodemaps are uploaded at install time before discovery happens. If they are uploaded after discovery, then the node xnames need to be manually updated with the correct NIDs. You can update NIDs for individual components by using PATCH /State/Components/{xname}/NID.  ### /Inventory/Hardware  This resource shows the hardware inventory of the entire system and contains FRU information in location. All entries are displayed as a flat array. ### /Inventory/HardwareByFRU  Every component has FRU information. This resource shows the hardware inventory for all FRUs or for a specific FRU irrespective of the location. This information is constant regardless of where the hardware item is currently in the system. If a HWInventoryByLocation entry is currently populated with a piece of hardware, it will have the corresponding HWInventoryByFRU object embedded. This FRU info can also be looked up by FRU ID regardless of the current location. ### /Inventory/Hardware/Query/{xname}  This resource gets you information about a specific component and it's sub-components. The xname can be a component, partition, ALL, or s0. Both ALL and s0 represent the entire system. ### /Inventory/RedfishEndpoints  This is a BMC or other Redfish controller that has a Redfish entry point and Redfish service root. It is used to discover the components managed by this endpoint during discovery and handles all Redfish interactions by these subcomponents.  If the endpoint has been discovered, this entry will include the ComponentEndpoint entries for these managed subcomponents. You can also create a Redfish Endpoint or update the definition for a Redfish Endpoint. The xname identifies the location of all components in the system, including chassis, controllers, nodes, and so on. Redfish endpoints are given to State Manager. ### /Inventory/ComponentEndpoints  Component Endpoints are the specific URLs for each individual component that are under the Redfish endpoint. Component endpoints are discovered during inventory discovery. They are the management-plane representation of system components and are linked to the parent Redfish Endpoint. They provide a glue layer to bridge the higher-level representation of a component with how it is represented locally by Redfish.  The collection of ComponentEndpoints can be obtained in full, optionally filtered on certain criteria (e.g. obtain just Node components), or accessed by their xname IDs individually. ### /Inventory/ServiceEndpoints  ServiceEndpoints help you do things on Redfish like updating the firmware. They are discovered during inventory discovery. ### /groups  Groups are named sets of system components, most commonly nodes. A group groups components under an administratively chosen label (group name). Each component may belong to any number of groups. If a group has exclusiveGroup=<excl-label> set, then a node may only be a member of one group that matches that exclusive label. For example, if the exclusive group label 'colors' is associated with groups 'blue', 'red', and 'green', then a component that is part of 'green' could not also be placed in 'red'. You can create, modify, or delete a group and its members. You can also use group names as filters for API calls. ### /partitions  A partition is a formal, non-overlapping division of the system that forms an administratively distinct sub-system. Each component may belong to at most one partition. Partitions are used as an access control mechanism or for implementing multi-tenancy. You can create, modify, or delete a partition and its members. You can also use partitions as filters for other API calls. ### /memberships  A membership shows the association of a component xname to its set of group labels and partition names. There can be many group labels and up to one partition per component. Memberships are not modified directly, as the underlying group or partition is modified instead. A component can be removed from one of the listed groups or partitions or added via POST as well as being present in the initial set of members when a partition or group is created. You can retrieve the memberships for components or memberships for a specific xname. ### /Inventory/DiscoveryStatus  Check discovery status for all components or you can track the status for a specific job ID. You can also check per-endpoint discover status for each RedfishEndpoint. Contains status information about the discovery operation for clients to query. The discover operation returns a link or links to status objects so that a client can determine when the discovery operation is complete. ### /Inventory/Discover  Discover subcomponents by querying all RedfishEndpoints. Once the RedfishEndpoint objects are created, inventory discovery will query these controllers and create or update management plane and managed plane objects representing the components (e.g. nodes, node enclosures, node cards for Mountain chassis CMM endpoints). ### /Subscriptions/SCN  Manage subscriptions to state change notifications (SCNs) from HSM. You can also subscribe to state change notifications by using the HMS Notification Fanout Daemon API. ## Workflows  ### Add and Delete a Redfish Endpoint #### POST /Inventory/RedfishEndpoints When you manually create Redfish endpoints, the discovery is automatically initiated. You would create Redfish endpoints for components that are not automatically discovered by REDS or MEDS. #### GET /Inventory/RedfishEndpoints Check the Redfish endpoints that have been added and check the status of discovery. #### DELETE /Inventory/RedfishEndpoints/{xname} Delete a specific Redfish endpoint. ### Perform Inventory Discovery #### POST /Inventory/Discover Start inventory discovery of a system's subcomponents by querying all Redfish endpoints. If needed, specify an ID or hostname (xname) in the payload. #### GET /Inventory/DiscoveryStatus Check the discovery status of all Redfish endpoints. You can also check the discovery status for each individual component by providing ID. ### Query and Update HMS Components (State/NID) #### GET /State/Components Retrieve all HMS Components found by inventory discovery as a named (\"Components\") array.  #### PATCH /State/Components/{xname}/Enabled Modify the component's Enabled field.  #### DELETE /State/Components/{xname} Delete a specific HMS component by providing its xname. As noted, components are not automatically deleted when RedfishEndpoints or ComponentEndpoints are deleted. ### Create and Delete a New Group #### GET /hsm/v2/State/Components Retrieve a list of desired components and their state. Select the nodes that you want to group.  #### POST /groups Create the new group with desired members. Provide a group label (required), description, name, members etc. in the JSON payload. #### GET /groups/{group_label} Retrieve the group that was create with the label. #### GET /State/Components/{group_label} Retrieve the current state for all the components in the group. #### DELETE /groups/{group_label} Delete the group specified by {group_label}. ## Valid State Transitions ``` Prior State -> New State     - Reason Ready       -> Standby       - HBTD if node has many missed heartbeats Ready       -> Ready/Warning - HBTD if node has a few missed heartbeats Standby     -> Ready         - HBTD Node re-starts heartbeating On          -> Ready         - HBTD Node started heartbeating Off         -> Ready         - HBTD sees heartbeats before Redfish Event (On) Standby     -> On            - Redfish Event (On) or if re-discovered while in the standby state Off         -> On            - Redfish Event (On) Standby     -> Off           - Redfish Event (Off) Ready       -> Off           - Redfish Event (Off) On          -> Off           - Redfish Event (Off) Any State   -> Empty         - Redfish Endpoint is disabled meaning component removal ``` Generally, nodes transition 'Off' -> 'On' -> 'Ready' when going from 'Off' to booted, and 'Ready' -> 'Ready/Warning' -> 'Standby' -> 'Off' when shutdown.
 *
 * API version: 1.0.0
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package hsm_client

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/antihax/optional"
)

// Linger please
var (
	_ context.Context
)

type HWInventoryHistoryApiService service

/*
HWInventoryHistoryApiService Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}
Delete history for an entry in the HWInventoryByFRU collection.
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  - @param fruid Locational xname of HWInventoryByFRU record to delete history for.

@return Response100
*/
func (a *HWInventoryHistoryApiService) DoHWInvHistByFRUDelete(ctx context.Context, fruid string) (Response100, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Response100
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/HardwareByFRU/History/{fruid}"
	localVarPath = strings.Replace(localVarPath, "{"+"fruid"+"}", fmt.Sprintf("%v", fruid), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Response100
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService Retrieve the history entries for the HWInventoryByFRU for {fruid}
Retrieve the history entries for the HWInventoryByFRU for a specific fruID.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param fruid Global HMS field-replaceable (FRU) identifier (serial number, etc.) of the hardware component to select.
 * @param optional nil or *HWInventoryHistoryApiDoHWInvHistByFRUGetOpts - Optional Parameters:
     * @param "Eventtype" (optional.String) -  Retrieve the history entries of a specific type (Added, Removed, etc) for a HWInventoryByFRU entry.
     * @param "Starttime" (optional.String) -  Retrieve the history entries from after the requested history window start time for a HWInventoryByFRU entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
     * @param "Endtime" (optional.String) -  Retrieve the history entries from before the requested history window end time for a HWInventoryByFRU entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
@return HwInventory100HwInventoryHistoryArray
*/

type HWInventoryHistoryApiDoHWInvHistByFRUGetOpts struct {
	Eventtype optional.String
	Starttime optional.String
	Endtime   optional.String
}

func (a *HWInventoryHistoryApiService) DoHWInvHistByFRUGet(ctx context.Context, fruid string, localVarOptionals *HWInventoryHistoryApiDoHWInvHistByFRUGetOpts) (HwInventory100HwInventoryHistoryArray, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue HwInventory100HwInventoryHistoryArray
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/HardwareByFRU/History/{fruid}"
	localVarPath = strings.Replace(localVarPath, "{"+"fruid"+"}", fmt.Sprintf("%v", fruid), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.Eventtype.IsSet() {
		localVarQueryParams.Add("eventtype", parameterToString(localVarOptionals.Eventtype.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Starttime.IsSet() {
		localVarQueryParams.Add("starttime", parameterToString(localVarOptionals.Starttime.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Endtime.IsSet() {
		localVarQueryParams.Add("endtime", parameterToString(localVarOptionals.Endtime.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v HwInventory100HwInventoryHistoryArray
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService Retrieve the history entries for all HWInventoryByFRU entries.
Retrieve the history entries for all HWInventoryByFRU entries. Sorted by FRU.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param optional nil or *HWInventoryHistoryApiDoHWInvHistByFRUsGetOpts - Optional Parameters:
     * @param "Fruid" (optional.String) -  Retrieve the history entries for HWInventoryByFRU entries with the given FRU ID.
     * @param "Eventtype" (optional.String) -  Retrieve the history entries of a specific type (Added, Removed, etc) for HWInventoryByFRU entries.
     * @param "Starttime" (optional.String) -  Retrieve the history entries from after the requested history window start time for HWInventoryByFRU entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
     * @param "Endtime" (optional.String) -  Retrieve the history entries from before the requested history window end time for HWInventoryByFRU entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
@return HwInventory100HwInventoryHistoryCollection
*/

type HWInventoryHistoryApiDoHWInvHistByFRUsGetOpts struct {
	Fruid     optional.String
	Eventtype optional.String
	Starttime optional.String
	Endtime   optional.String
}

func (a *HWInventoryHistoryApiService) DoHWInvHistByFRUsGet(ctx context.Context, localVarOptionals *HWInventoryHistoryApiDoHWInvHistByFRUsGetOpts) (HwInventory100HwInventoryHistoryCollection, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue HwInventory100HwInventoryHistoryCollection
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/HardwareByFRU/History"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.Fruid.IsSet() {
		localVarQueryParams.Add("fruid", parameterToString(localVarOptionals.Fruid.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Eventtype.IsSet() {
		localVarQueryParams.Add("eventtype", parameterToString(localVarOptionals.Eventtype.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Starttime.IsSet() {
		localVarQueryParams.Add("starttime", parameterToString(localVarOptionals.Starttime.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Endtime.IsSet() {
		localVarQueryParams.Add("endtime", parameterToString(localVarOptionals.Endtime.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v HwInventory100HwInventoryHistoryCollection
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService DELETE history for the HWInventoryByLocation entry with ID (location) {xname}
Delete history for the HWInventoryByLocation entry for a specific xname.
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
  - @param xname Locational xname of HWInventoryByLocation record to delete history for.

@return Response100
*/
func (a *HWInventoryHistoryApiService) DoHWInvHistByLocationDelete(ctx context.Context, xname string) (Response100, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Response100
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/Hardware/History/{xname}"
	localVarPath = strings.Replace(localVarPath, "{"+"xname"+"}", fmt.Sprintf("%v", xname), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Response100
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService Clear the HWInventory history.
Delete all HWInventory history entries. Note that this also deletes history for any associated HWInventoryByFRU entries.
  - @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().

@return Response100
*/
func (a *HWInventoryHistoryApiService) DoHWInvHistByLocationDeleteAll(ctx context.Context) (Response100, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Delete")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue Response100
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/Hardware/History"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v Response100
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService Retrieve the history entries for the HWInventoryByLocation entry at {xname}
Retrieve the history entries for a HWInventoryByLocation entry with a specific xname.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param xname Locational xname of hardware inventory record to return history for.
 * @param optional nil or *HWInventoryHistoryApiDoHWInvHistByLocationGetOpts - Optional Parameters:
     * @param "Eventtype" (optional.String) -  Retrieve the history entries of a specific type (Added, Removed, etc) for a HWInventoryByLocation entry.
     * @param "Starttime" (optional.String) -  Retrieve the history entries from after the requested history window start time for a HWInventoryByLocation entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
     * @param "Endtime" (optional.String) -  Retrieve the history entries from before the requested history window end time for a HWInventoryByLocation entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
@return HwInventory100HwInventoryHistoryArray
*/

type HWInventoryHistoryApiDoHWInvHistByLocationGetOpts struct {
	Eventtype optional.String
	Starttime optional.String
	Endtime   optional.String
}

func (a *HWInventoryHistoryApiService) DoHWInvHistByLocationGet(ctx context.Context, xname string, localVarOptionals *HWInventoryHistoryApiDoHWInvHistByLocationGetOpts) (HwInventory100HwInventoryHistoryArray, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue HwInventory100HwInventoryHistoryArray
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/Hardware/History/{xname}"
	localVarPath = strings.Replace(localVarPath, "{"+"xname"+"}", fmt.Sprintf("%v", xname), -1)

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.Eventtype.IsSet() {
		localVarQueryParams.Add("eventtype", parameterToString(localVarOptionals.Eventtype.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Starttime.IsSet() {
		localVarQueryParams.Add("starttime", parameterToString(localVarOptionals.Starttime.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Endtime.IsSet() {
		localVarQueryParams.Add("endtime", parameterToString(localVarOptionals.Endtime.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v HwInventory100HwInventoryHistoryArray
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}

/*
HWInventoryHistoryApiService Retrieve the history entries for all HWInventoryByLocation entries
Retrieve the history entries for all HWInventoryByLocation entries.
 * @param ctx context.Context - for authentication, logging, cancellation, deadlines, tracing, etc. Passed from http.Request or context.Background().
 * @param optional nil or *HWInventoryHistoryApiDoHWInvHistByLocationsGetOpts - Optional Parameters:
     * @param "Id" (optional.String) -  Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames.
     * @param "Eventtype" (optional.String) -  Retrieve the history entries of a specific type (Added, Removed, etc) for HWInventoryByLocation entries.
     * @param "Starttime" (optional.String) -  Retrieve the history entries from after the requested history window start time for HWInventoryByLocation entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
     * @param "Endtime" (optional.String) -  Retrieve the history entries from before the requested history window end time for HWInventoryByLocation entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00).
@return HwInventory100HwInventoryHistoryCollection
*/

type HWInventoryHistoryApiDoHWInvHistByLocationsGetOpts struct {
	Id        optional.String
	Eventtype optional.String
	Starttime optional.String
	Endtime   optional.String
}

func (a *HWInventoryHistoryApiService) DoHWInvHistByLocationsGet(ctx context.Context, localVarOptionals *HWInventoryHistoryApiDoHWInvHistByLocationsGetOpts) (HwInventory100HwInventoryHistoryCollection, *http.Response, error) {
	var (
		localVarHttpMethod  = strings.ToUpper("Get")
		localVarPostBody    interface{}
		localVarFileName    string
		localVarFileBytes   []byte
		localVarReturnValue HwInventory100HwInventoryHistoryCollection
	)

	// create path and map variables
	localVarPath := a.client.cfg.BasePath + "/Inventory/Hardware/History"

	localVarHeaderParams := make(map[string]string)
	localVarQueryParams := url.Values{}
	localVarFormParams := url.Values{}

	if localVarOptionals != nil && localVarOptionals.Id.IsSet() {
		localVarQueryParams.Add("id", parameterToString(localVarOptionals.Id.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Eventtype.IsSet() {
		localVarQueryParams.Add("eventtype", parameterToString(localVarOptionals.Eventtype.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Starttime.IsSet() {
		localVarQueryParams.Add("starttime", parameterToString(localVarOptionals.Starttime.Value(), ""))
	}
	if localVarOptionals != nil && localVarOptionals.Endtime.IsSet() {
		localVarQueryParams.Add("endtime", parameterToString(localVarOptionals.Endtime.Value(), ""))
	}
	// to determine the Content-Type header
	localVarHttpContentTypes := []string{}

	// set Content-Type header
	localVarHttpContentType := selectHeaderContentType(localVarHttpContentTypes)
	if localVarHttpContentType != "" {
		localVarHeaderParams["Content-Type"] = localVarHttpContentType
	}

	// to determine the Accept header
	localVarHttpHeaderAccepts := []string{"application/json", "application/problem+json"}

	// set Accept header
	localVarHttpHeaderAccept := selectHeaderAccept(localVarHttpHeaderAccepts)
	if localVarHttpHeaderAccept != "" {
		localVarHeaderParams["Accept"] = localVarHttpHeaderAccept
	}
	r, err := a.client.prepareRequest(ctx, localVarPath, localVarHttpMethod, localVarPostBody, localVarHeaderParams, localVarQueryParams, localVarFormParams, localVarFileName, localVarFileBytes)
	if err != nil {
		return localVarReturnValue, nil, err
	}

	localVarHttpResponse, err := a.client.callAPI(r)
	if err != nil || localVarHttpResponse == nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	localVarBody, err := ioutil.ReadAll(localVarHttpResponse.Body)
	localVarHttpResponse.Body.Close()
	if err != nil {
		return localVarReturnValue, localVarHttpResponse, err
	}

	if localVarHttpResponse.StatusCode < 300 {
		// If we succeed, return the data, otherwise pass on to decode error.
		err = a.client.decode(&localVarReturnValue, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
		if err == nil {
			return localVarReturnValue, localVarHttpResponse, err
		}
	}

	if localVarHttpResponse.StatusCode >= 300 {
		newErr := GenericSwaggerError{
			body:  localVarBody,
			error: localVarHttpResponse.Status,
		}
		if localVarHttpResponse.StatusCode == 200 {
			var v HwInventory100HwInventoryHistoryCollection
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 400 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 404 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		if localVarHttpResponse.StatusCode == 0 {
			var v Problem7807
			err = a.client.decode(&v, localVarBody, localVarHttpResponse.Header.Get("Content-Type"))
			if err != nil {
				newErr.error = err.Error()
				return localVarReturnValue, localVarHttpResponse, newErr
			}
			newErr.model = v
			return localVarReturnValue, localVarHttpResponse, newErr
		}
		return localVarReturnValue, localVarHttpResponse, newErr
	}

	return localVarReturnValue, localVarHttpResponse, nil
}
