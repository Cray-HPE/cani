# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvHistByFRUDelete**](HWInventoryHistoryApi.md#DoHWInvHistByFRUDelete) | **Delete** /Inventory/HardwareByFRU/History/{fruid} | Delete history for the HWInventoryByFRU entry with FRU identifier {fruid}
[**DoHWInvHistByFRUGet**](HWInventoryHistoryApi.md#DoHWInvHistByFRUGet) | **Get** /Inventory/HardwareByFRU/History/{fruid} | Retrieve the history entries for the HWInventoryByFRU for {fruid}
[**DoHWInvHistByFRUsGet**](HWInventoryHistoryApi.md#DoHWInvHistByFRUsGet) | **Get** /Inventory/HardwareByFRU/History | Retrieve the history entries for all HWInventoryByFRU entries.
[**DoHWInvHistByLocationDelete**](HWInventoryHistoryApi.md#DoHWInvHistByLocationDelete) | **Delete** /Inventory/Hardware/History/{xname} | DELETE history for the HWInventoryByLocation entry with ID (location) {xname}
[**DoHWInvHistByLocationDeleteAll**](HWInventoryHistoryApi.md#DoHWInvHistByLocationDeleteAll) | **Delete** /Inventory/Hardware/History | Clear the HWInventory history.
[**DoHWInvHistByLocationGet**](HWInventoryHistoryApi.md#DoHWInvHistByLocationGet) | **Get** /Inventory/Hardware/History/{xname} | Retrieve the history entries for the HWInventoryByLocation entry at {xname}
[**DoHWInvHistByLocationsGet**](HWInventoryHistoryApi.md#DoHWInvHistByLocationsGet) | **Get** /Inventory/Hardware/History | Retrieve the history entries for all HWInventoryByLocation entries

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

# **DoHWInvHistByFRUGet**
> HwInventory100HwInventoryHistoryArray DoHWInvHistByFRUGet(ctx, fruid, optional)
Retrieve the history entries for the HWInventoryByFRU for {fruid}

Retrieve the history entries for the HWInventoryByFRU for a specific fruID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **fruid** | **string**| Global HMS field-replaceable (FRU) identifier (serial number, etc.) of the hardware component to select. | 
 **optional** | ***HWInventoryHistoryApiDoHWInvHistByFRUGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryHistoryApiDoHWInvHistByFRUGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **eventtype** | **optional.String**| Retrieve the history entries of a specific type (Added, Removed, etc) for a HWInventoryByFRU entry. | 
 **starttime** | **optional.String**| Retrieve the history entries from after the requested history window start time for a HWInventoryByFRU entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 
 **endtime** | **optional.String**| Retrieve the history entries from before the requested history window end time for a HWInventoryByFRU entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 

### Return type

[**HwInventory100HwInventoryHistoryArray**](HWInventory.1.0.0_HWInventoryHistoryArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvHistByFRUsGet**
> HwInventory100HwInventoryHistoryCollection DoHWInvHistByFRUsGet(ctx, optional)
Retrieve the history entries for all HWInventoryByFRU entries.

Retrieve the history entries for all HWInventoryByFRU entries. Sorted by FRU.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***HWInventoryHistoryApiDoHWInvHistByFRUsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryHistoryApiDoHWInvHistByFRUsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fruid** | **optional.String**| Retrieve the history entries for HWInventoryByFRU entries with the given FRU ID. | 
 **eventtype** | **optional.String**| Retrieve the history entries of a specific type (Added, Removed, etc) for HWInventoryByFRU entries. | 
 **starttime** | **optional.String**| Retrieve the history entries from after the requested history window start time for HWInventoryByFRU entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 
 **endtime** | **optional.String**| Retrieve the history entries from before the requested history window end time for HWInventoryByFRU entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 

### Return type

[**HwInventory100HwInventoryHistoryCollection**](HWInventory.1.0.0_HWInventoryHistoryCollection.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvHistByLocationDelete**
> Response100 DoHWInvHistByLocationDelete(ctx, xname)
DELETE history for the HWInventoryByLocation entry with ID (location) {xname}

Delete history for the HWInventoryByLocation entry for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of HWInventoryByLocation record to delete history for. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvHistByLocationDeleteAll**
> Response100 DoHWInvHistByLocationDeleteAll(ctx, )
Clear the HWInventory history.

Delete all HWInventory history entries. Note that this also deletes history for any associated HWInventoryByFRU entries.

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

# **DoHWInvHistByLocationGet**
> HwInventory100HwInventoryHistoryArray DoHWInvHistByLocationGet(ctx, xname, optional)
Retrieve the history entries for the HWInventoryByLocation entry at {xname}

Retrieve the history entries for a HWInventoryByLocation entry with a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of hardware inventory record to return history for. | 
 **optional** | ***HWInventoryHistoryApiDoHWInvHistByLocationGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryHistoryApiDoHWInvHistByLocationGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **eventtype** | **optional.String**| Retrieve the history entries of a specific type (Added, Removed, etc) for a HWInventoryByLocation entry. | 
 **starttime** | **optional.String**| Retrieve the history entries from after the requested history window start time for a HWInventoryByLocation entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 
 **endtime** | **optional.String**| Retrieve the history entries from before the requested history window end time for a HWInventoryByLocation entry. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 

### Return type

[**HwInventory100HwInventoryHistoryArray**](HWInventory.1.0.0_HWInventoryHistoryArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvHistByLocationsGet**
> HwInventory100HwInventoryHistoryCollection DoHWInvHistByLocationsGet(ctx, optional)
Retrieve the history entries for all HWInventoryByLocation entries

Retrieve the history entries for all HWInventoryByLocation entries.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***HWInventoryHistoryApiDoHWInvHistByLocationsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryHistoryApiDoHWInvHistByLocationsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **optional.String**| Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames. | 
 **eventtype** | **optional.String**| Retrieve the history entries of a specific type (Added, Removed, etc) for HWInventoryByLocation entries. | 
 **starttime** | **optional.String**| Retrieve the history entries from after the requested history window start time for HWInventoryByLocation entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 
 **endtime** | **optional.String**| Retrieve the history entries from before the requested history window end time for HWInventoryByLocation entries. This takes an RFC3339 formatted string (2006-01-02T15:04:05Z07:00). | 

### Return type

[**HwInventory100HwInventoryHistoryCollection**](HWInventory.1.0.0_HWInventoryHistoryCollection.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

