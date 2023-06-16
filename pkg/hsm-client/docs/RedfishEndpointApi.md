# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoRedfishEndpointDelete**](RedfishEndpointApi.md#DoRedfishEndpointDelete) | **Delete** /Inventory/RedfishEndpoints/{xname} | Delete RedfishEndpoint with ID {xname}
[**DoRedfishEndpointGet**](RedfishEndpointApi.md#DoRedfishEndpointGet) | **Get** /Inventory/RedfishEndpoints/{xname} | Retrieve RedfishEndpoint at {xname}
[**DoRedfishEndpointPatch**](RedfishEndpointApi.md#DoRedfishEndpointPatch) | **Patch** /Inventory/RedfishEndpoints/{xname} | Update (PATCH) definition for RedfishEndpoint ID {xname}
[**DoRedfishEndpointPut**](RedfishEndpointApi.md#DoRedfishEndpointPut) | **Put** /Inventory/RedfishEndpoints/{xname} | Update definition for RedfishEndpoint ID {xname}
[**DoRedfishEndpointQueryGet**](RedfishEndpointApi.md#DoRedfishEndpointQueryGet) | **Get** /Inventory/RedfishEndpoints/Query/{xname} | Retrieve RedfishEndpoint query for {xname}, returning RedfishEndpointArray
[**DoRedfishEndpointsDeleteAll**](RedfishEndpointApi.md#DoRedfishEndpointsDeleteAll) | **Delete** /Inventory/RedfishEndpoints | Delete all RedfishEndpoints
[**DoRedfishEndpointsGet**](RedfishEndpointApi.md#DoRedfishEndpointsGet) | **Get** /Inventory/RedfishEndpoints | Retrieve all RedfishEndpoints, returning RedfishEndpointArray
[**DoRedfishEndpointsPost**](RedfishEndpointApi.md#DoRedfishEndpointsPost) | **Post** /Inventory/RedfishEndpoints | Create RedfishEndpoint(s)

# **DoRedfishEndpointDelete**
> Response100 DoRedfishEndpointDelete(ctx, xname)
Delete RedfishEndpoint with ID {xname}

Delete RedfishEndpoint record for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of RedfishEndpoint record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointGet**
> RedfishEndpoint100RedfishEndpoint DoRedfishEndpointGet(ctx, xname)
Retrieve RedfishEndpoint at {xname}

Retrieve RedfishEndpoint, located at physical location {xname}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of RedfishEndpoint record to return. | 

### Return type

[**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint.1.0.0_RedfishEndpoint.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointPatch**
> RedfishEndpoint100RedfishEndpoint DoRedfishEndpointPatch(ctx, body, xname)
Update (PATCH) definition for RedfishEndpoint ID {xname}

Update (PATCH) RedfishEndpoint record for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint100RedfishEndpoint.md)|  | 
  **xname** | **string**| Locational xname of RedfishEndpoint record to create or update. | 

### Return type

[**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint.1.0.0_RedfishEndpoint.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointPut**
> RedfishEndpoint100RedfishEndpoint DoRedfishEndpointPut(ctx, body, xname)
Update definition for RedfishEndpoint ID {xname}

Create or update RedfishEndpoint record for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint100RedfishEndpoint.md)|  | 
  **xname** | **string**| Locational xname of RedfishEndpoint record to create or update. | 

### Return type

[**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint.1.0.0_RedfishEndpoint.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointQueryGet**
> RedfishEndpointArrayRedfishEndpointArray DoRedfishEndpointQueryGet(ctx, xname)
Retrieve RedfishEndpoint query for {xname}, returning RedfishEndpointArray

Given xname and modifiers in query string, retrieve zero or more RedfishEndpoint entries in the form of a RedfishEndpointArray.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of RedfishEndpoint to query. | 

### Return type

[**RedfishEndpointArrayRedfishEndpointArray**](RedfishEndpointArray_RedfishEndpointArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointsDeleteAll**
> Response100 DoRedfishEndpointsDeleteAll(ctx, )
Delete all RedfishEndpoints

Delete all entries in the RedfishEndpoint collection.

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

# **DoRedfishEndpointsGet**
> RedfishEndpointArrayRedfishEndpointArray DoRedfishEndpointsGet(ctx, optional)
Retrieve all RedfishEndpoints, returning RedfishEndpointArray

Retrieve all Redfish endpoint entries as a named array, optionally filtering it.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***RedfishEndpointApiDoRedfishEndpointsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a RedfishEndpointApiDoRedfishEndpointsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **optional.String**| Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames. | 
 **fqdn** | **optional.String**| Retrieve RedfishEndpoint with the given FQDN | 
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **uuid** | **optional.String**| Retrieve the RedfishEndpoint with the given UUID. | 
 **macaddr** | **optional.String**| Retrieve the RedfishEndpoint with the given MAC address. | 
 **ipaddress** | **optional.String**| Retrieve the RedfishEndpoint with the given IP address. A blank string will get Redfish endpoints without IP addresses. | 
 **lastdiscoverystatus** | **optional.String**| Retrieve the RedfishEndpoints with the given discovery status. This can be negated (i.e. !DiscoverOK). Valid values are: EndpointInvalid, EPResponseFailedDecode, HTTPsGetFailed, NotYetQueried, VerificationFailed, ChildVerificationFailed, DiscoverOK | 

### Return type

[**RedfishEndpointArrayRedfishEndpointArray**](RedfishEndpointArray_RedfishEndpointArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointsPost**
> []ResourceUri100 DoRedfishEndpointsPost(ctx, body)
Create RedfishEndpoint(s)

Create a new RedfishEndpoint whose ID field is a valid xname. ID can be given explicitly, or if the Hostname or hostname portion of the FQDN is given, and is a valid xname, this will be used for the ID instead.  The Hostname/Domain can be given as separate fields and will be used to create a FQDN if one is not given. The reverse is also true.  If FQDN is an IP address it will be treated as a hostname with a blank domain.  The domain field is used currently to assign the domain for discovered nodes automatically.  If ID is given and is a valid XName, the hostname/domain/FQDN does not need to have an XName as the hostname portion. It can be any address. The ID and FQDN must be unique across all entries.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint100RedfishEndpoint.md)|  | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

