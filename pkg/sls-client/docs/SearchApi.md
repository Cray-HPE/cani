# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**SearchHardwareGet**](SearchApi.md#SearchHardwareGet) | **Get** /search/hardware | Search for nodes matching a set of criteria
[**SearchNetworksGet**](SearchApi.md#SearchNetworksGet) | **Get** /search/networks | Perform a search for networks matching a set of criteria.

# **SearchHardwareGet**
> []Hardware SearchHardwareGet(ctx, optional)
Search for nodes matching a set of criteria

Search for nodes matching a set of criteria. Any of the properties of any entry in the database may be used as search keys.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SearchApiSearchHardwareGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SearchApiSearchHardwareGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **xname** | [**optional.Interface of string**](.md)| Matches the specified xname | 
 **parent** | [**optional.Interface of string**](.md)| Matches all objects that are direct children of the given xname | 
 **class** | [**optional.Interface of Hwclass**](.md)| Matches all objects of the given class | 
 **type_** | [**optional.Interface of string**](.md)| Matches all objects of the given type | 
 **powerConnector** | [**optional.Interface of string**](.md)| Matches all objects with the given xname in their power_connector property | 
 **object** | [**optional.Interface of string**](.md)| Matches all objects with the given xname in their object property. | 
 **nodeNics** | [**optional.Interface of string**](.md)| Matches all objects with the given xname in their node_nics property | 
 **networks** | **optional.String**| Matches all objects with the given xname in their networks property | 
 **peers** | [**optional.Interface of string**](.md)| Matches all objects with the given xname in their peers property | 

### Return type

[**[]Hardware**](hardware.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **SearchNetworksGet**
> []Network SearchNetworksGet(ctx, optional)
Perform a search for networks matching a set of criteria.

Perform a search for networks matching a set of criteria.  Any of the properties of any entry in the database may be used as search keys.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SearchApiSearchNetworksGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SearchApiSearchNetworksGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **name** | **optional.String**| Matches the specified network name | 
 **fullName** | **optional.String**| Matches the specified network full name | 
 **type_** | [**optional.Interface of string**](.md)| Matches the specified network type | 
 **ipAddress** | [**optional.Interface of string**](.md)| Matches all networks that could contain the specified IP address in their IP ranges | 

### Return type

[**[]Network**](network.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

