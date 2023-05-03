# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoInventoryDiscoverPost**](DiscoverApi.md#DoInventoryDiscoverPost) | **Post** /Inventory/Discover | Create Discover operation request

# **DoInventoryDiscoverPost**
> []ResourceUri100 DoInventoryDiscoverPost(ctx, optional)
Create Discover operation request

Discover and populate database with component data (ComponentEndpoints, HMS Components, HWInventory) based on interrogating RedfishEndpoint entries.  If not all RedfishEndpoints should be discovered, an array of xnames can be provided in the DiscoverInput payload.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***DiscoverApiDoInventoryDiscoverPostOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a DiscoverApiDoInventoryDiscoverPostOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of Discover100DiscoverInput**](Discover100DiscoverInput.md)|  | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

