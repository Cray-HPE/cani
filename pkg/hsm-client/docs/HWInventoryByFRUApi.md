# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvByFRUDelete**](HWInventoryByFRUApi.md#DoHWInvByFRUDelete) | **Delete** /Inventory/HardwareByFRU/{fruid} | Delete HWInventoryByFRU entry with FRU identifier {fruid}
[**DoHWInvByFRUDeleteAll**](HWInventoryByFRUApi.md#DoHWInvByFRUDeleteAll) | **Delete** /Inventory/HardwareByFRU | Delete all HWInventoryByFRU entries
[**DoHWInvByFRUGet**](HWInventoryByFRUApi.md#DoHWInvByFRUGet) | **Get** /Inventory/HardwareByFRU/{fruid} | Retrieve HWInventoryByFRU for {fruid}
[**DoHWInvByFRUGetAll**](HWInventoryByFRUApi.md#DoHWInvByFRUGetAll) | **Get** /Inventory/HardwareByFRU | Retrieve all HWInventoryByFRU entries in a flat array

# **DoHWInvByFRUDelete**
> Response100 DoHWInvByFRUDelete(ctx, fruid)
Delete HWInventoryByFRU entry with FRU identifier {fruid}

Delete an entry in the HWInventoryByFRU collection. Note that this does not delete the associated HWInventoryByLocation entry if the FRU is currently residing in the system. In fact, if the FRU ID is associated with a HWInventoryByLocation currently, the deletion will fail.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **fruid** | **string**| Locational xname of HWInventoryByFRU record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

# **DoHWInvByFRUGet**
> HwInventory100HwInventoryByFru DoHWInvByFRUGet(ctx, fruid)
Retrieve HWInventoryByFRU for {fruid}

Retrieve HWInventoryByFRU for a specific fruID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **fruid** | **string**| Global HMS field-replaceable (FRU) identifier (serial number, etc.) of the hardware component to select. | 

### Return type

[**HwInventory100HwInventoryByFru**](HWInventory.1.0.0_HWInventoryByFRU.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvByFRUGetAll**
> []HwInventory100HwInventoryByFru DoHWInvByFRUGetAll(ctx, optional)
Retrieve all HWInventoryByFRU entries in a flat array

Retrieve all HWInventoryByFRU entries. Note that there is no organization of the data, the entries are presented as a flat array. For most purposes, you will want to use /Inventory/Hardware/Query unless you are interested in components that are not currently installed anywhere.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***HWInventoryByFRUApiDoHWInvByFRUGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryByFRUApiDoHWInvByFRUGetAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fruid** | **optional.String**| Retrieve HWInventoryByFRU entries with the given FRU ID. | 
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **manufacturer** | **optional.String**| Retrieve HWInventoryByFRU entries with the given Manufacturer. | 
 **partnumber** | **optional.String**| Retrieve HWInventoryByFRU entries with the given part number. | 
 **serialnumber** | **optional.String**| Retrieve HWInventoryByFRU entries with the given serial number. | 

### Return type

[**[]HwInventory100HwInventoryByFru**](HWInventory.1.0.0_HWInventoryByFRU.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

