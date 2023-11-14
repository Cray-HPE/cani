# {{classname}}

All URIs are relative to *https://localhost:8080/cmu/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**Get**](ApplicationApi.md#Get) | **Get** / | Lists application entry points
[**GetSettings**](ApplicationApi.md#GetSettings) | **Get** /settings | Lists application current settings

# **Get**
> ApplicationDescriptionDto Get(ctx, optional)
Lists application entry points

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ApplicationApiGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ApplicationApiGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**ApplicationDescriptionDto**](ApplicationDescriptionDTO.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetSettings**
> map[string]PropertieDto GetSettings(ctx, optional)
Lists application current settings

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ApplicationApiGetSettingsOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ApplicationApiGetSettingsOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

[**map[string]PropertieDto**](PropertieDTO.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

