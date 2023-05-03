# ServiceEndpoint100ServiceEndpoint

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**RedfishEndpointID** | **string** |  | [optional] [default to null]
**RedfishType** | [***RedfishType100**](RedfishType.1.0.0.md) |  | [optional] [default to null]
**RedfishSubtype** | [***RedfishSubtype100**](RedfishSubtype.1.0.0.md) |  | [optional] [default to null]
**UUID** | **string** |  | [optional] [default to null]
**OdataID** | **string** |  | [optional] [default to null]
**RedfishEndpointFQDN** | **string** | This is a back-reference to the fully-qualified domain name of the parent Redfish endpoint that was used to discover the component.  It is the RedfishEndpointID field i.e. the hostname/xname plus its current domain. | [optional] [default to null]
**RedfishURL** | **string** | This is the complete URL to the corresponding Redfish object, combining the RedfishEndpoint&#x27;s FQDN and the OdataID. | [optional] [default to null]
**ServiceInfo** | [***ServiceEndpoint100ServiceInfo**](ServiceEndpoint.1.0.0_ServiceInfo.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

