# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoPowerMapDelete**](PowerMapApi.md#DoPowerMapDelete) | **Delete** /sysinfo/powermaps/{xname} | Delete PowerMap with ID {xname}
[**DoPowerMapGet**](PowerMapApi.md#DoPowerMapGet) | **Get** /sysinfo/powermaps/{xname} | Retrieve PowerMap at {xname}
[**DoPowerMapPut**](PowerMapApi.md#DoPowerMapPut) | **Put** /sysinfo/powermaps/{xname} | Update definition for PowerMap ID {xname}
[**DoPowerMapsDeleteAll**](PowerMapApi.md#DoPowerMapsDeleteAll) | **Delete** /sysinfo/powermaps | Delete all PowerMap entities
[**DoPowerMapsGet**](PowerMapApi.md#DoPowerMapsGet) | **Get** /sysinfo/powermaps | Retrieve all PowerMaps, returning PowerMapArray
[**DoPowerMapsPost**](PowerMapApi.md#DoPowerMapsPost) | **Post** /sysinfo/powermaps | Create or Modify PowerMaps

# **DoPowerMapDelete**
> Response100 DoPowerMapDelete(ctx, xname)
Delete PowerMap with ID {xname}

Delete PowerMap entry for a specific component {xname}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of PowerMap record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPowerMapGet**
> PowerMap100PowerMap DoPowerMapGet(ctx, xname)
Retrieve PowerMap at {xname}

Retrieve PowerMap for a component located at physical location {xname}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of PowerMap record to return. | 

### Return type

[**PowerMap100PowerMap**](PowerMap.1.0.0_PowerMap.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPowerMapPut**
> Response100 DoPowerMapPut(ctx, body, xname)
Update definition for PowerMap ID {xname}

Update or create an entry for an individual component xname using PUT. If the PUT operation contains an xname that already exists, the entry will be overwritten with the new entry.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**PowerMap100PowerMap**](PowerMap100PowerMap.md)|  | 
  **xname** | **string**| Locational xname of PowerMap record to create or update. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPowerMapsDeleteAll**
> Response100 DoPowerMapsDeleteAll(ctx, )
Delete all PowerMap entities

Delete all entries in the PowerMaps collection.

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

# **DoPowerMapsGet**
> []PowerMap100PostPowerMap DoPowerMapsGet(ctx, )
Retrieve all PowerMaps, returning PowerMapArray

Retrieve all power map entries as a named array, or an empty array if the collection is empty.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]PowerMap100PostPowerMap**](array.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPowerMapsPost**
> Response100 DoPowerMapsPost(ctx, body)
Create or Modify PowerMaps

Create or update the given set of PowerMaps whose ID fields are each a valid xname. The poweredBy field is required.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]PowerMap100PostPowerMap**](PowerMap.1.0.0_PostPowerMap.md)|  | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

