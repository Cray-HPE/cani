# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LoadstatePost**](CliFromFileApi.md#LoadstatePost) | **Post** /loadstate | Load services state and overwrite current service state
[**NetworksNetworkPut**](CliFromFileApi.md#NetworksNetworkPut) | **Put** /networks/{network} | Update a network object
[**NetworksPost**](CliFromFileApi.md#NetworksPost) | **Post** /networks | Create a new network

# **LoadstatePost**
> LoadstatePost(ctx, optional)
Load services state and overwrite current service state

\"Load services state and overwrite current service state. The format of the upload is implementation specific.\"

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***CliFromFileApiLoadstatePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a CliFromFileApiLoadstatePostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **slsDump** | [**optional.Interface of SlsState**](.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: multipart/form-data
 - **Accept**: Not defined

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
 **optional** | ***CliFromFileApiNetworksNetworkPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a CliFromFileApiNetworksNetworkPutOpts struct
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
 **optional** | ***CliFromFileApiNetworksPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a CliFromFileApiNetworksPostOpts struct
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

