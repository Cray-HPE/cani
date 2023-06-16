# Go API client for hsm_client

The Hardware State Manager (HSM) inventories, monitors, and manages hardware, and tracks the logical and dynamic component states, such as roles, NIDs, and other basic metadata needed to provide most common administrative and operational functions. HSM is the single source of truth for the state of the system. It contains the component state and information on Redfish endpoints for communicating with components via Redfish. It also allows administrators to create partitions and groups for other uses. ## Resources ### /State/Components HMS components are created during inventory discovery and provide a higher-level representation of the component, including state, NID, role (i.e. compute/service), subtype, and so on. Unlike ComponentEndpoints, however, they are not strictly linked to the parent RedfishEndpoint, and are not automatically deleted when the RedfishEndpoints are (though they can be deleted via a separate call). This is because these components can also represent abstract components, such as removed components (e.g. which would remain, but have their states changed to \"Empty\" upon removal). ### /Defaults/NodeMaps  This resource allows a mapping file (NodeMaps) to be uploaded that maps node xnames to Node IDs, and optionally, to roles and subroles. These mappings are used when discovering nodes for the first time. These mappings should be uploaded prior to discovery and should contain mappings for each valid node xname in the system, whether populated or not. Nodemap is a JSON file that contains the xname of the node, node ID, and optionally role and subrole. Role can be Compute, Application, Storage, Management etc. The NodeMaps collection can be uploaded to HSM automatically at install time by specifying it as a JSON file. As a result, the endpoints are then automatically discovered by REDS, and inventory discovery is performed by HSM. The desired NID numbers will be set as soon as the nodes are created using the NodeMaps collection.  It is recommended that Nodemaps are uploaded at install time before discovery happens. If they are uploaded after discovery, then the node xnames need to be manually updated with the correct NIDs. You can update NIDs for individual components by using PATCH /State/Components/{xname}/NID.  ### /Inventory/Hardware  This resource shows the hardware inventory of the entire system and contains FRU information in location. All entries are displayed as a flat array. ### /Inventory/HardwareByFRU  Every component has FRU information. This resource shows the hardware inventory for all FRUs or for a specific FRU irrespective of the location. This information is constant regardless of where the hardware item is currently in the system. If a HWInventoryByLocation entry is currently populated with a piece of hardware, it will have the corresponding HWInventoryByFRU object embedded. This FRU info can also be looked up by FRU ID regardless of the current location. ### /Inventory/Hardware/Query/{xname}  This resource gets you information about a specific component and it's sub-components. The xname can be a component, partition, ALL, or s0. Both ALL and s0 represent the entire system. ### /Inventory/RedfishEndpoints  This is a BMC or other Redfish controller that has a Redfish entry point and Redfish service root. It is used to discover the components managed by this endpoint during discovery and handles all Redfish interactions by these subcomponents.  If the endpoint has been discovered, this entry will include the ComponentEndpoint entries for these managed subcomponents. You can also create a Redfish Endpoint or update the definition for a Redfish Endpoint. The xname identifies the location of all components in the system, including chassis, controllers, nodes, and so on. Redfish endpoints are given to State Manager. ### /Inventory/ComponentEndpoints  Component Endpoints are the specific URLs for each individual component that are under the Redfish endpoint. Component endpoints are discovered during inventory discovery. They are the management-plane representation of system components and are linked to the parent Redfish Endpoint. They provide a glue layer to bridge the higher-level representation of a component with how it is represented locally by Redfish.  The collection of ComponentEndpoints can be obtained in full, optionally filtered on certain criteria (e.g. obtain just Node components), or accessed by their xname IDs individually. ### /Inventory/ServiceEndpoints  ServiceEndpoints help you do things on Redfish like updating the firmware. They are discovered during inventory discovery. ### /groups  Groups are named sets of system components, most commonly nodes. A group groups components under an administratively chosen label (group name). Each component may belong to any number of groups. If a group has exclusiveGroup=<excl-label> set, then a node may only be a member of one group that matches that exclusive label. For example, if the exclusive group label 'colors' is associated with groups 'blue', 'red', and 'green', then a component that is part of 'green' could not also be placed in 'red'. You can create, modify, or delete a group and its members. You can also use group names as filters for API calls. ### /partitions  A partition is a formal, non-overlapping division of the system that forms an administratively distinct sub-system. Each component may belong to at most one partition. Partitions are used as an access control mechanism or for implementing multi-tenancy. You can create, modify, or delete a partition and its members. You can also use partitions as filters for other API calls. ### /memberships  A membership shows the association of a component xname to its set of group labels and partition names. There can be many group labels and up to one partition per component. Memberships are not modified directly, as the underlying group or partition is modified instead. A component can be removed from one of the listed groups or partitions or added via POST as well as being present in the initial set of members when a partition or group is created. You can retrieve the memberships for components or memberships for a specific xname. ### /Inventory/DiscoveryStatus  Check discovery status for all components or you can track the status for a specific job ID. You can also check per-endpoint discover status for each RedfishEndpoint. Contains status information about the discovery operation for clients to query. The discover operation returns a link or links to status objects so that a client can determine when the discovery operation is complete. ### /Inventory/Discover  Discover subcomponents by querying all RedfishEndpoints. Once the RedfishEndpoint objects are created, inventory discovery will query these controllers and create or update management plane and managed plane objects representing the components (e.g. nodes, node enclosures, node cards for Mountain chassis CMM endpoints). ### /Subscriptions/SCN  Manage subscriptions to state change notifications (SCNs) from HSM. You can also subscribe to state change notifications by using the HMS Notification Fanout Daemon API. ## Workflows  ### Add and Delete a Redfish Endpoint #### POST /Inventory/RedfishEndpoints When you manually create Redfish endpoints, the discovery is automatically initiated. You would create Redfish endpoints for components that are not automatically discovered by REDS or MEDS. #### GET /Inventory/RedfishEndpoints Check the Redfish endpoints that have been added and check the status of discovery. #### DELETE /Inventory/RedfishEndpoints/{xname} Delete a specific Redfish endpoint. ### Perform Inventory Discovery #### POST /Inventory/Discover Start inventory discovery of a system's subcomponents by querying all Redfish endpoints. If needed, specify an ID or hostname (xname) in the payload. #### GET /Inventory/DiscoveryStatus Check the discovery status of all Redfish endpoints. You can also check the discovery status for each individual component by providing ID. ### Query and Update HMS Components (State/NID) #### GET /State/Components Retrieve all HMS Components found by inventory discovery as a named (\"Components\") array.  #### PATCH /State/Components/{xname}/Enabled Modify the component's Enabled field.  #### DELETE /State/Components/{xname} Delete a specific HMS component by providing its xname. As noted, components are not automatically deleted when RedfishEndpoints or ComponentEndpoints are deleted. ### Create and Delete a New Group #### GET /hsm/v2/State/Components Retrieve a list of desired components and their state. Select the nodes that you want to group.  #### POST /groups Create the new group with desired members. Provide a group label (required), description, name, members etc. in the JSON payload. #### GET /groups/{group_label} Retrieve the group that was create with the label. #### GET /State/Components/{group_label} Retrieve the current state for all the components in the group. #### DELETE /groups/{group_label} Delete the group specified by {group_label}. ## Valid State Transitions ``` Prior State -> New State     - Reason Ready       -> Standby       - HBTD if node has many missed heartbeats Ready       -> Ready/Warning - HBTD if node has a few missed heartbeats Standby     -> Ready         - HBTD Node re-starts heartbeating On          -> Ready         - HBTD Node started heartbeating Off         -> Ready         - HBTD sees heartbeats before Redfish Event (On) Standby     -> On            - Redfish Event (On) or if re-discovered while in the standby state Off         -> On            - Redfish Event (On) Standby     -> Off           - Redfish Event (Off) Ready       -> Off           - Redfish Event (Off) On          -> Off           - Redfish Event (Off) Any State   -> Empty         - Redfish Endpoint is disabled meaning component removal ``` Generally, nodes transition 'Off' -> 'On' -> 'Ready' when going from 'Off' to booted, and 'Ready' -> 'Ready/Warning' -> 'Standby' -> 'Off' when shutdown.

## Overview
This API client was generated by the [swagger-codegen](https://github.com/swagger-api/swagger-codegen) project.  By using the [swagger-spec](https://github.com/swagger-api/swagger-spec) from a remote server, you can easily generate an API client.

- API version: 1.0.0
- Package version: 1.0.0
- Build package: io.swagger.codegen.v3.generators.go.GoClientCodegen

## Installation
Put the package under your project folder and add the following in import:
```golang
import "./hsm_client"
```

## Documentation for API Endpoints

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AdminLocksApi* | [**LocksDisablePost**](docs/AdminLocksApi.md#locksdisablepost) | **Post** /locks/disable | Disables the ability to create a reservation on components.
*AdminLocksApi* | [**LocksLockPost**](docs/AdminLocksApi.md#lockslockpost) | **Post** /locks/lock | Locks components.
*AdminLocksApi* | [**LocksRepairPost**](docs/AdminLocksApi.md#locksrepairpost) | **Post** /locks/repair | Repair components lock and reservation ability.
*AdminLocksApi* | [**LocksStatusGet**](docs/AdminLocksApi.md#locksstatusget) | **Get** /locks/status | Retrieve lock status for all components or a filtered subset of components.
*AdminLocksApi* | [**LocksStatusPost**](docs/AdminLocksApi.md#locksstatuspost) | **Post** /locks/status | Retrieve lock status for component IDs.
*AdminLocksApi* | [**LocksUnlockPost**](docs/AdminLocksApi.md#locksunlockpost) | **Post** /locks/unlock | Unlocks components.
*AdminReservationsApi* | [**LocksReservationsPost**](docs/AdminReservationsApi.md#locksreservationspost) | **Post** /locks/reservations | Create reservations
*AdminReservationsApi* | [**LocksReservationsReleasePost**](docs/AdminReservationsApi.md#locksreservationsreleasepost) | **Post** /locks/reservations/release | Releases existing reservations.
*AdminReservationsApi* | [**LocksReservationsRemovePost**](docs/AdminReservationsApi.md#locksreservationsremovepost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.
*CliDangerThisWillDeleteAllComponentEndpointsContinueApi* | [**DoComponentEndpointsDeleteAll**](docs/CliDangerThisWillDeleteAllComponentEndpointsContinueApi.md#docomponentendpointsdeleteall) | **Delete** /Inventory/ComponentEndpoints | Delete all ComponentEndpoints
*CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApi* | [**DoCompEthInterfaceDeleteAllV2**](docs/CliDangerThisWillDeleteAllComponentEthernetInterfacesContinueApi.md#docompethinterfacedeleteallv2) | **Delete** /Inventory/EthernetInterfaces | Clear the component Ethernet interface collection.
*CliDangerThisWillDeleteAllComponentsInHSMContinueApi* | [**DoComponentsDeleteAll**](docs/CliDangerThisWillDeleteAllComponentsInHSMContinueApi.md#docomponentsdeleteall) | **Delete** /State/Components | Delete all components
*CliDangerThisWillDeleteAllFRUsForHSMContinueApi* | [**DoHWInvByFRUDeleteAll**](docs/CliDangerThisWillDeleteAllFRUsForHSMContinueApi.md#dohwinvbyfrudeleteall) | **Delete** /Inventory/HardwareByFRU | Delete all HWInventoryByFRU entries
*CliDangerThisWillDeleteAllHardwareHistoryContinueApi* | [**DoHWInvHistByLocationDeleteAll**](docs/CliDangerThisWillDeleteAllHardwareHistoryContinueApi.md#dohwinvhistbylocationdeleteall) | **Delete** /Inventory/Hardware/History | Clear the HWInventory history.
*CliDangerThisWillDeleteAllHardwareInventoryContinueApi* | [**DoHWInvByLocationDeleteAll**](docs/CliDangerThisWillDeleteAllHardwareInventoryContinueApi.md#dohwinvbylocationdeleteall) | **Delete** /Inventory/Hardware | Delete all HWInventoryByLocation entries
*CliDangerThisWillDeleteAllHistoryForThisFRUContinueApi* | [**DoHWInvHistByFRUDelete**](docs/CliDangerThisWillDeleteAllHistoryForThisFRUContinueApi.md#dohwinvhistbyfrudelete) | **Delete** /Inventory/HardwareByFRU/History/{fruid} | Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}
*CliDangerThisWillDeleteAllHistoryForThisXnameContinueApi* | [**DoHWInvHistByLocationDelete**](docs/CliDangerThisWillDeleteAllHistoryForThisXnameContinueApi.md#dohwinvhistbylocationdelete) | **Delete** /Inventory/Hardware/History/{xname} | DELETE history for the HWInventoryByLocation entry with ID (location) {xname}
*CliDangerThisWillDeleteAllNodeMapsContinueApi* | [**DoNodeMapsDeleteAll**](docs/CliDangerThisWillDeleteAllNodeMapsContinueApi.md#donodemapsdeleteall) | **Delete** /Defaults/NodeMaps | Delete all NodeMap entities
*CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApi* | [**DoRedfishEndpointsDeleteAll**](docs/CliDangerThisWillDeleteAllRedfishEndpointsInHSMContinueApi.md#doredfishendpointsdeleteall) | **Delete** /Inventory/RedfishEndpoints | Delete all RedfishEndpoints
*CliDangerThisWillDeleteAllServiceEndpointsContinueApi* | [**DoServiceEndpointsDeleteAll**](docs/CliDangerThisWillDeleteAllServiceEndpointsContinueApi.md#doserviceendpointsdeleteall) | **Delete** /Inventory/ServiceEndpoints | Delete all ServiceEndpoints
*CliIgnoreApi* | [**DoDeleteSCNSubscription**](docs/CliIgnoreApi.md#dodeletescnsubscription) | **Delete** /Subscriptions/SCN/{id} | Delete a state change notification subscription
*CliIgnoreApi* | [**DoDeleteSCNSubscriptionsAll**](docs/CliIgnoreApi.md#dodeletescnsubscriptionsall) | **Delete** /Subscriptions/SCN | Delete all state change notification subscriptions
*CliIgnoreApi* | [**DoGetSCNSubscription**](docs/CliIgnoreApi.md#dogetscnsubscription) | **Get** /Subscriptions/SCN/{id} | Retrieve a currently-held state change notification subscription
*CliIgnoreApi* | [**DoGetSCNSubscriptionsAll**](docs/CliIgnoreApi.md#dogetscnsubscriptionsall) | **Get** /Subscriptions/SCN | Retrieve currently-held state change notification subscriptions
*CliIgnoreApi* | [**DoHWInvByLocationPost**](docs/CliIgnoreApi.md#dohwinvbylocationpost) | **Post** /Inventory/Hardware | Create/Update hardware inventory entries
*CliIgnoreApi* | [**DoPatchSCNSubscription**](docs/CliIgnoreApi.md#dopatchscnsubscription) | **Patch** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
*CliIgnoreApi* | [**DoPostSCNSubscription**](docs/CliIgnoreApi.md#dopostscnsubscription) | **Post** /Subscriptions/SCN | Create a subscription for state change notifications
*CliIgnoreApi* | [**DoPowerMapsDeleteAll**](docs/CliIgnoreApi.md#dopowermapsdeleteall) | **Delete** /sysinfo/powermaps | Delete all PowerMap entities
*CliIgnoreApi* | [**DoPutSCNSubscription**](docs/CliIgnoreApi.md#doputscnsubscription) | **Put** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
*CliIgnoreApi* | [**DoRedfishEndpointPut**](docs/CliIgnoreApi.md#doredfishendpointput) | **Put** /Inventory/RedfishEndpoints/{xname} | Update definition for RedfishEndpoint ID {xname}
*CliIgnoreApi* | [**LocksReservationsPost**](docs/CliIgnoreApi.md#locksreservationspost) | **Post** /locks/reservations | Create reservations
*CliIgnoreApi* | [**LocksReservationsReleasePost**](docs/CliIgnoreApi.md#locksreservationsreleasepost) | **Post** /locks/reservations/release | Releases existing reservations.
*CliIgnoreApi* | [**LocksReservationsRemovePost**](docs/CliIgnoreApi.md#locksreservationsremovepost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.
*CliIgnoreApi* | [**LocksServiceReservationsCheckPost**](docs/CliIgnoreApi.md#locksservicereservationscheckpost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
*CliIgnoreApi* | [**LocksServiceReservationsPost**](docs/CliIgnoreApi.md#locksservicereservationspost) | **Post** /locks/service/reservations | Create reservations
*CliIgnoreApi* | [**LocksServiceReservationsReleasePost**](docs/CliIgnoreApi.md#locksservicereservationsreleasepost) | **Post** /locks/service/reservations/release | Releases existing reservations.
*CliIgnoreApi* | [**LocksServiceReservationsRenewPost**](docs/CliIgnoreApi.md#locksservicereservationsrenewpost) | **Post** /locks/service/reservations/renew | Renew existing reservations.
*ComponentApi* | [**DoCompArrayNIDPatch**](docs/ComponentApi.md#docomparraynidpatch) | **Patch** /State/Components/BulkNID | Update multiple components&#x27; NIDs via ComponentArray
*ComponentApi* | [**DoCompBulkEnabledPatch**](docs/ComponentApi.md#docompbulkenabledpatch) | **Patch** /State/Components/BulkEnabled | Update multiple components&#x27; Enabled values via a list of xnames
*ComponentApi* | [**DoCompBulkFlagOnlyPatch**](docs/ComponentApi.md#docompbulkflagonlypatch) | **Patch** /State/Components/BulkFlagOnly | Update multiple components&#x27; Flag values via a list of xnames
*ComponentApi* | [**DoCompBulkRolePatch**](docs/ComponentApi.md#docompbulkrolepatch) | **Patch** /State/Components/BulkRole | Update multiple components&#x27; Role values via a list of xnames
*ComponentApi* | [**DoCompBulkStateDataPatch**](docs/ComponentApi.md#docompbulkstatedatapatch) | **Patch** /State/Components/BulkStateData | Update multiple components&#x27; state data via a list of xnames
*ComponentApi* | [**DoCompBulkSwStatusPatch**](docs/ComponentApi.md#docompbulkswstatuspatch) | **Patch** /State/Components/BulkSoftwareStatus | Update multiple components&#x27; SoftwareStatus values via a list of xnames
*ComponentApi* | [**DoCompEnabledPatch**](docs/ComponentApi.md#docompenabledpatch) | **Patch** /State/Components/{xname}/Enabled | Update component Enabled value at {xname}
*ComponentApi* | [**DoCompFlagOnlyPatch**](docs/ComponentApi.md#docompflagonlypatch) | **Patch** /State/Components/{xname}/FlagOnly | Update component Flag value at {xname}
*ComponentApi* | [**DoCompNIDPatch**](docs/ComponentApi.md#docompnidpatch) | **Patch** /State/Components/{xname}/NID | Update component NID value at {xname}
*ComponentApi* | [**DoCompRolePatch**](docs/ComponentApi.md#docomprolepatch) | **Patch** /State/Components/{xname}/Role | Update component Role and SubRole values at {xname}
*ComponentApi* | [**DoCompStatePatch**](docs/ComponentApi.md#docompstatepatch) | **Patch** /State/Components/{xname}/StateData | Update component state data at {xname}
*ComponentApi* | [**DoCompSwStatusPatch**](docs/ComponentApi.md#docompswstatuspatch) | **Patch** /State/Components/{xname}/SoftwareStatus | Update component SoftwareStatus value at {xname}
*ComponentApi* | [**DoComponentByNIDGet**](docs/ComponentApi.md#docomponentbynidget) | **Get** /State/Components/ByNID/{nid} | Retrieve component with NID&#x3D;{nid}
*ComponentApi* | [**DoComponentByNIDQueryPost**](docs/ComponentApi.md#docomponentbynidquerypost) | **Post** /State/Components/ByNID/Query | Create component query (by NID ranges), returning ComponentArray
*ComponentApi* | [**DoComponentDelete**](docs/ComponentApi.md#docomponentdelete) | **Delete** /State/Components/{xname} | Delete component with ID {xname}
*ComponentApi* | [**DoComponentGet**](docs/ComponentApi.md#docomponentget) | **Get** /State/Components/{xname} | Retrieve component at {xname}
*ComponentApi* | [**DoComponentPut**](docs/ComponentApi.md#docomponentput) | **Put** /State/Components/{xname} | Create/Update an HMS Component
*ComponentApi* | [**DoComponentQueryGet**](docs/ComponentApi.md#docomponentqueryget) | **Get** /State/Components/Query/{xname} | Retrieve component query for {xname}, returning ComponentArray
*ComponentApi* | [**DoComponentsDeleteAll**](docs/ComponentApi.md#docomponentsdeleteall) | **Delete** /State/Components | Delete all components
*ComponentApi* | [**DoComponentsGet**](docs/ComponentApi.md#docomponentsget) | **Get** /State/Components | Retrieve collection of HMS Components
*ComponentApi* | [**DoComponentsPost**](docs/ComponentApi.md#docomponentspost) | **Post** /State/Components | Create/Update a collection of HMS Components
*ComponentApi* | [**DoComponentsQueryPost**](docs/ComponentApi.md#docomponentsquerypost) | **Post** /State/Components/Query | Create component query (by xname list), returning ComponentArray
*ComponentEndpointApi* | [**DoComponentEndpointDelete**](docs/ComponentEndpointApi.md#docomponentendpointdelete) | **Delete** /Inventory/ComponentEndpoints/{xname} | Delete ComponentEndpoint with ID {xname}
*ComponentEndpointApi* | [**DoComponentEndpointGet**](docs/ComponentEndpointApi.md#docomponentendpointget) | **Get** /Inventory/ComponentEndpoints/{xname} | Retrieve ComponentEndpoint at {xname}
*ComponentEndpointApi* | [**DoComponentEndpointsDeleteAll**](docs/ComponentEndpointApi.md#docomponentendpointsdeleteall) | **Delete** /Inventory/ComponentEndpoints | Delete all ComponentEndpoints
*ComponentEndpointApi* | [**DoComponentEndpointsGet**](docs/ComponentEndpointApi.md#docomponentendpointsget) | **Get** /Inventory/ComponentEndpoints | Retrieve ComponentEndpoints Collection
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceDeleteAllV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacedeleteallv2) | **Delete** /Inventory/EthernetInterfaces | Clear the component Ethernet interface collection.
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceDeleteV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacedeletev2) | **Delete** /Inventory/EthernetInterfaces/{ethInterfaceID} | DELETE existing component Ethernet interface with {ethInterfaceID}
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceGetV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacegetv2) | **Get** /Inventory/EthernetInterfaces/{ethInterfaceID} | GET existing component Ethernet interface {ethInterfaceID}
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceIPAddressDeleteV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfaceipaddressdeletev2) | **Delete** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses/{ipAddress} | DELETE existing IP address mapping with {ipAddress} from a component Ethernet interface with {ethInterfaceID}
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceIPAddressPatchV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfaceipaddresspatchv2) | **Patch** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses/{ipAddress} | UPDATE metadata for existing IP address {ipAddress} in a component Ethernet interface {ethInterfaceID
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceIPAddressesGetV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfaceipaddressesgetv2) | **Get** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses | Retrieve all IP addresses of a component Ethernet interface {ethInterfaceID}
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfaceIPAddressesPostV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfaceipaddressespostv2) | **Post** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses | CREATE a new IP address mapping in a component Ethernet interface (via POST)
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfacePatchV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacepatchv2) | **Patch** /Inventory/EthernetInterfaces/{ethInterfaceID} | UPDATE metadata for existing component Ethernet interface {ethInterfaceID}
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfacePostV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacepostv2) | **Post** /Inventory/EthernetInterfaces | CREATE a new component Ethernet interface (via POST)
*ComponentEthernetInterfacesApi* | [**DoCompEthInterfacesGetV2**](docs/ComponentEthernetInterfacesApi.md#docompethinterfacesgetv2) | **Get** /Inventory/EthernetInterfaces | GET ALL existing component Ethernet interfaces
*DiscoverApi* | [**DoInventoryDiscoverPost**](docs/DiscoverApi.md#doinventorydiscoverpost) | **Post** /Inventory/Discover | Create Discover operation request
*DiscoveryStatusApi* | [**DoDiscoveryStatusGet**](docs/DiscoveryStatusApi.md#dodiscoverystatusget) | **Get** /Inventory/DiscoveryStatus/{id} | Retrieve DiscoveryStatus entry matching {id}
*DiscoveryStatusApi* | [**DoDiscoveryStatusGetAll**](docs/DiscoveryStatusApi.md#dodiscoverystatusgetall) | **Get** /Inventory/DiscoveryStatus | Retrieve all DiscoveryStatus entries in collection
*GroupApi* | [**DoGroupDelete**](docs/GroupApi.md#dogroupdelete) | **Delete** /groups/{group_label} | Delete existing group with {group_label}
*GroupApi* | [**DoGroupGet**](docs/GroupApi.md#dogroupget) | **Get** /groups/{group_label} | Retrieve existing group {group_label}
*GroupApi* | [**DoGroupLabelsGet**](docs/GroupApi.md#dogrouplabelsget) | **Get** /groups/labels | Retrieve all existing group labels
*GroupApi* | [**DoGroupMemberDelete**](docs/GroupApi.md#dogroupmemberdelete) | **Delete** /groups/{group_label}/members/{xname_id} | Delete member from existing group
*GroupApi* | [**DoGroupMembersGet**](docs/GroupApi.md#dogroupmembersget) | **Get** /groups/{group_label}/members | Retrieve all members of existing group
*GroupApi* | [**DoGroupMembersPost**](docs/GroupApi.md#dogroupmemberspost) | **Post** /groups/{group_label}/members | Create new member of existing group (via POST)
*GroupApi* | [**DoGroupPatch**](docs/GroupApi.md#dogrouppatch) | **Patch** /groups/{group_label} | Update metadata for existing group {group_label}
*GroupApi* | [**DoGroupsGet**](docs/GroupApi.md#dogroupsget) | **Get** /groups | Retrieve all existing groups
*GroupApi* | [**DoGroupsPost**](docs/GroupApi.md#dogroupspost) | **Post** /groups | Create a new group
*HWInventoryApi* | [**DoHWInvByLocationQueryGet**](docs/HWInventoryApi.md#dohwinvbylocationqueryget) | **Get** /Inventory/Hardware/Query/{xname} | Retrieve results of HWInventory query starting at {xname}
*HWInventoryByFRUApi* | [**DoHWInvByFRUDelete**](docs/HWInventoryByFRUApi.md#dohwinvbyfrudelete) | **Delete** /Inventory/HardwareByFRU/{fruid} | Delete HWInventoryByFRU entry with FRU identifier {fruid}
*HWInventoryByFRUApi* | [**DoHWInvByFRUDeleteAll**](docs/HWInventoryByFRUApi.md#dohwinvbyfrudeleteall) | **Delete** /Inventory/HardwareByFRU | Delete all HWInventoryByFRU entries
*HWInventoryByFRUApi* | [**DoHWInvByFRUGet**](docs/HWInventoryByFRUApi.md#dohwinvbyfruget) | **Get** /Inventory/HardwareByFRU/{fruid} | Retrieve HWInventoryByFRU for {fruid}
*HWInventoryByFRUApi* | [**DoHWInvByFRUGetAll**](docs/HWInventoryByFRUApi.md#dohwinvbyfrugetall) | **Get** /Inventory/HardwareByFRU | Retrieve all HWInventoryByFRU entries in a flat array
*HWInventoryByLocationApi* | [**DoHWInvByLocationDelete**](docs/HWInventoryByLocationApi.md#dohwinvbylocationdelete) | **Delete** /Inventory/Hardware/{xname} | DELETE HWInventoryByLocation entry with ID (location) {xname}
*HWInventoryByLocationApi* | [**DoHWInvByLocationDeleteAll**](docs/HWInventoryByLocationApi.md#dohwinvbylocationdeleteall) | **Delete** /Inventory/Hardware | Delete all HWInventoryByLocation entries
*HWInventoryByLocationApi* | [**DoHWInvByLocationGet**](docs/HWInventoryByLocationApi.md#dohwinvbylocationget) | **Get** /Inventory/Hardware/{xname} | Retrieve HWInventoryByLocation entry at {xname}
*HWInventoryByLocationApi* | [**DoHWInvByLocationGetAll**](docs/HWInventoryByLocationApi.md#dohwinvbylocationgetall) | **Get** /Inventory/Hardware | Retrieve all HWInventoryByLocation entries in array
*HWInventoryByLocationApi* | [**DoHWInvByLocationPost**](docs/HWInventoryByLocationApi.md#dohwinvbylocationpost) | **Post** /Inventory/Hardware | Create/Update hardware inventory entries
*HWInventoryHistoryApi* | [**DoHWInvHistByFRUDelete**](docs/HWInventoryHistoryApi.md#dohwinvhistbyfrudelete) | **Delete** /Inventory/HardwareByFRU/History/{fruid} | Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}
*HWInventoryHistoryApi* | [**DoHWInvHistByFRUGet**](docs/HWInventoryHistoryApi.md#dohwinvhistbyfruget) | **Get** /Inventory/HardwareByFRU/History/{fruid} | Retrieve the history entries for the HWInventoryByFRU for {fruid}
*HWInventoryHistoryApi* | [**DoHWInvHistByFRUsGet**](docs/HWInventoryHistoryApi.md#dohwinvhistbyfrusget) | **Get** /Inventory/HardwareByFRU/History | Retrieve the history entries for all HWInventoryByFRU entries.
*HWInventoryHistoryApi* | [**DoHWInvHistByLocationDelete**](docs/HWInventoryHistoryApi.md#dohwinvhistbylocationdelete) | **Delete** /Inventory/Hardware/History/{xname} | DELETE history for the HWInventoryByLocation entry with ID (location) {xname}
*HWInventoryHistoryApi* | [**DoHWInvHistByLocationDeleteAll**](docs/HWInventoryHistoryApi.md#dohwinvhistbylocationdeleteall) | **Delete** /Inventory/Hardware/History | Clear the HWInventory history.
*HWInventoryHistoryApi* | [**DoHWInvHistByLocationGet**](docs/HWInventoryHistoryApi.md#dohwinvhistbylocationget) | **Get** /Inventory/Hardware/History/{xname} | Retrieve the history entries for the HWInventoryByLocation entry at {xname}
*HWInventoryHistoryApi* | [**DoHWInvHistByLocationsGet**](docs/HWInventoryHistoryApi.md#dohwinvhistbylocationsget) | **Get** /Inventory/Hardware/History | Retrieve the history entries for all HWInventoryByLocation entries
*LockingApi* | [**LocksDisablePost**](docs/LockingApi.md#locksdisablepost) | **Post** /locks/disable | Disables the ability to create a reservation on components.
*LockingApi* | [**LocksLockPost**](docs/LockingApi.md#lockslockpost) | **Post** /locks/lock | Locks components.
*LockingApi* | [**LocksRepairPost**](docs/LockingApi.md#locksrepairpost) | **Post** /locks/repair | Repair components lock and reservation ability.
*LockingApi* | [**LocksReservationsPost**](docs/LockingApi.md#locksreservationspost) | **Post** /locks/reservations | Create reservations
*LockingApi* | [**LocksReservationsReleasePost**](docs/LockingApi.md#locksreservationsreleasepost) | **Post** /locks/reservations/release | Releases existing reservations.
*LockingApi* | [**LocksReservationsRemovePost**](docs/LockingApi.md#locksreservationsremovepost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.
*LockingApi* | [**LocksServiceReservationsCheckPost**](docs/LockingApi.md#locksservicereservationscheckpost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
*LockingApi* | [**LocksServiceReservationsPost**](docs/LockingApi.md#locksservicereservationspost) | **Post** /locks/service/reservations | Create reservations
*LockingApi* | [**LocksServiceReservationsReleasePost**](docs/LockingApi.md#locksservicereservationsreleasepost) | **Post** /locks/service/reservations/release | Releases existing reservations.
*LockingApi* | [**LocksServiceReservationsRenewPost**](docs/LockingApi.md#locksservicereservationsrenewpost) | **Post** /locks/service/reservations/renew | Renew existing reservations.
*LockingApi* | [**LocksStatusGet**](docs/LockingApi.md#locksstatusget) | **Get** /locks/status | Retrieve lock status for all components or a filtered subset of components.
*LockingApi* | [**LocksStatusPost**](docs/LockingApi.md#locksstatuspost) | **Post** /locks/status | Retrieve lock status for component IDs.
*LockingApi* | [**LocksUnlockPost**](docs/LockingApi.md#locksunlockpost) | **Post** /locks/unlock | Unlocks components.
*MembershipApi* | [**DoMembershipGet**](docs/MembershipApi.md#domembershipget) | **Get** /memberships/{xname} | Retrieve membership for component {xname}
*MembershipApi* | [**DoMembershipsGet**](docs/MembershipApi.md#domembershipsget) | **Get** /memberships | Retrieve all memberships for components
*NodeMapApi* | [**DoNodeMapDelete**](docs/NodeMapApi.md#donodemapdelete) | **Delete** /Defaults/NodeMaps/{xname} | Delete NodeMap with ID {xname}
*NodeMapApi* | [**DoNodeMapGet**](docs/NodeMapApi.md#donodemapget) | **Get** /Defaults/NodeMaps/{xname} | Retrieve NodeMap at {xname}
*NodeMapApi* | [**DoNodeMapPost**](docs/NodeMapApi.md#donodemappost) | **Post** /Defaults/NodeMaps | Create or Modify NodeMaps
*NodeMapApi* | [**DoNodeMapPut**](docs/NodeMapApi.md#donodemapput) | **Put** /Defaults/NodeMaps/{xname} | Update definition for NodeMap ID {xname}
*NodeMapApi* | [**DoNodeMapsDeleteAll**](docs/NodeMapApi.md#donodemapsdeleteall) | **Delete** /Defaults/NodeMaps | Delete all NodeMap entities
*NodeMapApi* | [**DoNodeMapsGet**](docs/NodeMapApi.md#donodemapsget) | **Get** /Defaults/NodeMaps | Retrieve all NodeMaps, returning NodeMapArray
*PartitionApi* | [**DoPartitionDelete**](docs/PartitionApi.md#dopartitiondelete) | **Delete** /partitions/{partition_name} | Delete existing partition with {partition_name}
*PartitionApi* | [**DoPartitionGet**](docs/PartitionApi.md#dopartitionget) | **Get** /partitions/{partition_name} | Retrieve existing partition {partition_name}
*PartitionApi* | [**DoPartitionMemberDelete**](docs/PartitionApi.md#dopartitionmemberdelete) | **Delete** /partitions/{partition_name}/members/{xname_id} | Delete member from existing partition
*PartitionApi* | [**DoPartitionMembersGet**](docs/PartitionApi.md#dopartitionmembersget) | **Get** /partitions/{partition_name}/members | Retrieve all members of existing partition
*PartitionApi* | [**DoPartitionMembersPost**](docs/PartitionApi.md#dopartitionmemberspost) | **Post** /partitions/{partition_name}/members | Create new member of existing partition (via POST)
*PartitionApi* | [**DoPartitionNamesGet**](docs/PartitionApi.md#dopartitionnamesget) | **Get** /partitions/names | Retrieve all existing partition names
*PartitionApi* | [**DoPartitionPatch**](docs/PartitionApi.md#dopartitionpatch) | **Patch** /partitions/{partition_name} | Update metadata for existing partition {partition_name}
*PartitionApi* | [**DoPartitionsGet**](docs/PartitionApi.md#dopartitionsget) | **Get** /partitions | Retrieve all existing partitions
*PartitionApi* | [**DoPartitionsPost**](docs/PartitionApi.md#dopartitionspost) | **Post** /partitions | Create new partition (via POST)
*PowerMapApi* | [**DoPowerMapDelete**](docs/PowerMapApi.md#dopowermapdelete) | **Delete** /sysinfo/powermaps/{xname} | Delete PowerMap with ID {xname}
*PowerMapApi* | [**DoPowerMapGet**](docs/PowerMapApi.md#dopowermapget) | **Get** /sysinfo/powermaps/{xname} | Retrieve PowerMap at {xname}
*PowerMapApi* | [**DoPowerMapPut**](docs/PowerMapApi.md#dopowermapput) | **Put** /sysinfo/powermaps/{xname} | Update definition for PowerMap ID {xname}
*PowerMapApi* | [**DoPowerMapsDeleteAll**](docs/PowerMapApi.md#dopowermapsdeleteall) | **Delete** /sysinfo/powermaps | Delete all PowerMap entities
*PowerMapApi* | [**DoPowerMapsGet**](docs/PowerMapApi.md#dopowermapsget) | **Get** /sysinfo/powermaps | Retrieve all PowerMaps, returning PowerMapArray
*PowerMapApi* | [**DoPowerMapsPost**](docs/PowerMapApi.md#dopowermapspost) | **Post** /sysinfo/powermaps | Create or Modify PowerMaps
*RedfishEndpointApi* | [**DoRedfishEndpointDelete**](docs/RedfishEndpointApi.md#doredfishendpointdelete) | **Delete** /Inventory/RedfishEndpoints/{xname} | Delete RedfishEndpoint with ID {xname}
*RedfishEndpointApi* | [**DoRedfishEndpointGet**](docs/RedfishEndpointApi.md#doredfishendpointget) | **Get** /Inventory/RedfishEndpoints/{xname} | Retrieve RedfishEndpoint at {xname}
*RedfishEndpointApi* | [**DoRedfishEndpointPatch**](docs/RedfishEndpointApi.md#doredfishendpointpatch) | **Patch** /Inventory/RedfishEndpoints/{xname} | Update (PATCH) definition for RedfishEndpoint ID {xname}
*RedfishEndpointApi* | [**DoRedfishEndpointPut**](docs/RedfishEndpointApi.md#doredfishendpointput) | **Put** /Inventory/RedfishEndpoints/{xname} | Update definition for RedfishEndpoint ID {xname}
*RedfishEndpointApi* | [**DoRedfishEndpointQueryGet**](docs/RedfishEndpointApi.md#doredfishendpointqueryget) | **Get** /Inventory/RedfishEndpoints/Query/{xname} | Retrieve RedfishEndpoint query for {xname}, returning RedfishEndpointArray
*RedfishEndpointApi* | [**DoRedfishEndpointsDeleteAll**](docs/RedfishEndpointApi.md#doredfishendpointsdeleteall) | **Delete** /Inventory/RedfishEndpoints | Delete all RedfishEndpoints
*RedfishEndpointApi* | [**DoRedfishEndpointsGet**](docs/RedfishEndpointApi.md#doredfishendpointsget) | **Get** /Inventory/RedfishEndpoints | Retrieve all RedfishEndpoints, returning RedfishEndpointArray
*RedfishEndpointApi* | [**DoRedfishEndpointsPost**](docs/RedfishEndpointApi.md#doredfishendpointspost) | **Post** /Inventory/RedfishEndpoints | Create RedfishEndpoint(s)
*SCNApi* | [**DoDeleteSCNSubscription**](docs/SCNApi.md#dodeletescnsubscription) | **Delete** /Subscriptions/SCN/{id} | Delete a state change notification subscription
*SCNApi* | [**DoDeleteSCNSubscriptionsAll**](docs/SCNApi.md#dodeletescnsubscriptionsall) | **Delete** /Subscriptions/SCN | Delete all state change notification subscriptions
*SCNApi* | [**DoGetSCNSubscription**](docs/SCNApi.md#dogetscnsubscription) | **Get** /Subscriptions/SCN/{id} | Retrieve a currently-held state change notification subscription
*SCNApi* | [**DoGetSCNSubscriptionsAll**](docs/SCNApi.md#dogetscnsubscriptionsall) | **Get** /Subscriptions/SCN | Retrieve currently-held state change notification subscriptions
*SCNApi* | [**DoPatchSCNSubscription**](docs/SCNApi.md#dopatchscnsubscription) | **Patch** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
*SCNApi* | [**DoPostSCNSubscription**](docs/SCNApi.md#dopostscnsubscription) | **Post** /Subscriptions/SCN | Create a subscription for state change notifications
*SCNApi* | [**DoPutSCNSubscription**](docs/SCNApi.md#doputscnsubscription) | **Put** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
*ServiceEndpointApi* | [**DoServiceEndpointDelete**](docs/ServiceEndpointApi.md#doserviceendpointdelete) | **Delete** /Inventory/ServiceEndpoints/{service}/RedfishEndpoints/{xname} | Delete the {service} ServiceEndpoint managed by {xname}
*ServiceEndpointApi* | [**DoServiceEndpointGet**](docs/ServiceEndpointApi.md#doserviceendpointget) | **Get** /Inventory/ServiceEndpoints/{service}/RedfishEndpoints/{xname} | Retrieve the ServiceEndpoint of a {service} managed by {xname}
*ServiceEndpointApi* | [**DoServiceEndpointsDeleteAll**](docs/ServiceEndpointApi.md#doserviceendpointsdeleteall) | **Delete** /Inventory/ServiceEndpoints | Delete all ServiceEndpoints
*ServiceEndpointApi* | [**DoServiceEndpointsGet**](docs/ServiceEndpointApi.md#doserviceendpointsget) | **Get** /Inventory/ServiceEndpoints/{service} | Retrieve all ServiceEndpoints of a {service}
*ServiceEndpointApi* | [**DoServiceEndpointsGetAll**](docs/ServiceEndpointApi.md#doserviceendpointsgetall) | **Get** /Inventory/ServiceEndpoints | Retrieve ServiceEndpoints Collection
*ServiceInfoApi* | [**DoArchValuesGet**](docs/ServiceInfoApi.md#doarchvaluesget) | **Get** /service/values/arch | Retrieve all valid values for use with the &#x27;arch&#x27; parameter
*ServiceInfoApi* | [**DoClassValuesGet**](docs/ServiceInfoApi.md#doclassvaluesget) | **Get** /service/values/class | Retrieve all valid values for use with the &#x27;class&#x27; parameter
*ServiceInfoApi* | [**DoFlagValuesGet**](docs/ServiceInfoApi.md#doflagvaluesget) | **Get** /service/values/flag | Retrieve all valid values for use with the &#x27;flag&#x27; parameter
*ServiceInfoApi* | [**DoLivenessGet**](docs/ServiceInfoApi.md#dolivenessget) | **Get** /service/liveness | Kubernetes liveness endpoint to monitor service health
*ServiceInfoApi* | [**DoNetTypeValuesGet**](docs/ServiceInfoApi.md#donettypevaluesget) | **Get** /service/values/nettype | Retrieve all valid values for use with the &#x27;nettype&#x27; parameter
*ServiceInfoApi* | [**DoReadyGet**](docs/ServiceInfoApi.md#doreadyget) | **Get** /service/ready | Kubernetes readiness endpoint to monitor service health
*ServiceInfoApi* | [**DoRoleValuesGet**](docs/ServiceInfoApi.md#dorolevaluesget) | **Get** /service/values/role | Retrieve all valid values for use with the &#x27;role&#x27; parameter
*ServiceInfoApi* | [**DoStateValuesGet**](docs/ServiceInfoApi.md#dostatevaluesget) | **Get** /service/values/state | Retrieve all valid values for use with the &#x27;state&#x27; parameter
*ServiceInfoApi* | [**DoSubRoleValuesGet**](docs/ServiceInfoApi.md#dosubrolevaluesget) | **Get** /service/values/subrole | Retrieve all valid values for use with the &#x27;subrole&#x27; parameter
*ServiceInfoApi* | [**DoTypeValuesGet**](docs/ServiceInfoApi.md#dotypevaluesget) | **Get** /service/values/type | Retrieve all valid values for use with the &#x27;type&#x27; parameter
*ServiceInfoApi* | [**DoValuesGet**](docs/ServiceInfoApi.md#dovaluesget) | **Get** /service/values | Retrieve all valid values for use as parameters
*ServiceReservationsApi* | [**LocksServiceReservationsCheckPost**](docs/ServiceReservationsApi.md#locksservicereservationscheckpost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
*ServiceReservationsApi* | [**LocksServiceReservationsPost**](docs/ServiceReservationsApi.md#locksservicereservationspost) | **Post** /locks/service/reservations | Create reservations
*ServiceReservationsApi* | [**LocksServiceReservationsReleasePost**](docs/ServiceReservationsApi.md#locksservicereservationsreleasepost) | **Post** /locks/service/reservations/release | Releases existing reservations.
*ServiceReservationsApi* | [**LocksServiceReservationsRenewPost**](docs/ServiceReservationsApi.md#locksservicereservationsrenewpost) | **Post** /locks/service/reservations/renew | Renew existing reservations.

## Documentation For Models

 - [Actions100ChassisActions](docs/Actions100ChassisActions.md)
 - [Actions100ChassisActionsChassisReset](docs/Actions100ChassisActionsChassisReset.md)
 - [Actions100ComputerSystemActions](docs/Actions100ComputerSystemActions.md)
 - [Actions100ComputerSystemActionsComputerSystemReset](docs/Actions100ComputerSystemActionsComputerSystemReset.md)
 - [Actions100ManagerActions](docs/Actions100ManagerActions.md)
 - [Actions100ManagerActionsManagerReset](docs/Actions100ManagerActionsManagerReset.md)
 - [Actions100OutletActions](docs/Actions100OutletActions.md)
 - [Actions100OutletActionsOutletPowerControl](docs/Actions100OutletActionsOutletPowerControl.md)
 - [Actions100OutletActionsOutletResetBreaker](docs/Actions100OutletActionsOutletResetBreaker.md)
 - [Actions100OutletActionsOutletResetStatistics](docs/Actions100OutletActionsOutletResetStatistics.md)
 - [AdminLock100](docs/AdminLock100.md)
 - [AdminReservationCreate100](docs/AdminReservationCreate100.md)
 - [AdminReservationCreateResponse100](docs/AdminReservationCreateResponse100.md)
 - [AdminReservationRemove100](docs/AdminReservationRemove100.md)
 - [AdminStatusCheckResponse100](docs/AdminStatusCheckResponse100.md)
 - [CompEthInterface100](docs/CompEthInterface100.md)
 - [CompEthInterface100IpAddressMapping](docs/CompEthInterface100IpAddressMapping.md)
 - [CompEthInterface100IpAddressMappingPatch](docs/CompEthInterface100IpAddressMappingPatch.md)
 - [CompEthInterface100Patch](docs/CompEthInterface100Patch.md)
 - [Component100Component](docs/Component100Component.md)
 - [Component100ComponentCreate](docs/Component100ComponentCreate.md)
 - [Component100PatchArrayItemNid](docs/Component100PatchArrayItemNid.md)
 - [Component100PatchEnabled](docs/Component100PatchEnabled.md)
 - [Component100PatchFlagOnly](docs/Component100PatchFlagOnly.md)
 - [Component100PatchNid](docs/Component100PatchNid.md)
 - [Component100PatchRole](docs/Component100PatchRole.md)
 - [Component100PatchSoftwareStatus](docs/Component100PatchSoftwareStatus.md)
 - [Component100PatchStateData](docs/Component100PatchStateData.md)
 - [Component100Put](docs/Component100Put.md)
 - [Component100ResourceUriCollection](docs/Component100ResourceUriCollection.md)
 - [ComponentArrayComponentArray](docs/ComponentArrayComponentArray.md)
 - [ComponentArrayPatchArrayEnabled](docs/ComponentArrayPatchArrayEnabled.md)
 - [ComponentArrayPatchArrayFlagOnly](docs/ComponentArrayPatchArrayFlagOnly.md)
 - [ComponentArrayPatchArrayNid](docs/ComponentArrayPatchArrayNid.md)
 - [ComponentArrayPatchArrayRole](docs/ComponentArrayPatchArrayRole.md)
 - [ComponentArrayPatchArraySoftwareStatus](docs/ComponentArrayPatchArraySoftwareStatus.md)
 - [ComponentArrayPatchArrayStateData](docs/ComponentArrayPatchArrayStateData.md)
 - [ComponentArrayPostArray](docs/ComponentArrayPostArray.md)
 - [ComponentArrayPostByNidQuery](docs/ComponentArrayPostByNidQuery.md)
 - [ComponentArrayPostQuery](docs/ComponentArrayPostQuery.md)
 - [ComponentByNid100ResourceUriCollection](docs/ComponentByNid100ResourceUriCollection.md)
 - [ComponentEndpoint100ComponentEndpoint](docs/ComponentEndpoint100ComponentEndpoint.md)
 - [ComponentEndpoint100RedfishChassisInfo](docs/ComponentEndpoint100RedfishChassisInfo.md)
 - [ComponentEndpoint100RedfishManagerInfo](docs/ComponentEndpoint100RedfishManagerInfo.md)
 - [ComponentEndpoint100RedfishOutletInfo](docs/ComponentEndpoint100RedfishOutletInfo.md)
 - [ComponentEndpoint100RedfishPowerDistributionInfo](docs/ComponentEndpoint100RedfishPowerDistributionInfo.md)
 - [ComponentEndpoint100RedfishSystemInfo](docs/ComponentEndpoint100RedfishSystemInfo.md)
 - [ComponentEndpoint100ResourceUriCollection](docs/ComponentEndpoint100ResourceUriCollection.md)
 - [ComponentEndpointArrayComponentEndpointArray](docs/ComponentEndpointArrayComponentEndpointArray.md)
 - [ComponentEndpointArrayPostQuery](docs/ComponentEndpointArrayPostQuery.md)
 - [ComponentEndpointChassis](docs/ComponentEndpointChassis.md)
 - [ComponentEndpointComputerSystem](docs/ComponentEndpointComputerSystem.md)
 - [ComponentEndpointManager](docs/ComponentEndpointManager.md)
 - [ComponentEndpointOutlet](docs/ComponentEndpointOutlet.md)
 - [ComponentEndpointPowerDistribution](docs/ComponentEndpointPowerDistribution.md)
 - [ComponentStatus100](docs/ComponentStatus100.md)
 - [Counts100](docs/Counts100.md)
 - [DeputyKeys100](docs/DeputyKeys100.md)
 - [Discover100DiscoverInput](docs/Discover100DiscoverInput.md)
 - [DiscoveryStatus100Details](docs/DiscoveryStatus100Details.md)
 - [DiscoveryStatus100DiscoveryStatus](docs/DiscoveryStatus100DiscoveryStatus.md)
 - [EthernetNicInfo100](docs/EthernetNicInfo100.md)
 - [FailedXnames100](docs/FailedXnames100.md)
 - [Group100](docs/Group100.md)
 - [Group100Patch](docs/Group100Patch.md)
 - [HmsArch100](docs/HmsArch100.md)
 - [HmsClass100](docs/HmsClass100.md)
 - [HmsFlag100](docs/HmsFlag100.md)
 - [HmsState100](docs/HmsState100.md)
 - [HmsType100](docs/HmsType100.md)
 - [HsnInfo100](docs/HsnInfo100.md)
 - [HsnInfoEntry100](docs/HsnInfoEntry100.md)
 - [HwInvByFruCabinet](docs/HwInvByFruCabinet.md)
 - [HwInvByFruChassis](docs/HwInvByFruChassis.md)
 - [HwInvByFruComputeModule](docs/HwInvByFruComputeModule.md)
 - [HwInvByFruDrive](docs/HwInvByFruDrive.md)
 - [HwInvByFruMemory](docs/HwInvByFruMemory.md)
 - [HwInvByFruMgmtHlSwitch](docs/HwInvByFruMgmtHlSwitch.md)
 - [HwInvByFruMgmtSwitch](docs/HwInvByFruMgmtSwitch.md)
 - [HwInvByFruNode](docs/HwInvByFruNode.md)
 - [HwInvByFruNodeAccel](docs/HwInvByFruNodeAccel.md)
 - [HwInvByFruNodeAccelRiser](docs/HwInvByFruNodeAccelRiser.md)
 - [HwInvByFruNodeBmc](docs/HwInvByFruNodeBmc.md)
 - [HwInvByFruNodeEnclosure](docs/HwInvByFruNodeEnclosure.md)
 - [HwInvByFruNodeEnclosurePowerSupply](docs/HwInvByFruNodeEnclosurePowerSupply.md)
 - [HwInvByFruOutlet](docs/HwInvByFruOutlet.md)
 - [HwInvByFruProcessor](docs/HwInvByFruProcessor.md)
 - [HwInvByFruRouterBmc](docs/HwInvByFruRouterBmc.md)
 - [HwInvByFruRouterModule](docs/HwInvByFruRouterModule.md)
 - [HwInvByFrucduMgmtSwitch](docs/HwInvByFrucduMgmtSwitch.md)
 - [HwInvByFrucmmRectifier](docs/HwInvByFrucmmRectifier.md)
 - [HwInvByFruhsnBoard](docs/HwInvByFruhsnBoard.md)
 - [HwInvByFruhsnnic](docs/HwInvByFruhsnnic.md)
 - [HwInvByFrupdu](docs/HwInvByFrupdu.md)
 - [HwInvByLocCabinet](docs/HwInvByLocCabinet.md)
 - [HwInvByLocCduMgmtSwitch](docs/HwInvByLocCduMgmtSwitch.md)
 - [HwInvByLocChassis](docs/HwInvByLocChassis.md)
 - [HwInvByLocCmmRectifier](docs/HwInvByLocCmmRectifier.md)
 - [HwInvByLocComputeModule](docs/HwInvByLocComputeModule.md)
 - [HwInvByLocDrive](docs/HwInvByLocDrive.md)
 - [HwInvByLocHsnBoard](docs/HwInvByLocHsnBoard.md)
 - [HwInvByLocHsnnic](docs/HwInvByLocHsnnic.md)
 - [HwInvByLocMemory](docs/HwInvByLocMemory.md)
 - [HwInvByLocMgmtHlSwitch](docs/HwInvByLocMgmtHlSwitch.md)
 - [HwInvByLocMgmtSwitch](docs/HwInvByLocMgmtSwitch.md)
 - [HwInvByLocNode](docs/HwInvByLocNode.md)
 - [HwInvByLocNodeAccel](docs/HwInvByLocNodeAccel.md)
 - [HwInvByLocNodeAccelRiser](docs/HwInvByLocNodeAccelRiser.md)
 - [HwInvByLocNodeBmc](docs/HwInvByLocNodeBmc.md)
 - [HwInvByLocNodeEnclosure](docs/HwInvByLocNodeEnclosure.md)
 - [HwInvByLocNodeEnclosurePowerSupply](docs/HwInvByLocNodeEnclosurePowerSupply.md)
 - [HwInvByLocOutlet](docs/HwInvByLocOutlet.md)
 - [HwInvByLocPdu](docs/HwInvByLocPdu.md)
 - [HwInvByLocProcessor](docs/HwInvByLocProcessor.md)
 - [HwInvByLocRouterBmc](docs/HwInvByLocRouterBmc.md)
 - [HwInvByLocRouterModule](docs/HwInvByLocRouterModule.md)
 - [HwInventory100HsnnicLocationInfo](docs/HwInventory100HsnnicLocationInfo.md)
 - [HwInventory100HsnnicfruInfo](docs/HwInventory100HsnnicfruInfo.md)
 - [HwInventory100HwInventory](docs/HwInventory100HwInventory.md)
 - [HwInventory100HwInventoryByFru](docs/HwInventory100HwInventoryByFru.md)
 - [HwInventory100HwInventoryByLocation](docs/HwInventory100HwInventoryByLocation.md)
 - [HwInventory100HwInventoryHistory](docs/HwInventory100HwInventoryHistory.md)
 - [HwInventory100HwInventoryHistoryArray](docs/HwInventory100HwInventoryHistoryArray.md)
 - [HwInventory100HwInventoryHistoryCollection](docs/HwInventory100HwInventoryHistoryCollection.md)
 - [HwInventory100RedfishChassisFruInfo](docs/HwInventory100RedfishChassisFruInfo.md)
 - [HwInventory100RedfishChassisLocationInfo](docs/HwInventory100RedfishChassisLocationInfo.md)
 - [HwInventory100RedfishCmmRectifierFruInfo](docs/HwInventory100RedfishCmmRectifierFruInfo.md)
 - [HwInventory100RedfishCmmRectifierLocationInfo](docs/HwInventory100RedfishCmmRectifierLocationInfo.md)
 - [HwInventory100RedfishDriveFruInfo](docs/HwInventory100RedfishDriveFruInfo.md)
 - [HwInventory100RedfishDriveLocationInfo](docs/HwInventory100RedfishDriveLocationInfo.md)
 - [HwInventory100RedfishManagerFruInfo](docs/HwInventory100RedfishManagerFruInfo.md)
 - [HwInventory100RedfishManagerLocationInfo](docs/HwInventory100RedfishManagerLocationInfo.md)
 - [HwInventory100RedfishMemoryFruInfo](docs/HwInventory100RedfishMemoryFruInfo.md)
 - [HwInventory100RedfishMemoryLocationInfo](docs/HwInventory100RedfishMemoryLocationInfo.md)
 - [HwInventory100RedfishMemoryLocationInfoMemoryLocation](docs/HwInventory100RedfishMemoryLocationInfoMemoryLocation.md)
 - [HwInventory100RedfishNodeAccelRiserFruInfo](docs/HwInventory100RedfishNodeAccelRiserFruInfo.md)
 - [HwInventory100RedfishNodeAccelRiserLocationInfo](docs/HwInventory100RedfishNodeAccelRiserLocationInfo.md)
 - [HwInventory100RedfishNodeEnclosurePowerSupplyFruInfo](docs/HwInventory100RedfishNodeEnclosurePowerSupplyFruInfo.md)
 - [HwInventory100RedfishNodeEnclosurePowerSupplyLocationInfo](docs/HwInventory100RedfishNodeEnclosurePowerSupplyLocationInfo.md)
 - [HwInventory100RedfishOutletFruInfo](docs/HwInventory100RedfishOutletFruInfo.md)
 - [HwInventory100RedfishOutletLocationInfo](docs/HwInventory100RedfishOutletLocationInfo.md)
 - [HwInventory100RedfishPduLocationInfo](docs/HwInventory100RedfishPduLocationInfo.md)
 - [HwInventory100RedfishPdufruInfo](docs/HwInventory100RedfishPdufruInfo.md)
 - [HwInventory100RedfishPdufruInfoCircuitSummary](docs/HwInventory100RedfishPdufruInfoCircuitSummary.md)
 - [HwInventory100RedfishProcessorFruInfo](docs/HwInventory100RedfishProcessorFruInfo.md)
 - [HwInventory100RedfishProcessorFruInfoProcessorId](docs/HwInventory100RedfishProcessorFruInfoProcessorId.md)
 - [HwInventory100RedfishProcessorLocationInfo](docs/HwInventory100RedfishProcessorLocationInfo.md)
 - [HwInventory100RedfishSystemFruInfo](docs/HwInventory100RedfishSystemFruInfo.md)
 - [HwInventory100RedfishSystemLocationInfo](docs/HwInventory100RedfishSystemLocationInfo.md)
 - [HwInventory100RedfishSystemLocationInfoMemorySummary](docs/HwInventory100RedfishSystemLocationInfoMemorySummary.md)
 - [HwInventory100RedfishSystemLocationInfoProcessorSummary](docs/HwInventory100RedfishSystemLocationInfoProcessorSummary.md)
 - [InventoryHardwareBody](docs/InventoryHardwareBody.md)
 - [Lock100](docs/Lock100.md)
 - [Lock100Patch](docs/Lock100Patch.md)
 - [MemberId](docs/MemberId.md)
 - [Members100](docs/Members100.md)
 - [Membership100](docs/Membership100.md)
 - [Message100ExtendedInfo](docs/Message100ExtendedInfo.md)
 - [NetType100](docs/NetType100.md)
 - [NodeMap100NodeMap](docs/NodeMap100NodeMap.md)
 - [NodeMap100PostNodeMap](docs/NodeMap100PostNodeMap.md)
 - [NodeMapArrayNodeMapArray](docs/NodeMapArrayNodeMapArray.md)
 - [Partition100](docs/Partition100.md)
 - [Partition100Patch](docs/Partition100Patch.md)
 - [PowerControl100](docs/PowerControl100.md)
 - [PowerControl100Oem](docs/PowerControl100Oem.md)
 - [PowerControl100OemCray](docs/PowerControl100OemCray.md)
 - [PowerControl100OemCrayPowerLimit](docs/PowerControl100OemCrayPowerLimit.md)
 - [PowerControl100RelatedItem](docs/PowerControl100RelatedItem.md)
 - [PowerMap100PostPowerMap](docs/PowerMap100PostPowerMap.md)
 - [PowerMap100PowerMap](docs/PowerMap100PowerMap.md)
 - [Problem7807](docs/Problem7807.md)
 - [RedfishEndpoint100RedfishEndpoint](docs/RedfishEndpoint100RedfishEndpoint.md)
 - [RedfishEndpoint100RedfishEndpointDiscoveryInfo](docs/RedfishEndpoint100RedfishEndpointDiscoveryInfo.md)
 - [RedfishEndpoint100ResourceUriCollection](docs/RedfishEndpoint100ResourceUriCollection.md)
 - [RedfishEndpointArrayPostQuery](docs/RedfishEndpointArrayPostQuery.md)
 - [RedfishEndpointArrayRedfishEndpointArray](docs/RedfishEndpointArrayRedfishEndpointArray.md)
 - [RedfishSubtype100](docs/RedfishSubtype100.md)
 - [RedfishType100](docs/RedfishType100.md)
 - [ReservedKeys100](docs/ReservedKeys100.md)
 - [ReservedKeysWithRenewal100](docs/ReservedKeysWithRenewal100.md)
 - [ResourceUri100](docs/ResourceUri100.md)
 - [ResourceUriCollectionResourceUriCollection](docs/ResourceUriCollectionResourceUriCollection.md)
 - [Response100](docs/Response100.md)
 - [ServiceEndpoint100ServiceEndpoint](docs/ServiceEndpoint100ServiceEndpoint.md)
 - [ServiceEndpoint100ServiceInfo](docs/ServiceEndpoint100ServiceInfo.md)
 - [ServiceEndpointArrayServiceEndpointArray](docs/ServiceEndpointArrayServiceEndpointArray.md)
 - [ServiceReservationCheckResponse100](docs/ServiceReservationCheckResponse100.md)
 - [ServiceReservationCreate100](docs/ServiceReservationCreate100.md)
 - [ServiceReservationCreateResponse100](docs/ServiceReservationCreateResponse100.md)
 - [SubscriptionsScnPatchSubscription](docs/SubscriptionsScnPatchSubscription.md)
 - [SubscriptionsScnPostSubscription](docs/SubscriptionsScnPostSubscription.md)
 - [SubscriptionsScnSubscriptionArray](docs/SubscriptionsScnSubscriptionArray.md)
 - [SubscriptionsScnSubscriptionArrayItem100](docs/SubscriptionsScnSubscriptionArrayItem100.md)
 - [Values100ArchArray](docs/Values100ArchArray.md)
 - [Values100ClassArray](docs/Values100ClassArray.md)
 - [Values100FlagArray](docs/Values100FlagArray.md)
 - [Values100NetTypeArray](docs/Values100NetTypeArray.md)
 - [Values100RoleArray](docs/Values100RoleArray.md)
 - [Values100StateArray](docs/Values100StateArray.md)
 - [Values100SubRoleArray](docs/Values100SubRoleArray.md)
 - [Values100TypeArray](docs/Values100TypeArray.md)
 - [Values100Values](docs/Values100Values.md)
 - [XnameKeys100](docs/XnameKeys100.md)
 - [XnameKeysDeputyExpire100](docs/XnameKeysDeputyExpire100.md)
 - [XnameKeysNoExpire100](docs/XnameKeysNoExpire100.md)
 - [XnameResponse100](docs/XnameResponse100.md)
 - [XnameWithKey100](docs/XnameWithKey100.md)
 - [Xnames](docs/Xnames.md)

## Documentation For Authorization
 Endpoints do not require authorization.


## Author


