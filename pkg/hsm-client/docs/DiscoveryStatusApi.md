# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoDiscoveryStatusGet**](DiscoveryStatusApi.md#DoDiscoveryStatusGet) | **Get** /Inventory/DiscoveryStatus/{id} | Retrieve DiscoveryStatus entry matching {id}
[**DoDiscoveryStatusGetAll**](DiscoveryStatusApi.md#DoDiscoveryStatusGetAll) | **Get** /Inventory/DiscoveryStatus | Retrieve all DiscoveryStatus entries in collection

# **DoDiscoveryStatusGet**
> DiscoveryStatus100DiscoveryStatus DoDiscoveryStatusGet(ctx, id)
Retrieve DiscoveryStatus entry matching {id}

Retrieve DiscoveryStatus entry with the specific ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **float64**| Positive integer ID of DiscoveryStatus entry to retrieve | 

### Return type

[**DiscoveryStatus100DiscoveryStatus**](DiscoveryStatus.1.0.0_DiscoveryStatus.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoDiscoveryStatusGetAll**
> []DiscoveryStatus100DiscoveryStatus DoDiscoveryStatusGetAll(ctx, )
Retrieve all DiscoveryStatus entries in collection

Retrieve all DiscoveryStatus entries as an unnamed array.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**[]DiscoveryStatus100DiscoveryStatus**](DiscoveryStatus.1.0.0_DiscoveryStatus.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

