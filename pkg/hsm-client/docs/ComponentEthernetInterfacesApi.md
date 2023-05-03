# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoCompEthInterfaceDeleteAllV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceDeleteAllV2) | **Delete** /Inventory/EthernetInterfaces | Clear the component Ethernet interface collection.
[**DoCompEthInterfaceDeleteV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceDeleteV2) | **Delete** /Inventory/EthernetInterfaces/{ethInterfaceID} | DELETE existing component Ethernet interface with {ethInterfaceID}
[**DoCompEthInterfaceGetV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceGetV2) | **Get** /Inventory/EthernetInterfaces/{ethInterfaceID} | GET existing component Ethernet interface {ethInterfaceID}
[**DoCompEthInterfaceIPAddressDeleteV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceIPAddressDeleteV2) | **Delete** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses/{ipAddress} | DELETE existing IP address mapping with {ipAddress} from a component Ethernet interface with {ethInterfaceID}
[**DoCompEthInterfaceIPAddressPatchV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceIPAddressPatchV2) | **Patch** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses/{ipAddress} | UPDATE metadata for existing IP address {ipAddress} in a component Ethernet interface {ethInterfaceID
[**DoCompEthInterfaceIPAddressesGetV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceIPAddressesGetV2) | **Get** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses | Retrieve all IP addresses of a component Ethernet interface {ethInterfaceID}
[**DoCompEthInterfaceIPAddressesPostV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfaceIPAddressesPostV2) | **Post** /Inventory/EthernetInterfaces/{ethInterfaceID}/IPAddresses | CREATE a new IP address mapping in a component Ethernet interface (via POST)
[**DoCompEthInterfacePatchV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfacePatchV2) | **Patch** /Inventory/EthernetInterfaces/{ethInterfaceID} | UPDATE metadata for existing component Ethernet interface {ethInterfaceID}
[**DoCompEthInterfacePostV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfacePostV2) | **Post** /Inventory/EthernetInterfaces | CREATE a new component Ethernet interface (via POST)
[**DoCompEthInterfacesGetV2**](ComponentEthernetInterfacesApi.md#DoCompEthInterfacesGetV2) | **Get** /Inventory/EthernetInterfaces | GET ALL existing component Ethernet interfaces

# **DoCompEthInterfaceDeleteAllV2**
> Response100 DoCompEthInterfaceDeleteAllV2(ctx, )
Clear the component Ethernet interface collection.

Delete all component Ethernet interface entries.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceDeleteV2**
> Response100 DoCompEthInterfaceDeleteV2(ctx, ethInterfaceID)
DELETE existing component Ethernet interface with {ethInterfaceID}

Delete the given component Ethernet interface with {ethInterfaceID}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceGetV2**
> CompEthInterface100 DoCompEthInterfaceGetV2(ctx, ethInterfaceID)
GET existing component Ethernet interface {ethInterfaceID}

Retrieve the component Ethernet interface which was created with the given {ethInterfaceID}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to return. | 

### Return type

[**CompEthInterface100**](CompEthInterface.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceIPAddressDeleteV2**
> Response100 DoCompEthInterfaceIPAddressDeleteV2(ctx, ethInterfaceID, ipAddress)
DELETE existing IP address mapping with {ipAddress} from a component Ethernet interface with {ethInterfaceID}

Delete the given IP address mapping with {ipAddress} from a component Ethernet interface with {ethInterfaceID}. The 'LastUpdate' field of the component Ethernet interface will be updated\"

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to delete the IP address from | 
  **ipAddress** | **string**| The IP address to delete from the component Ethernet interface. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceIPAddressPatchV2**
> DoCompEthInterfaceIPAddressPatchV2(ctx, body, ethInterfaceID, ipAddress)
UPDATE metadata for existing IP address {ipAddress} in a component Ethernet interface {ethInterfaceID

\"To update the network of an IP address in a component Ethernet interface, a PATCH operation can be used. Omitted fields are not updated. The 'LastUpdate' field of the component Ethernet interface will be updated\"

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CompEthInterface100IpAddressMappingPatch**](CompEthInterface100IpAddressMappingPatch.md)|  | 
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface with the IP address to patch. | 
  **ipAddress** | **string**| The IP address to patch from the component Ethernet interface. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceIPAddressesGetV2**
> []CompEthInterface100IpAddressMapping DoCompEthInterfaceIPAddressesGetV2(ctx, ethInterfaceID)
Retrieve all IP addresses of a component Ethernet interface {ethInterfaceID}

Retrieve all IP addresses of a component Ethernet interface {ethInterfaceID}

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to retrieve the IP addresses of. | 

### Return type

[**[]CompEthInterface100IpAddressMapping**](CompEthInterface.1.0.0_IPAddressMapping.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfaceIPAddressesPostV2**
> ResourceUri100 DoCompEthInterfaceIPAddressesPostV2(ctx, body, ethInterfaceID)
CREATE a new IP address mapping in a component Ethernet interface (via POST)

Create a new IP address mapping in a component Ethernet interface {ethInterfaceID}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CompEthInterface100IpAddressMapping**](CompEthInterface100IpAddressMapping.md)|  | 
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to add the IP address to. | 

### Return type

[**ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfacePatchV2**
> DoCompEthInterfacePatchV2(ctx, body, ethInterfaceID)
UPDATE metadata for existing component Ethernet interface {ethInterfaceID}

To update the IP address, CompID, and/or description of a component Ethernet interface, a PATCH operation can be used. Omitted fields are not updated. The 'LastUpdate' field will be updated if an IP address is provided.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CompEthInterface100Patch**](CompEthInterface100Patch.md)|  | 
  **ethInterfaceID** | **string**| The ID of the component Ethernet interface to update. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfacePostV2**
> ResourceUri100 DoCompEthInterfacePostV2(ctx, body)
CREATE a new component Ethernet interface (via POST)

Create a new component Ethernet interface.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**CompEthInterface100**](CompEthInterface100.md)|  | 

### Return type

[**ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEthInterfacesGetV2**
> []CompEthInterface100 DoCompEthInterfacesGetV2(ctx, optional)
GET ALL existing component Ethernet interfaces

Get all component Ethernet interfaces that currently exist, optionally filtering the set, returning an array of component Ethernet interfaces.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ComponentEthernetInterfacesApiDoCompEthInterfacesGetV2Opts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ComponentEthernetInterfacesApiDoCompEthInterfacesGetV2Opts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mACAddress** | **optional.String**| Retrieve the component Ethernet interface with the provided MAC address. Can be repeated to select multiple component Ethernet interfaces. | 
 **iPAddress** | **optional.String**| Retrieve the component Ethernet interface with the provided IP address. Can be repeated to select multiple component Ethernet interfaces. A blank string will retrieve component Ethernet interfaces that have no IP address. | 
 **network** | **optional.String**| Retrieve the component Ethernet interface with a IP addresses on the provided  network. Can be repeated to select multiple component Ethernet interfaces. A blank string will retrieve component Ethernet interfaces that have an IP address with no  network. | 
 **componentID** | **optional.String**| Retrieve all component Ethernet interfaces with the provided component ID. Can be repeated to select multiple component Ethernet interfaces. | 
 **type_** | **optional.String**| Retrieve all component Ethernet interfaces with the provided parent HMS type. Can be repeated to select multiple component Ethernet interfaces. | 
 **olderThan** | **optional.String**| Retrieve all component Ethernet interfaces that were last updated before the specified time. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 
 **newerThan** | **optional.String**| Retrieve all component Ethernet interfaces that were last updated after the specified time. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 

### Return type

[**[]CompEthInterface100**](CompEthInterface.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

