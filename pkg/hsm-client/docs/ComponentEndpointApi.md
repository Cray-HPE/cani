# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoComponentEndpointDelete**](ComponentEndpointApi.md#DoComponentEndpointDelete) | **Delete** /Inventory/ComponentEndpoints/{xname} | Delete ComponentEndpoint with ID {xname}
[**DoComponentEndpointGet**](ComponentEndpointApi.md#DoComponentEndpointGet) | **Get** /Inventory/ComponentEndpoints/{xname} | Retrieve ComponentEndpoint at {xname}
[**DoComponentEndpointsDeleteAll**](ComponentEndpointApi.md#DoComponentEndpointsDeleteAll) | **Delete** /Inventory/ComponentEndpoints | Delete all ComponentEndpoints
[**DoComponentEndpointsGet**](ComponentEndpointApi.md#DoComponentEndpointsGet) | **Get** /Inventory/ComponentEndpoints | Retrieve ComponentEndpoints Collection

# **DoComponentEndpointDelete**
> Response100 DoComponentEndpointDelete(ctx, xname)
Delete ComponentEndpoint with ID {xname}

Delete ComponentEndpoint for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of ComponentEndpoint record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentEndpointGet**
> ComponentEndpoint100ComponentEndpoint DoComponentEndpointGet(ctx, xname)
Retrieve ComponentEndpoint at {xname}

Retrieve ComponentEndpoint record for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of ComponentEndpoint record to return. | 

### Return type

[**ComponentEndpoint100ComponentEndpoint**](ComponentEndpoint.1.0.0_ComponentEndpoint.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentEndpointsDeleteAll**
> Response100 DoComponentEndpointsDeleteAll(ctx, )
Delete all ComponentEndpoints

Delete all entries in the ComponentEndpoint collection.

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

# **DoComponentEndpointsGet**
> ComponentEndpointArrayComponentEndpointArray DoComponentEndpointsGet(ctx, optional)
Retrieve ComponentEndpoints Collection

Retrieve the full collection of ComponentEndpoints in the form of a ComponentEndpointArray. Full results can also be filtered by query parameters. Only the first filter parameter of each type is used and the parameters are applied in an AND fashion. If the collection is empty or the filters have no match, an empty array is returned.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ComponentEndpointApiDoComponentEndpointsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ComponentEndpointApiDoComponentEndpointsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **optional.String**| Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames. | 
 **redfishEp** | **optional.String**| Retrieve all ComponentEndpoints managed by the parent Redfish EP. | 
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **redfishType** | **optional.String**| Retrieve all ComponentEndpoints with the given Redfish type. | 

### Return type

[**ComponentEndpointArrayComponentEndpointArray**](ComponentEndpointArray_ComponentEndpointArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

