# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**HardwareGet**](HardwareApi.md#HardwareGet) | **Get** /hardware | Retrieve a list of hardware in the system.
[**HardwarePost**](HardwareApi.md#HardwarePost) | **Post** /hardware | Create a new hardware object
[**HardwareXnameDelete**](HardwareApi.md#HardwareXnameDelete) | **Delete** /hardware/{xname} | Delete the xname
[**HardwareXnameGet**](HardwareApi.md#HardwareXnameGet) | **Get** /hardware/{xname} | Retrieve information about the requested xname
[**HardwareXnamePut**](HardwareApi.md#HardwareXnamePut) | **Put** /hardware/{xname} | Update a hardware object

# **HardwareGet**
> []Hardware HardwareGet(ctx, )
Retrieve a list of hardware in the system.

Retrieve a JSON list of the networks available in the system.  Return value is an array of hardware objects representing all the hardware in the system.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]Hardware**](hardware.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **HardwarePost**
> Hardware HardwarePost(ctx, optional)
Create a new hardware object

Create a new hardware object.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***HardwareApiHardwarePostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HardwareApiHardwarePostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of HardwarePost**](HardwarePost.md)|  | 

### Return type

[**Hardware**](hardware.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **HardwareXnameDelete**
> HardwareXnameDelete(ctx, xname)
Delete the xname

Delete the requested xname from SLS. Note that if you delete a parent object, then the children are also deleted from SLS. If the child object happens to be a parent, then the deletion can cascade down levels. If you delete a child object, it does not affect the parent.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | [**string**](.md)| The xname to look up or alter. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **HardwareXnameGet**
> Hardware HardwareXnameGet(ctx, xname)
Retrieve information about the requested xname

Retrieve information about the requested xname. All properties are returned as a JSON array.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | [**string**](.md)| The xname to look up or alter. | 

### Return type

[**Hardware**](hardware.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **HardwareXnamePut**
> Hardware HardwareXnamePut(ctx, xname, optional)
Update a hardware object

Update a hardware object.  Parent objects will be created, if possible.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | [**string**](.md)| The xname to look up or alter. | 
 **optional** | ***HardwareApiHardwareXnamePutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HardwareApiHardwareXnamePutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of HardwarePut**](HardwarePut.md)|  | 

### Return type

[**Hardware**](hardware.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

