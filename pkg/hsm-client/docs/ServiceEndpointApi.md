# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoServiceEndpointDelete**](ServiceEndpointApi.md#DoServiceEndpointDelete) | **Delete** /Inventory/ServiceEndpoints/{service}/RedfishEndpoints/{xname} | Delete the {service} ServiceEndpoint managed by {xname}
[**DoServiceEndpointGet**](ServiceEndpointApi.md#DoServiceEndpointGet) | **Get** /Inventory/ServiceEndpoints/{service}/RedfishEndpoints/{xname} | Retrieve the ServiceEndpoint of a {service} managed by {xname}
[**DoServiceEndpointsDeleteAll**](ServiceEndpointApi.md#DoServiceEndpointsDeleteAll) | **Delete** /Inventory/ServiceEndpoints | Delete all ServiceEndpoints
[**DoServiceEndpointsGet**](ServiceEndpointApi.md#DoServiceEndpointsGet) | **Get** /Inventory/ServiceEndpoints/{service} | Retrieve all ServiceEndpoints of a {service}
[**DoServiceEndpointsGetAll**](ServiceEndpointApi.md#DoServiceEndpointsGetAll) | **Get** /Inventory/ServiceEndpoints | Retrieve ServiceEndpoints Collection

# **DoServiceEndpointDelete**
> Response100 DoServiceEndpointDelete(ctx, service, xname)
Delete the {service} ServiceEndpoint managed by {xname}

Delete the {service} ServiceEndpoint managed by {xname}

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **service** | **string**| The Redfish service type of the ServiceEndpoint record to delete. | 
  **xname** | **string**| The locational xname of the RedfishEndpoint that manages the ServiceEndpoint record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoServiceEndpointGet**
> ServiceEndpoint100ServiceEndpoint DoServiceEndpointGet(ctx, service, xname)
Retrieve the ServiceEndpoint of a {service} managed by {xname}

Retrieve the ServiceEndpoint for a Redfish service that is managed by xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **service** | **string**| The Redfish service type of the ServiceEndpoint record to return. | 
  **xname** | **string**| The locational xname of the RedfishEndpoint that manages the ServiceEndpoint record to return. | 

### Return type

[**ServiceEndpoint100ServiceEndpoint**](ServiceEndpoint.1.0.0_ServiceEndpoint.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoServiceEndpointsDeleteAll**
> Response100 DoServiceEndpointsDeleteAll(ctx, )
Delete all ServiceEndpoints

Delete all entries in the ServiceEndpoint collection.

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

# **DoServiceEndpointsGet**
> ServiceEndpointArrayServiceEndpointArray DoServiceEndpointsGet(ctx, service, optional)
Retrieve all ServiceEndpoints of a {service}

Retrieve all ServiceEndpoint records for the Redfish service.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **service** | **string**| The Redfish service type of the ServiceEndpoint records to return. | 
 **optional** | ***ServiceEndpointApiDoServiceEndpointsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ServiceEndpointApiDoServiceEndpointsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **redfishEp** | **optional.String**| Retrieve all ServiceEndpoints of type {service} managed by the parent Redfish EP. Can be repeated to select groups of endpoints. | 

### Return type

[**ServiceEndpointArrayServiceEndpointArray**](ServiceEndpointArray_ServiceEndpointArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoServiceEndpointsGetAll**
> ServiceEndpointArrayServiceEndpointArray DoServiceEndpointsGetAll(ctx, optional)
Retrieve ServiceEndpoints Collection

Retrieve the full collection of ServiceEndpoints in the form of a ServiceEndpointArray. Full results can also be filtered by query parameters.  Only the first filter parameter of each type is used and the parameters are applied in an AND fashion. If the collection is empty or the filters have no match, an empty array is returned.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ServiceEndpointApiDoServiceEndpointsGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ServiceEndpointApiDoServiceEndpointsGetAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **redfishEp** | **optional.String**| Retrieve all ServiceEndpoints managed by the parent Redfish EP. Can be repeated to select groups of endpoints. | 
 **service** | **optional.String**| Retrieve all ServiceEndpoints of the given Redfish service. | 

### Return type

[**ServiceEndpointArrayServiceEndpointArray**](ServiceEndpointArray_ServiceEndpointArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

