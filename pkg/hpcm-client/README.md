# Go API client for hpcm_client

HPE Performance Cluster Manager 'cmdb' service features a REST API. This section describes its implementation.  Standard REST API concepts (such as HTTP verbs, return codes, JSON, etc.) are not covered here.

## Overview
This API client was generated by the [swagger-codegen](https://github.com/swagger-api/swagger-codegen) project.  By using the [swagger-spec](https://github.com/swagger-api/swagger-spec) from a remote server, you can easily generate an API client.

- API version: v1
- Package version: 1.0.0
- Build package: io.swagger.codegen.v3.generators.go.GoClientCodegen

## Installation
Put the package under your project folder and add the following in import:
```golang
import "./hpcm_client"
```

## Documentation for API Endpoints

All URIs are relative to *https://localhost:8080/cmu/v1*

Class | Method | HTTP request | Description
------------ | ------------- | ------------- | -------------
*AdministrationOperationsApi* | [**Backup**](docs/AdministrationOperationsApi.md#backup) | **Post** /admin/db/backup | Trigger a database backup on disk
*AlertOperationsApi* | [**Get**](docs/AlertOperationsApi.md#get) | **Get** /alerts/{identifier} | Gets one or more alert(s)
*AlertOperationsApi* | [**GetAll**](docs/AlertOperationsApi.md#getall) | **Get** /alerts | Lists all alerts
*AlertOperationsApi* | [**GetAttributes**](docs/AlertOperationsApi.md#getattributes) | **Get** /alerts/{identifier}/attributes | Gets all attributes of a single alert
*ApplicationApi* | [**Get**](docs/ApplicationApi.md#get) | **Get** / | Lists application entry points
*ApplicationApi* | [**GetSettings**](docs/ApplicationApi.md#getsettings) | **Get** /settings | Lists application current settings
*ArchitectureOperationsApi* | [**Get**](docs/ArchitectureOperationsApi.md#get) | **Get** /architectures/{identifier} | Gets one or more architecture(s)
*ArchitectureOperationsApi* | [**GetAll**](docs/ArchitectureOperationsApi.md#getall) | **Get** /architectures | Lists all architectures
*ArchitectureOperationsApi* | [**GetAttributes**](docs/ArchitectureOperationsApi.md#getattributes) | **Get** /architectures/{identifier}/attributes | Gets all attributes of a single architecture
*ControllerOperationsApi* | [**AddAll**](docs/ControllerOperationsApi.md#addall) | **Post** /controllers | Creates one or multiple new controllers
*ControllerOperationsApi* | [**AddNodes**](docs/ControllerOperationsApi.md#addnodes) | **Post** /controllers/{identifier}/nodes | Adds nodes to an existing group
*ControllerOperationsApi* | [**Delete**](docs/ControllerOperationsApi.md#delete) | **Delete** /controllers/{identifier} | Deletes an existing controller
*ControllerOperationsApi* | [**DeleteAll**](docs/ControllerOperationsApi.md#deleteall) | **Delete** /controllers | Deletes a set of existing controllers
*ControllerOperationsApi* | [**DeleteAttributes**](docs/ControllerOperationsApi.md#deleteattributes) | **Delete** /controllers/{identifier}/attributes | Removes all attributes of an existing group
*ControllerOperationsApi* | [**DeleteGlobalAttribute**](docs/ControllerOperationsApi.md#deleteglobalattribute) | **Delete** /controllers/attributes/{label} | Deletes a global attribute defined for groups
*ControllerOperationsApi* | [**Get**](docs/ControllerOperationsApi.md#get) | **Get** /controllers/{identifier} | Gets one or more group(s)
*ControllerOperationsApi* | [**GetAll**](docs/ControllerOperationsApi.md#getall) | **Get** /controllers | Lists all groups
*ControllerOperationsApi* | [**GetAttribute**](docs/ControllerOperationsApi.md#getattribute) | **Get** /controllers/attributes/{label} | Gets a global attribute defined for groups
*ControllerOperationsApi* | [**GetAttributes**](docs/ControllerOperationsApi.md#getattributes) | **Get** /controllers/{identifier}/attributes | Gets all attributes of a single group
*ControllerOperationsApi* | [**GetAvailableAction**](docs/ControllerOperationsApi.md#getavailableaction) | **Get** /controllers/{identifier}/actions | Gets list of available actions on an existing group
*ControllerOperationsApi* | [**GetGlobalAttributes**](docs/ControllerOperationsApi.md#getglobalattributes) | **Get** /controllers/attributes | Gets all global attributes defined for groups
*ControllerOperationsApi* | [**GetNode**](docs/ControllerOperationsApi.md#getnode) | **Get** /controllers/{identifier}/nodes/{node_id} | Gets one node of an existing group
*ControllerOperationsApi* | [**GetNodes**](docs/ControllerOperationsApi.md#getnodes) | **Get** /controllers/{identifier}/nodes | Gets all nodes of an existing group
*ControllerOperationsApi* | [**Put**](docs/ControllerOperationsApi.md#put) | **Put** /controllers/{identifier} | Updates an existing controller
*ControllerOperationsApi* | [**PutAll**](docs/ControllerOperationsApi.md#putall) | **Put** /controllers | Updates a set of existing controllers
*ControllerOperationsApi* | [**PutAttributes**](docs/ControllerOperationsApi.md#putattributes) | **Put** /controllers/{identifier}/attributes | Adds or modifies attributes of an existing group
*ControllerOperationsApi* | [**PutGlobalAttributes**](docs/ControllerOperationsApi.md#putglobalattributes) | **Put** /controllers/attributes | Adds or modifies global attributes for groups
*ControllerOperationsApi* | [**RemoveNode**](docs/ControllerOperationsApi.md#removenode) | **Delete** /controllers/{identifier}/nodes/{node_id} | Removes one node from an existing group
*ControllerOperationsApi* | [**RemoveNodes**](docs/ControllerOperationsApi.md#removenodes) | **Delete** /controllers/{identifier}/nodes | Removes some or all nodes from an existing group
*ControllerOperationsApi* | [**RunAction**](docs/ControllerOperationsApi.md#runaction) | **Post** /controllers/{identifier}/actions/{action} | Runs an action on a set of existing groups
*CustomGroupOperationsApi* | [**AddAll**](docs/CustomGroupOperationsApi.md#addall) | **Post** /customgroups | Creates one or multiple new custom group(s)
*CustomGroupOperationsApi* | [**AddNodes**](docs/CustomGroupOperationsApi.md#addnodes) | **Post** /customgroups/{identifier}/nodes | Adds nodes to an existing group
*CustomGroupOperationsApi* | [**Delete**](docs/CustomGroupOperationsApi.md#delete) | **Delete** /customgroups/{identifier} | Deletes or archive an existing custom group
*CustomGroupOperationsApi* | [**DeleteAll**](docs/CustomGroupOperationsApi.md#deleteall) | **Delete** /customgroups | Deletes or archive a set of existing custom groups
*CustomGroupOperationsApi* | [**DeleteAttributes**](docs/CustomGroupOperationsApi.md#deleteattributes) | **Delete** /customgroups/{identifier}/attributes | Removes all attributes of an existing group
*CustomGroupOperationsApi* | [**DeleteGlobalAttribute**](docs/CustomGroupOperationsApi.md#deleteglobalattribute) | **Delete** /customgroups/attributes/{label} | Deletes a global attribute defined for groups
*CustomGroupOperationsApi* | [**Get**](docs/CustomGroupOperationsApi.md#get) | **Get** /customgroups/{identifier} | Gets one or more group(s)
*CustomGroupOperationsApi* | [**GetAll**](docs/CustomGroupOperationsApi.md#getall) | **Get** /customgroups | Lists all groups
*CustomGroupOperationsApi* | [**GetAttribute**](docs/CustomGroupOperationsApi.md#getattribute) | **Get** /customgroups/attributes/{label} | Gets a global attribute defined for groups
*CustomGroupOperationsApi* | [**GetAttributes**](docs/CustomGroupOperationsApi.md#getattributes) | **Get** /customgroups/{identifier}/attributes | Gets all attributes of a single group
*CustomGroupOperationsApi* | [**GetAvailableAction**](docs/CustomGroupOperationsApi.md#getavailableaction) | **Get** /customgroups/{identifier}/actions | Gets list of available actions on an existing group
*CustomGroupOperationsApi* | [**GetGlobalAttributes**](docs/CustomGroupOperationsApi.md#getglobalattributes) | **Get** /customgroups/attributes | Gets all global attributes defined for groups
*CustomGroupOperationsApi* | [**GetNode**](docs/CustomGroupOperationsApi.md#getnode) | **Get** /customgroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
*CustomGroupOperationsApi* | [**GetNodes**](docs/CustomGroupOperationsApi.md#getnodes) | **Get** /customgroups/{identifier}/nodes | Gets all nodes of an existing group
*CustomGroupOperationsApi* | [**Put**](docs/CustomGroupOperationsApi.md#put) | **Put** /customgroups/{identifier} | 
*CustomGroupOperationsApi* | [**PutAll**](docs/CustomGroupOperationsApi.md#putall) | **Put** /customgroups | 
*CustomGroupOperationsApi* | [**PutAttributes**](docs/CustomGroupOperationsApi.md#putattributes) | **Put** /customgroups/{identifier}/attributes | Adds or modifies attributes of an existing group
*CustomGroupOperationsApi* | [**PutGlobalAttributes**](docs/CustomGroupOperationsApi.md#putglobalattributes) | **Put** /customgroups/attributes | Adds or modifies global attributes for groups
*CustomGroupOperationsApi* | [**RemoveNode**](docs/CustomGroupOperationsApi.md#removenode) | **Delete** /customgroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
*CustomGroupOperationsApi* | [**RemoveNodes**](docs/CustomGroupOperationsApi.md#removenodes) | **Delete** /customgroups/{identifier}/nodes | Removes some or all nodes from an existing group
*CustomGroupOperationsApi* | [**RunAction**](docs/CustomGroupOperationsApi.md#runaction) | **Post** /customgroups/{identifier}/actions/{action} | Runs an action on a set of existing groups
*DefaultApi* | [**GetExternalGrammar**](docs/DefaultApi.md#getexternalgrammar) | **Get** /application.wadl/{path} | 
*DefaultApi* | [**GetWadl**](docs/DefaultApi.md#getwadl) | **Get** /application.wadl | 
*EventHookOperationsApi* | [**AddAll**](docs/EventHookOperationsApi.md#addall) | **Post** /eventhooks | Creates one or multiple new event hooks
*EventHookOperationsApi* | [**Delete**](docs/EventHookOperationsApi.md#delete) | **Delete** /eventhooks/{identifier} | Deletes an existing event hook
*EventHookOperationsApi* | [**DeleteAll**](docs/EventHookOperationsApi.md#deleteall) | **Delete** /eventhooks | Deletes a set of existing event hooks
*EventHookOperationsApi* | [**DeleteAttributes**](docs/EventHookOperationsApi.md#deleteattributes) | **Delete** /eventhooks/{identifier}/attributes | Remove all attributes of an existing event hook
*EventHookOperationsApi* | [**DeleteGlobalAttribute**](docs/EventHookOperationsApi.md#deleteglobalattribute) | **Delete** /eventhooks/attributes/{label} | Deletes a global attribute defined for event hooks
*EventHookOperationsApi* | [**Get**](docs/EventHookOperationsApi.md#get) | **Get** /eventhooks/{identifier} | Gets one or more event hook(s)
*EventHookOperationsApi* | [**GetAll**](docs/EventHookOperationsApi.md#getall) | **Get** /eventhooks | Lists all event hooks
*EventHookOperationsApi* | [**GetAttributes**](docs/EventHookOperationsApi.md#getattributes) | **Get** /eventhooks/{identifier}/attributes | Gets all attributes of a single event hook
*EventHookOperationsApi* | [**GetGlobalAttribute**](docs/EventHookOperationsApi.md#getglobalattribute) | **Get** /eventhooks/attributes/{label} | Gets a global attribute defined for event hooks
*EventHookOperationsApi* | [**GetGlobalAttributes**](docs/EventHookOperationsApi.md#getglobalattributes) | **Get** /eventhooks/attributes | Gets all global attributes defined for event hooks
*EventHookOperationsApi* | [**Put**](docs/EventHookOperationsApi.md#put) | **Put** /eventhooks/{identifier} | Updates an existing event hook
*EventHookOperationsApi* | [**PutAll**](docs/EventHookOperationsApi.md#putall) | **Put** /eventhooks | Updates a set of existing event hooks
*EventHookOperationsApi* | [**PutAttributes**](docs/EventHookOperationsApi.md#putattributes) | **Put** /eventhooks/{identifier}/attributes | Adds or modifies attributes of an existing event hook
*EventHookOperationsApi* | [**PutGlobalAttributes**](docs/EventHookOperationsApi.md#putglobalattributes) | **Put** /eventhooks/attributes | Adds or modifies global attributes for event hooks
*EventsOperationsApi* | [**GetAll**](docs/EventsOperationsApi.md#getall) | **Get** /events | 
*EventsOperationsApi* | [**Summary**](docs/EventsOperationsApi.md#summary) | **Get** /events/summary | 
*ImageGroupOperationsApi* | [**AddAll**](docs/ImageGroupOperationsApi.md#addall) | **Post** /imagegroups | Creates one or multiple new image group(s)
*ImageGroupOperationsApi* | [**AddNodes**](docs/ImageGroupOperationsApi.md#addnodes) | **Post** /imagegroups/{identifier}/nodes | Adds nodes to an existing group
*ImageGroupOperationsApi* | [**Delete**](docs/ImageGroupOperationsApi.md#delete) | **Delete** /imagegroups/{identifier} | Deletes an existing image group
*ImageGroupOperationsApi* | [**DeleteAll**](docs/ImageGroupOperationsApi.md#deleteall) | **Delete** /imagegroups | Deletes a set of existing image groups
*ImageGroupOperationsApi* | [**DeleteAttributes**](docs/ImageGroupOperationsApi.md#deleteattributes) | **Delete** /imagegroups/{identifier}/attributes | Removes all attributes of an existing group
*ImageGroupOperationsApi* | [**DeleteGlobalAttribute**](docs/ImageGroupOperationsApi.md#deleteglobalattribute) | **Delete** /imagegroups/attributes/{label} | Deletes a global attribute defined for groups
*ImageGroupOperationsApi* | [**Get**](docs/ImageGroupOperationsApi.md#get) | **Get** /imagegroups/{identifier} | Gets one or more group(s)
*ImageGroupOperationsApi* | [**GetAll**](docs/ImageGroupOperationsApi.md#getall) | **Get** /imagegroups | Lists all groups
*ImageGroupOperationsApi* | [**GetAttribute**](docs/ImageGroupOperationsApi.md#getattribute) | **Get** /imagegroups/attributes/{label} | Gets a global attribute defined for groups
*ImageGroupOperationsApi* | [**GetAttributes**](docs/ImageGroupOperationsApi.md#getattributes) | **Get** /imagegroups/{identifier}/attributes | Gets all attributes of a single group
*ImageGroupOperationsApi* | [**GetAvailableAction**](docs/ImageGroupOperationsApi.md#getavailableaction) | **Get** /imagegroups/{identifier}/actions | Gets list of available actions on an existing group
*ImageGroupOperationsApi* | [**GetGlobalAttributes**](docs/ImageGroupOperationsApi.md#getglobalattributes) | **Get** /imagegroups/attributes | Gets all global attributes defined for groups
*ImageGroupOperationsApi* | [**GetNode**](docs/ImageGroupOperationsApi.md#getnode) | **Get** /imagegroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
*ImageGroupOperationsApi* | [**GetNodes**](docs/ImageGroupOperationsApi.md#getnodes) | **Get** /imagegroups/{identifier}/nodes | Gets all nodes of an existing group
*ImageGroupOperationsApi* | [**Put**](docs/ImageGroupOperationsApi.md#put) | **Put** /imagegroups/{identifier} | 
*ImageGroupOperationsApi* | [**PutAll**](docs/ImageGroupOperationsApi.md#putall) | **Put** /imagegroups | 
*ImageGroupOperationsApi* | [**PutAttributes**](docs/ImageGroupOperationsApi.md#putattributes) | **Put** /imagegroups/{identifier}/attributes | Adds or modifies attributes of an existing group
*ImageGroupOperationsApi* | [**PutGlobalAttributes**](docs/ImageGroupOperationsApi.md#putglobalattributes) | **Put** /imagegroups/attributes | Adds or modifies global attributes for groups
*ImageGroupOperationsApi* | [**RemoveNode**](docs/ImageGroupOperationsApi.md#removenode) | **Delete** /imagegroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
*ImageGroupOperationsApi* | [**RemoveNodes**](docs/ImageGroupOperationsApi.md#removenodes) | **Delete** /imagegroups/{identifier}/nodes | Removes some or all nodes from an existing group
*ImageGroupOperationsApi* | [**RunAction**](docs/ImageGroupOperationsApi.md#runaction) | **Post** /imagegroups/{identifier}/actions/{action} | Runs an action on a set of existing groups
*ManagementCardOperationsApi* | [**Get**](docs/ManagementCardOperationsApi.md#get) | **Get** /managementcards/{identifier} | Gets one or more management card(s)
*ManagementCardOperationsApi* | [**GetAll**](docs/ManagementCardOperationsApi.md#getall) | **Get** /managementcards | Lists all management cards
*ManagementCardOperationsApi* | [**GetAttributes**](docs/ManagementCardOperationsApi.md#getattributes) | **Get** /managementcards/{identifier}/attributes | Gets all attributes of a single management card
*MetricOperationsApi* | [**Get**](docs/MetricOperationsApi.md#get) | **Get** /metrics/{identifier} | Gets one or more metric(s)
*MetricOperationsApi* | [**GetAll**](docs/MetricOperationsApi.md#getall) | **Get** /metrics | Lists all metrics
*MetricOperationsApi* | [**GetAttributes**](docs/MetricOperationsApi.md#getattributes) | **Get** /metrics/{identifier}/attributes | Gets all attributes of a single metric
*NetworkGroupOperationsApi* | [**AddAll**](docs/NetworkGroupOperationsApi.md#addall) | **Post** /networkgroups | Creates one or multiple new network group(s)
*NetworkGroupOperationsApi* | [**AddNodes**](docs/NetworkGroupOperationsApi.md#addnodes) | **Post** /networkgroups/{identifier}/nodes | Adds nodes to an existing group
*NetworkGroupOperationsApi* | [**Delete**](docs/NetworkGroupOperationsApi.md#delete) | **Delete** /networkgroups/{identifier} | Deletes an existing network group
*NetworkGroupOperationsApi* | [**DeleteAll**](docs/NetworkGroupOperationsApi.md#deleteall) | **Delete** /networkgroups | Deletes a set of existing network groups
*NetworkGroupOperationsApi* | [**DeleteAttributes**](docs/NetworkGroupOperationsApi.md#deleteattributes) | **Delete** /networkgroups/{identifier}/attributes | Removes all attributes of an existing group
*NetworkGroupOperationsApi* | [**DeleteGlobalAttribute**](docs/NetworkGroupOperationsApi.md#deleteglobalattribute) | **Delete** /networkgroups/attributes/{label} | Deletes a global attribute defined for groups
*NetworkGroupOperationsApi* | [**Get**](docs/NetworkGroupOperationsApi.md#get) | **Get** /networkgroups/{identifier} | Gets one or more group(s)
*NetworkGroupOperationsApi* | [**GetAll**](docs/NetworkGroupOperationsApi.md#getall) | **Get** /networkgroups | Lists all groups
*NetworkGroupOperationsApi* | [**GetAttribute**](docs/NetworkGroupOperationsApi.md#getattribute) | **Get** /networkgroups/attributes/{label} | Gets a global attribute defined for groups
*NetworkGroupOperationsApi* | [**GetAttributes**](docs/NetworkGroupOperationsApi.md#getattributes) | **Get** /networkgroups/{identifier}/attributes | Gets all attributes of a single group
*NetworkGroupOperationsApi* | [**GetAvailableAction**](docs/NetworkGroupOperationsApi.md#getavailableaction) | **Get** /networkgroups/{identifier}/actions | Gets list of available actions on an existing group
*NetworkGroupOperationsApi* | [**GetGlobalAttributes**](docs/NetworkGroupOperationsApi.md#getglobalattributes) | **Get** /networkgroups/attributes | Gets all global attributes defined for groups
*NetworkGroupOperationsApi* | [**GetNode**](docs/NetworkGroupOperationsApi.md#getnode) | **Get** /networkgroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
*NetworkGroupOperationsApi* | [**GetNodes**](docs/NetworkGroupOperationsApi.md#getnodes) | **Get** /networkgroups/{identifier}/nodes | Gets all nodes of an existing group
*NetworkGroupOperationsApi* | [**Put**](docs/NetworkGroupOperationsApi.md#put) | **Put** /networkgroups/{identifier} | 
*NetworkGroupOperationsApi* | [**PutAll**](docs/NetworkGroupOperationsApi.md#putall) | **Put** /networkgroups | 
*NetworkGroupOperationsApi* | [**PutAttributes**](docs/NetworkGroupOperationsApi.md#putattributes) | **Put** /networkgroups/{identifier}/attributes | Adds or modifies attributes of an existing group
*NetworkGroupOperationsApi* | [**PutGlobalAttributes**](docs/NetworkGroupOperationsApi.md#putglobalattributes) | **Put** /networkgroups/attributes | Adds or modifies global attributes for groups
*NetworkGroupOperationsApi* | [**RemoveNode**](docs/NetworkGroupOperationsApi.md#removenode) | **Delete** /networkgroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
*NetworkGroupOperationsApi* | [**RemoveNodes**](docs/NetworkGroupOperationsApi.md#removenodes) | **Delete** /networkgroups/{identifier}/nodes | Removes some or all nodes from an existing group
*NetworkGroupOperationsApi* | [**RunAction**](docs/NetworkGroupOperationsApi.md#runaction) | **Post** /networkgroups/{identifier}/actions/{action} | Runs an action on a set of existing groups
*NetworkOperationsApi* | [**AddAll**](docs/NetworkOperationsApi.md#addall) | **Post** /networks | Creates one or multiple new networks
*NetworkOperationsApi* | [**Delete**](docs/NetworkOperationsApi.md#delete) | **Delete** /networks/{identifier} | Deletes an existing network
*NetworkOperationsApi* | [**DeleteAll**](docs/NetworkOperationsApi.md#deleteall) | **Delete** /networks | Deletes a set of existing networks
*NetworkOperationsApi* | [**DeleteAttributes**](docs/NetworkOperationsApi.md#deleteattributes) | **Delete** /networks/{identifier}/attributes | Removes all attributes of an existing network
*NetworkOperationsApi* | [**DeleteGlobalAttribute**](docs/NetworkOperationsApi.md#deleteglobalattribute) | **Delete** /networks/attributes/{label} | Deletes a global attribute defined for networks
*NetworkOperationsApi* | [**Get**](docs/NetworkOperationsApi.md#get) | **Get** /networks/{identifier} | Gets one or more network(s)
*NetworkOperationsApi* | [**GetAll**](docs/NetworkOperationsApi.md#getall) | **Get** /networks | Lists all networks
*NetworkOperationsApi* | [**GetAttributes**](docs/NetworkOperationsApi.md#getattributes) | **Get** /networks/{identifier}/attributes | Gets all attributes of a single network
*NetworkOperationsApi* | [**GetGlobalAttribute**](docs/NetworkOperationsApi.md#getglobalattribute) | **Get** /networks/attributes/{label} | Gets a global attribute defined for networks
*NetworkOperationsApi* | [**GetGlobalAttributes**](docs/NetworkOperationsApi.md#getglobalattributes) | **Get** /networks/attributes | Gets all global attributes defined for networks
*NetworkOperationsApi* | [**GetNic**](docs/NetworkOperationsApi.md#getnic) | **Get** /networks/{networkId}/nics/{nicId} | Gets one nic of a single network
*NetworkOperationsApi* | [**GetNics**](docs/NetworkOperationsApi.md#getnics) | **Get** /networks/{identifier}/nics | Gets all nics of a single network
*NetworkOperationsApi* | [**Put**](docs/NetworkOperationsApi.md#put) | **Put** /networks/{identifier} | Updates an existing network
*NetworkOperationsApi* | [**PutAll**](docs/NetworkOperationsApi.md#putall) | **Put** /networks | Updates a set of existing networks
*NetworkOperationsApi* | [**PutAttributes**](docs/NetworkOperationsApi.md#putattributes) | **Put** /networks/{identifier}/attributes | Adds or modifies attributes of an existing network
*NetworkOperationsApi* | [**PutGlobalAttributes**](docs/NetworkOperationsApi.md#putglobalattributes) | **Put** /networks/attributes | Adds or modifies global attributes for networks
*NicOperationsApi* | [**AddAll**](docs/NicOperationsApi.md#addall) | **Post** /nics | Creates one or multiple new nics
*NicOperationsApi* | [**Delete**](docs/NicOperationsApi.md#delete) | **Delete** /nics/{identifier} | Deletes an existing nic
*NicOperationsApi* | [**DeleteAll**](docs/NicOperationsApi.md#deleteall) | **Delete** /nics | Deletes a set of existing nics
*NicOperationsApi* | [**DeleteAttributes**](docs/NicOperationsApi.md#deleteattributes) | **Delete** /nics/{identifier}/attributes | Removes all attributes of an existing nic
*NicOperationsApi* | [**DeleteGlobalAttribute**](docs/NicOperationsApi.md#deleteglobalattribute) | **Delete** /nics/attributes/{label} | Deletes a global attribute defined for nics
*NicOperationsApi* | [**Get**](docs/NicOperationsApi.md#get) | **Get** /nics/{identifier} | Gets one or more nic(s)
*NicOperationsApi* | [**GetAll**](docs/NicOperationsApi.md#getall) | **Get** /nics | Lists all nics
*NicOperationsApi* | [**GetAttributes**](docs/NicOperationsApi.md#getattributes) | **Get** /nics/{identifier}/attributes | Gets all attributes of a single nic
*NicOperationsApi* | [**GetGlobalAttribute**](docs/NicOperationsApi.md#getglobalattribute) | **Get** /nics/attributes/{label} | Gets a global attribute defined for nics
*NicOperationsApi* | [**GetGlobalAttributes**](docs/NicOperationsApi.md#getglobalattributes) | **Get** /nics/attributes | Gets all global attributes defined for nics
*NicOperationsApi* | [**Put**](docs/NicOperationsApi.md#put) | **Put** /nics/{identifier} | Updates an existing nic
*NicOperationsApi* | [**PutAll**](docs/NicOperationsApi.md#putall) | **Put** /nics | Updates a set of existing nics
*NicOperationsApi* | [**PutAttributes**](docs/NicOperationsApi.md#putattributes) | **Put** /nics/{identifier}/attributes | Adds or modifies attributes of an existing nic
*NicOperationsApi* | [**PutGlobalAttributes**](docs/NicOperationsApi.md#putglobalattributes) | **Put** /nics/attributes | Adds or modifies global attributes for nics
*NodeOperationsApi* | [**AddAll**](docs/NodeOperationsApi.md#addall) | **Post** /nodes | Creates one or multiple new nodes
*NodeOperationsApi* | [**Delete**](docs/NodeOperationsApi.md#delete) | **Delete** /nodes/{identifier} | Deletes an existing node
*NodeOperationsApi* | [**DeleteAll**](docs/NodeOperationsApi.md#deleteall) | **Delete** /nodes | Deletes a set of existing nodes
*NodeOperationsApi* | [**DeleteAttributes**](docs/NodeOperationsApi.md#deleteattributes) | **Delete** /nodes/{identifier}/attributes | Removes all attributes of an existing node
*NodeOperationsApi* | [**DeleteGlobalAttribute**](docs/NodeOperationsApi.md#deleteglobalattribute) | **Delete** /nodes/attributes/{label} | Deletes a global attribute defined for nodes
*NodeOperationsApi* | [**Get**](docs/NodeOperationsApi.md#get) | **Get** /nodes/{identifier} | Gets one or more node(s)
*NodeOperationsApi* | [**GetAll**](docs/NodeOperationsApi.md#getall) | **Get** /nodes | Lists all nodes
*NodeOperationsApi* | [**GetAttributes**](docs/NodeOperationsApi.md#getattributes) | **Get** /nodes/{identifier}/attributes | Gets all attributes of a single node
*NodeOperationsApi* | [**GetAvailableAction**](docs/NodeOperationsApi.md#getavailableaction) | **Get** /nodes/{identifier}/actions | Gets list of available actions on an existing node
*NodeOperationsApi* | [**GetController**](docs/NodeOperationsApi.md#getcontroller) | **Get** /nodes/{nodeId}/controller | Get a node controller if existing
*NodeOperationsApi* | [**GetGlobalAttribute**](docs/NodeOperationsApi.md#getglobalattribute) | **Get** /nodes/attributes/{label} | Gets a global attribute defined for nodes
*NodeOperationsApi* | [**GetGlobalAttributes**](docs/NodeOperationsApi.md#getglobalattributes) | **Get** /nodes/attributes | Gets all global attributes defined for nodes
*NodeOperationsApi* | [**GetImageUnassigned**](docs/NodeOperationsApi.md#getimageunassigned) | **Get** /nodes/no_image | Lists all nodes that are not in any image group
*NodeOperationsApi* | [**GetNetworkUnassigned**](docs/NodeOperationsApi.md#getnetworkunassigned) | **Get** /nodes/no_network | Lists all nodes that are not in any network group
*NodeOperationsApi* | [**GetNic**](docs/NodeOperationsApi.md#getnic) | **Get** /nodes/{nodeId}/nics/{nicId} | Gets one nic of a single node
*NodeOperationsApi* | [**GetNics**](docs/NodeOperationsApi.md#getnics) | **Get** /nodes/{identifier}/nics | Gets all nics of a single node
*NodeOperationsApi* | [**Put**](docs/NodeOperationsApi.md#put) | **Put** /nodes/{identifier} | Updates an existing node
*NodeOperationsApi* | [**PutAll**](docs/NodeOperationsApi.md#putall) | **Put** /nodes | Updates a set of existing nodes
*NodeOperationsApi* | [**PutAttributes**](docs/NodeOperationsApi.md#putattributes) | **Put** /nodes/{identifier}/attributes | Adds or modifies attributes of an existing node
*NodeOperationsApi* | [**PutGlobalAttributes**](docs/NodeOperationsApi.md#putglobalattributes) | **Put** /nodes/attributes | Adds or modifies global attributes for nodes
*NodeOperationsApi* | [**RunAction**](docs/NodeOperationsApi.md#runaction) | **Post** /nodes/{identifier}/actions/{action} | Runs an action on a set of existing nodes
*NodeOperationsApi* | [**UnassignImage**](docs/NodeOperationsApi.md#unassignimage) | **Post** /nodes/no_image | Remove a set of nodes from their current image group
*NodeOperationsApi* | [**UnassignNetwork**](docs/NodeOperationsApi.md#unassignnetwork) | **Post** /nodes/no_network | Remove a set of nodes from their current network group
*NodeTemplateOperationsApi* | [**AddAll**](docs/NodeTemplateOperationsApi.md#addall) | **Post** /nodes/templates | Creates one or multiple new templates
*NodeTemplateOperationsApi* | [**Delete**](docs/NodeTemplateOperationsApi.md#delete) | **Delete** /nodes/templates/{identifier} | Deletes an existing template
*NodeTemplateOperationsApi* | [**DeleteAll**](docs/NodeTemplateOperationsApi.md#deleteall) | **Delete** /nodes/templates | Deletes a set of existing templates
*NodeTemplateOperationsApi* | [**DeleteAttributes**](docs/NodeTemplateOperationsApi.md#deleteattributes) | **Delete** /nodes/templates/{identifier}/attributes | Remove all attributes of an existing template
*NodeTemplateOperationsApi* | [**DeleteGlobalAttribute**](docs/NodeTemplateOperationsApi.md#deleteglobalattribute) | **Delete** /nodes/templates/attributes/{label} | Deletes a global attribute defined for templates
*NodeTemplateOperationsApi* | [**DeleteNicTemplate**](docs/NodeTemplateOperationsApi.md#deletenictemplate) | **Delete** /nodes/templates/{identifier}/nictemplates/{network} | Delete a nic template from the given node template
*NodeTemplateOperationsApi* | [**Get**](docs/NodeTemplateOperationsApi.md#get) | **Get** /nodes/templates/{identifier} | Gets one or more template(s)
*NodeTemplateOperationsApi* | [**GetAll**](docs/NodeTemplateOperationsApi.md#getall) | **Get** /nodes/templates | Lists all templates
*NodeTemplateOperationsApi* | [**GetAttributes**](docs/NodeTemplateOperationsApi.md#getattributes) | **Get** /nodes/templates/{identifier}/attributes | Gets all attributes of a single template
*NodeTemplateOperationsApi* | [**GetGlobalAttribute**](docs/NodeTemplateOperationsApi.md#getglobalattribute) | **Get** /nodes/templates/attributes/{label} | Gets a global attribute defined for templates
*NodeTemplateOperationsApi* | [**GetGlobalAttributes**](docs/NodeTemplateOperationsApi.md#getglobalattributes) | **Get** /nodes/templates/attributes | Gets all global attributes defined for templates
*NodeTemplateOperationsApi* | [**GetNicTemplate**](docs/NodeTemplateOperationsApi.md#getnictemplate) | **Get** /nodes/templates/{identifier}/nictemplates/{network} | Get a nic templates for the given network from a given node template
*NodeTemplateOperationsApi* | [**GetNicTemplates**](docs/NodeTemplateOperationsApi.md#getnictemplates) | **Get** /nodes/templates/{identifier}/nictemplates | Get all nic templates from a given node template
*NodeTemplateOperationsApi* | [**Put**](docs/NodeTemplateOperationsApi.md#put) | **Put** /nodes/templates/{identifier} | Updates an existing template
*NodeTemplateOperationsApi* | [**PutAll**](docs/NodeTemplateOperationsApi.md#putall) | **Put** /nodes/templates | Updates a set of existing templates
*NodeTemplateOperationsApi* | [**PutAttributes**](docs/NodeTemplateOperationsApi.md#putattributes) | **Put** /nodes/templates/{identifier}/attributes | Adds or modifies attributes of an existing template
*NodeTemplateOperationsApi* | [**PutGlobalAttributes**](docs/NodeTemplateOperationsApi.md#putglobalattributes) | **Put** /nodes/templates/attributes | Adds or modifies global attributes for templates
*NodeTemplateOperationsApi* | [**PutNicTemplate**](docs/NodeTemplateOperationsApi.md#putnictemplate) | **Put** /nodes/templates/{identifier}/nictemplates | Add or replace a nic template to the given node template
*SessionOperationsApi* | [**Delete**](docs/SessionOperationsApi.md#delete) | **Delete** /sessions/{token} | Deletes a session
*SessionOperationsApi* | [**GetAll**](docs/SessionOperationsApi.md#getall) | **Get** /sessions | Lists all sessions
*SessionOperationsApi* | [**GetSession**](docs/SessionOperationsApi.md#getsession) | **Get** /sessions/{token} | Gets a session
*SessionOperationsApi* | [**Login**](docs/SessionOperationsApi.md#login) | **Post** /sessions | Creates a new session
*SystemGroupOperationsApi* | [**AddAll**](docs/SystemGroupOperationsApi.md#addall) | **Post** /systemgroups | Creates one or multiple new system group(s)
*SystemGroupOperationsApi* | [**AddNodes**](docs/SystemGroupOperationsApi.md#addnodes) | **Post** /systemgroups/{identifier}/nodes | Adds nodes to an existing group
*SystemGroupOperationsApi* | [**Delete**](docs/SystemGroupOperationsApi.md#delete) | **Delete** /systemgroups/{identifier} | Deletes an existing system group
*SystemGroupOperationsApi* | [**DeleteAll**](docs/SystemGroupOperationsApi.md#deleteall) | **Delete** /systemgroups | Deletes a set of existing system groups
*SystemGroupOperationsApi* | [**DeleteAttributes**](docs/SystemGroupOperationsApi.md#deleteattributes) | **Delete** /systemgroups/{identifier}/attributes | Removes all attributes of an existing group
*SystemGroupOperationsApi* | [**DeleteGlobalAttribute**](docs/SystemGroupOperationsApi.md#deleteglobalattribute) | **Delete** /systemgroups/attributes/{label} | Deletes a global attribute defined for groups
*SystemGroupOperationsApi* | [**Get**](docs/SystemGroupOperationsApi.md#get) | **Get** /systemgroups/{identifier} | Gets one or more group(s)
*SystemGroupOperationsApi* | [**GetAll**](docs/SystemGroupOperationsApi.md#getall) | **Get** /systemgroups | Lists all groups
*SystemGroupOperationsApi* | [**GetAttribute**](docs/SystemGroupOperationsApi.md#getattribute) | **Get** /systemgroups/attributes/{label} | Gets a global attribute defined for groups
*SystemGroupOperationsApi* | [**GetAttributes**](docs/SystemGroupOperationsApi.md#getattributes) | **Get** /systemgroups/{identifier}/attributes | Gets all attributes of a single group
*SystemGroupOperationsApi* | [**GetAvailableAction**](docs/SystemGroupOperationsApi.md#getavailableaction) | **Get** /systemgroups/{identifier}/actions | Gets list of available actions on an existing group
*SystemGroupOperationsApi* | [**GetGlobalAttributes**](docs/SystemGroupOperationsApi.md#getglobalattributes) | **Get** /systemgroups/attributes | Gets all global attributes defined for groups
*SystemGroupOperationsApi* | [**GetNode**](docs/SystemGroupOperationsApi.md#getnode) | **Get** /systemgroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
*SystemGroupOperationsApi* | [**GetNodes**](docs/SystemGroupOperationsApi.md#getnodes) | **Get** /systemgroups/{identifier}/nodes | Gets all nodes of an existing group
*SystemGroupOperationsApi* | [**Put**](docs/SystemGroupOperationsApi.md#put) | **Put** /systemgroups/{identifier} | Updates a existing system group
*SystemGroupOperationsApi* | [**PutAll**](docs/SystemGroupOperationsApi.md#putall) | **Put** /systemgroups | Updates a set of existing system groups
*SystemGroupOperationsApi* | [**PutAttributes**](docs/SystemGroupOperationsApi.md#putattributes) | **Put** /systemgroups/{identifier}/attributes | Adds or modifies attributes of an existing group
*SystemGroupOperationsApi* | [**PutGlobalAttributes**](docs/SystemGroupOperationsApi.md#putglobalattributes) | **Put** /systemgroups/attributes | Adds or modifies global attributes for groups
*SystemGroupOperationsApi* | [**RemoveNode**](docs/SystemGroupOperationsApi.md#removenode) | **Delete** /systemgroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
*SystemGroupOperationsApi* | [**RemoveNodes**](docs/SystemGroupOperationsApi.md#removenodes) | **Delete** /systemgroups/{identifier}/nodes | Removes some or all nodes from an existing group
*SystemGroupOperationsApi* | [**RunAction**](docs/SystemGroupOperationsApi.md#runaction) | **Post** /systemgroups/{identifier}/actions/{action} | Runs an action on a set of existing groups
*TasksOperationsApi* | [**Delete**](docs/TasksOperationsApi.md#delete) | **Delete** /tasks/{identifier} | Deletes a single task
*TasksOperationsApi* | [**DeleteAll**](docs/TasksOperationsApi.md#deleteall) | **Delete** /tasks | Deletes a set of task
*TasksOperationsApi* | [**DeleteAttributes**](docs/TasksOperationsApi.md#deleteattributes) | **Delete** /tasks/{identifier}/attributes | 
*TasksOperationsApi* | [**Get**](docs/TasksOperationsApi.md#get) | **Get** /tasks/{identifier} | 
*TasksOperationsApi* | [**GetAll**](docs/TasksOperationsApi.md#getall) | **Get** /tasks | 
*TasksOperationsApi* | [**GetAttributes**](docs/TasksOperationsApi.md#getattributes) | **Get** /tasks/{identifier}/attributes | 
*TasksOperationsApi* | [**PutAttributes**](docs/TasksOperationsApi.md#putattributes) | **Put** /tasks/{identifier}/attributes | 

## Documentation For Models

 - [Action](docs/Action.md)
 - [ActionParameterObject](docs/ActionParameterObject.md)
 - [Actionable](docs/Actionable.md)
 - [Alert](docs/Alert.md)
 - [ApplicationDescriptionDto](docs/ApplicationDescriptionDto.md)
 - [Architecture](docs/Architecture.md)
 - [AttributeMapObject](docs/AttributeMapObject.md)
 - [AttributesDto](docs/AttributesDto.md)
 - [Controller](docs/Controller.md)
 - [ControllerSettings](docs/ControllerSettings.md)
 - [CustomGroup](docs/CustomGroup.md)
 - [Event](docs/Event.md)
 - [EventHook](docs/EventHook.md)
 - [Group](docs/Group.md)
 - [ImageGroup](docs/ImageGroup.md)
 - [ImageSettings](docs/ImageSettings.md)
 - [LocationSettings](docs/LocationSettings.md)
 - [LoginPasswordDto](docs/LoginPasswordDto.md)
 - [ManagementCard](docs/ManagementCard.md)
 - [ManagementSettings](docs/ManagementSettings.md)
 - [Metric](docs/Metric.md)
 - [MultipleIdentifierDto](docs/MultipleIdentifierDto.md)
 - [Network](docs/Network.md)
 - [NetworkGroup](docs/NetworkGroup.md)
 - [NetworkSettings](docs/NetworkSettings.md)
 - [Nic](docs/Nic.md)
 - [NicTemplate](docs/NicTemplate.md)
 - [Node](docs/Node.md)
 - [NodeTemplate](docs/NodeTemplate.md)
 - [Platform](docs/Platform.md)
 - [PlatformSettings](docs/PlatformSettings.md)
 - [PropertieDto](docs/PropertieDto.md)
 - [Role](docs/Role.md)
 - [SystemGroup](docs/SystemGroup.md)
 - [Task](docs/Task.md)
 - [Update](docs/Update.md)

## Documentation For Authorization

## X-Auth-Token
- **Type**: API key 

Example
```golang
auth := context.WithValue(context.Background(), sw.ContextAPIKey, sw.APIKey{
	Key: "APIKEY",
	Prefix: "Bearer", // Omit if not necessary.
})
r, err := client.Service.Operation(auth, args)
```

## Author


