# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoHWInvByLocationQueryGet**](HWInventoryApi.md#DoHWInvByLocationQueryGet) | **Get** /Inventory/Hardware/Query/{xname} | Retrieve results of HWInventory query starting at {xname}

# **DoHWInvByLocationQueryGet**
> HwInventory100HwInventory DoHWInvByLocationQueryGet(ctx, xname, optional)
Retrieve results of HWInventory query starting at {xname}

Retrieve zero or more HWInventoryByLocation entries in the form of a HWInventory by providing xname and modifiers in query string. The FRU (field-replaceable unit) data will be included in each HWInventoryByLocation entry if the location is populated.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of parent component, system (e.g. s0, all) or partition (p#.#) to target for hardware inventory | 
 **optional** | ***HWInventoryApiDoHWInvByLocationQueryGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a HWInventoryApiDoHWInvByLocationQueryGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **children** | **optional.Bool**| Also return children of the selected components. Default is true. | 
 **parents** | **optional.Bool**| Also return parents of the selected components. | 
 **partition** | **optional.String**| Restrict search to the given partition (p#.#). Child components are assumed to be in the same partition as the parent component when performing this kind of query. | 
 **format** | **optional.String**| How to display results   FullyFlat      All component types listed in their own                  arrays only.  No nesting of any children.   NestNodesOnly  Flat except that node subcomponents are nested                  hierarchically. Default is NestNodesOnly. | 

### Return type

[**HwInventory100HwInventory**](HWInventory.1.0.0_HWInventory.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

