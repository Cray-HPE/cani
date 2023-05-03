# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**NetworksGet**](NetworkApi.md#NetworksGet) | **Get** /networks | Retrieve a list of networks in the system
[**NetworksNetworkDelete**](NetworkApi.md#NetworksNetworkDelete) | **Delete** /networks/{network} | Delete the named network
[**NetworksNetworkGet**](NetworkApi.md#NetworksNetworkGet) | **Get** /networks/{network} | Retrieve a network item
[**NetworksNetworkPut**](NetworkApi.md#NetworksNetworkPut) | **Put** /networks/{network} | Update a network object
[**NetworksPost**](NetworkApi.md#NetworksPost) | **Post** /networks | Create a new network

# **NetworksGet**
> []Network NetworksGet(ctx, )
Retrieve a list of networks in the system

Retrieve a JSON list of the networks available in the system.  Return value is an array of strings with each string representing the name field of the network object. 

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]Network**](network.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **NetworksNetworkDelete**
> NetworksNetworkDelete(ctx, network)
Delete the named network

Delete the specific network from SLS.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **network** | **string**| The network to look up or alter. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **NetworksNetworkGet**
> Network NetworksNetworkGet(ctx, network)
Retrieve a network item

Retrieve the specific network.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **network** | **string**| The network to look up or alter. | 

### Return type

[**Network**](network.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **NetworksNetworkPut**
> Network NetworksNetworkPut(ctx, network, optional)
Update a network object

Update a network object. Parent objects will be created, if possible.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **network** | **string**| The network to look up or alter. | 
 **optional** | ***NetworkApiNetworksNetworkPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkApiNetworksNetworkPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of Network**](Network.md)|  | 

### Return type

[**Network**](network.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **NetworksPost**
> NetworksPost(ctx, optional)
Create a new network

Create a new network. Must include all fields at the time of upload.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NetworkApiNetworksPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkApiNetworksPostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of Network**](Network.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

