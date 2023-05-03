# EthernetNicInfo100

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RedfishId** | **string** | The Redfish &#x27;Id&#x27; field for the interface. | [optional] [default to null]
**OdataId** | **string** | This is the relative path to the EthernetInterface via the Redfish entry point. (i.e. the @odata.id field). | [optional] [default to null]
**Description** | **string** | The Redfish &#x27;Description&#x27; field for the interface. | [optional] [default to null]
**FQDN** | **string** | The Redfish &#x27;FQDN&#x27; of the interface.  This may or may not be set and is not necessarily the same as the FQDN of the ComponentEndpoint. | [optional] [default to null]
**Hostname** | **string** | The Redfish &#x27;Hostname field&#x27; for the interface.  This may or may not be set and is not necessarily the same as the Hostname of the ComponentEndpoint. | [optional] [default to null]
**InterfaceEnabled** | **bool** | The Redfish &#x27;InterfaceEnabled&#x27; field if provided by Redfish, else it will be omitted. | [optional] [default to null]
**MACAddress** | **string** | The Redfish &#x27;MacAddress&#x27; field for the interface.  This should normally be set but is not necessarily the same as the MacAddr of the ComponentEndpoint (as there may be multiple interfaces). | [optional] [default to null]
**PermanentMACAddress** | **string** | The Redfish &#x27;PermanentMacAddress&#x27; field for the interface. This may or may not be set and is not necessarily the same as the MacAddr of the ComponentEndpoint (as there may be multiple interfaces). | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

