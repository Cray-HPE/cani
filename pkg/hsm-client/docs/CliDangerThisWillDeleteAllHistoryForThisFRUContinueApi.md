# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvHistByFRUDelete**](CliDangerThisWillDeleteAllHistoryForThisFRUContinueApi.md#DoHWInvHistByFRUDelete) | **Delete** /Inventory/HardwareByFRU/History/{fruid} | Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}

# **DoHWInvHistByFRUDelete**
> Response100 DoHWInvHistByFRUDelete(ctx, fruid)
Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}

Delete history for an entry in the HWInventoryByFRU collection.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **fruid** | **string**| Locational xname of HWInventoryByFRU record to delete history for. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

