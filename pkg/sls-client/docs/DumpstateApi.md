# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DumpstateGet**](DumpstateApi.md#DumpstateGet) | **Get** /dumpstate | Retrieve a dump of current service state
[**LoadstatePost**](DumpstateApi.md#LoadstatePost) | **Post** /loadstate | Load services state and overwrite current service state

# **DumpstateGet**
> SlsState DumpstateGet(ctx, )
Retrieve a dump of current service state

Get a dump of current service state. The format of this is implementation-specific.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**SlsState**](slsState.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LoadstatePost**
> LoadstatePost(ctx, optional)
Load services state and overwrite current service state

\"Load services state and overwrite current service state. The format of the upload is implementation specific.\"

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***DumpstateApiLoadstatePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DumpstateApiLoadstatePostOpts struct
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

