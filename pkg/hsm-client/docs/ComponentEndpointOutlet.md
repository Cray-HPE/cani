# ComponentEndpointOutlet

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ID** | **string** |  | [optional] [default to null]
**Type_** | [***HmsType100**](HMSType.1.0.0.md) |  | [optional] [default to null]
**Domain** | **string** | Domain of component FQDN.  Hostname is always ID/xname | [optional] [default to null]
**FQDN** | **string** | Fully-qualified domain name of component on management network if for example the component is a node. | [optional] [default to null]
**RedfishType** | [***RedfishType100**](RedfishType.1.0.0.md) |  | [optional] [default to null]
**RedfishSubtype** | [***RedfishSubtype100**](RedfishSubtype.1.0.0.md) |  | [optional] [default to null]
**Enabled** | **bool** | To disable a component without deleting its data from the database, can be set to false | [optional] [default to null]
**ComponentEndpointType** | **string** | This is used as a discriminator to determine the additional RF-type- specific data that is kept for a ComponentEndpoint. | [default to null]
**MACAddr** | **string** | If the component e.g. a ComputerSystem/Node has a MAC on the management network, i.e. corresponding to the FQDN field&#x27;s Ethernet interface, this field will be present.  Not the HSN MAC.  Represented as the standard colon-separated 6 byte hex string. | [optional] [default to null]
**UUID** | **string** |  | [optional] [default to null]
**OdataID** | **string** |  | [optional] [default to null]
**RedfishEndpointID** | **string** |  | [optional] [default to null]
**RedfishEndpointFQDN** | **string** | This is a back-reference to the fully-qualified domain name of the parent Redfish endpoint that was used to discover the component.  It is the RedfishEndpointID field i.e. the hostname/xname plus its current plugin. | [optional] [default to null]
**RedfishURL** | **string** | Complete URL to the corresponding Redfish object, combining the RedfishEndpoint&#x27;s FQDN and the OdataID. | [optional] [default to null]
**RedfishChassisInfo** | [***ComponentEndpoint100RedfishOutletInfo**](ComponentEndpoint.1.0.0_RedfishOutletInfo.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

