# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvByLocationDelete**](HWInventoryByLocationApi.md#DoHWInvByLocationDelete) | **Delete** /Inventory/Hardware/{xname} | DELETE HWInventoryByLocation entry with ID (location) {xname}
[**DoHWInvByLocationDeleteAll**](HWInventoryByLocationApi.md#DoHWInvByLocationDeleteAll) | **Delete** /Inventory/Hardware | Delete all HWInventoryByLocation entries
[**DoHWInvByLocationGet**](HWInventoryByLocationApi.md#DoHWInvByLocationGet) | **Get** /Inventory/Hardware/{xname} | Retrieve HWInventoryByLocation entry at {xname}
[**DoHWInvByLocationGetAll**](HWInventoryByLocationApi.md#DoHWInvByLocationGetAll) | **Get** /Inventory/Hardware | Retrieve all HWInventoryByLocation entries in array
[**DoHWInvByLocationPost**](HWInventoryByLocationApi.md#DoHWInvByLocationPost) | **Post** /Inventory/Hardware | Create/Update hardware inventory entries

# **DoHWInvByLocationDelete**
> Response100 DoHWInvByLocationDelete(ctx, xname)
DELETE HWInventoryByLocation entry with ID (location) {xname}

Delete HWInventoryByLocation entry for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of HWInventoryByLocation record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvByLocationDeleteAll**
> Response100 DoHWInvByLocationDeleteAll(ctx, )
Delete all HWInventoryByLocation entries

Delete all entries in the HWInventoryByLocation collection. Note that this does not delete any associated HWInventoryByFRU entries.

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

# **DoHWInvByLocationGet**
> HwInventory100HwInventoryByLocation DoHWInvByLocationGet(ctx, xname)
Retrieve HWInventoryByLocation entry at {xname}

Retrieve HWInventoryByLocation entries for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of hardware inventory record to return. | 

### Return type

[**HwInventory100HwInventoryByLocation**](HWInventory.1.0.0_HWInventoryByLocation.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvByLocationGetAll**
> []HwInventory100HwInventoryByLocation DoHWInvByLocationGetAll(ctx, optional)
Retrieve all HWInventoryByLocation entries in array

Retrieve all HWInventoryByLocation entries. Note that all entries are displayed as a flat array. For most purposes, you will want to use /Inventory/Hardware/Query.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***HWInventoryByLocationApiDoHWInvByLocationGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryByLocationApiDoHWInvByLocationGetAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **optional.String**| Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames. | 
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **manufacturer** | **optional.String**| Retrieve HWInventoryByLocation entries with the given Manufacturer. | 
 **partnumber** | **optional.String**| Retrieve HWInventoryByLocation entries with the given part number. | 
 **serialnumber** | **optional.String**| Retrieve HWInventoryByLocation entries with the given serial number. | 
 **fruid** | **optional.String**| Retrieve HWInventoryByLocation entries with the given FRU ID. | 

### Return type

[**[]HwInventory100HwInventoryByLocation**](HWInventory.1.0.0_HWInventoryByLocation.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvByLocationPost**
> Response100 DoHWInvByLocationPost(ctx, body)
Create/Update hardware inventory entries

Create/Update hardware inventory entries

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**InventoryHardwareBody**](InventoryHardwareBody.md)|  | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

