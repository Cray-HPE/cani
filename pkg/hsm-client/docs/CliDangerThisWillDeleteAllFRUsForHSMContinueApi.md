# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvByFRUDeleteAll**](CliDangerThisWillDeleteAllFRUsForHSMContinueApi.md#DoHWInvByFRUDeleteAll) | **Delete** /Inventory/HardwareByFRU | Delete all HWInventoryByFRU entries

# **DoHWInvByFRUDeleteAll**
> Response100 DoHWInvByFRUDeleteAll(ctx, )
Delete all HWInventoryByFRU entries

Delete all entries in the HWInventoryByFRU collection. Note that this does not delete any associated HWInventoryByLocation entries. Also, if any items are associated with a HWInventoryByLocation, the deletion will fail.

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

