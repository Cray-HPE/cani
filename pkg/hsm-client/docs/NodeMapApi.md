# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoNodeMapDelete**](NodeMapApi.md#DoNodeMapDelete) | **Delete** /Defaults/NodeMaps/{xname} | Delete NodeMap with ID {xname}
[**DoNodeMapGet**](NodeMapApi.md#DoNodeMapGet) | **Get** /Defaults/NodeMaps/{xname} | Retrieve NodeMap at {xname}
[**DoNodeMapPost**](NodeMapApi.md#DoNodeMapPost) | **Post** /Defaults/NodeMaps | Create or Modify NodeMaps
[**DoNodeMapPut**](NodeMapApi.md#DoNodeMapPut) | **Put** /Defaults/NodeMaps/{xname} | Update definition for NodeMap ID {xname}
[**DoNodeMapsDeleteAll**](NodeMapApi.md#DoNodeMapsDeleteAll) | **Delete** /Defaults/NodeMaps | Delete all NodeMap entities
[**DoNodeMapsGet**](NodeMapApi.md#DoNodeMapsGet) | **Get** /Defaults/NodeMaps | Retrieve all NodeMaps, returning NodeMapArray

# **DoNodeMapDelete**
> Response100 DoNodeMapDelete(ctx, xname)
Delete NodeMap with ID {xname}

Delete NodeMap entry for a specific node {xname}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of NodeMap record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoNodeMapGet**
> NodeMap100NodeMap DoNodeMapGet(ctx, xname)
Retrieve NodeMap at {xname}

Retrieve NodeMap, i.e. defaults NID/Role/etc. for node located at physical location {xname}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of NodeMap record to return. | 

### Return type

[**NodeMap100NodeMap**](NodeMap.1.0.0_NodeMap.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoNodeMapPost**
> Response100 DoNodeMapPost(ctx, body)
Create or Modify NodeMaps

Create or update the given set of NodeMaps whose ID fields are each a valid xname. The NID field is required and serves as the NID that will be used when a component with the same xname ID is created for the first time by discovery. Role is an optional field. A node is assigned the default (e.g. Compute) role when it is first created during discovery. The NID must be unique across all entries. SubRole is an optional field. A node is assigned no subrole by default when it is first created during discovery.  The NodeMaps collection should be uploaded at install time by specifying it as a JSON file. As a result, when the endpoints are automatically discovered by REDS, and inventory discovery is performed by HSM, the desired NID numbers will be set as soon as the nodes are created using the NodeMaps collection. All node xnames that are expected to be used in the system should be included in the mapping, even if not currently populated.  It is recommended that NodeMaps are uploaded at install time before discovery happens. If they are uploaded after discovery, then the node xnames need to be manually updated with the correct NIDs. You can update NIDs for individual components by using PATCH /State/Components/{xname}/NID.  Note the following points: * If the POST operation contains an xname that already exists, the entry will be overwritten with the new entry (i.e. new NID, Role (if given), etc.). * The same NID cannot be used for more than one xname. If such a duplicate would be created, the operation will fail. * If the node has already been discovered for the first time (that is, it exists in /hsm/v2/State/Components and already has a previous/default NID), modifying the NodeMap entry will not automatically reassign the current NID. * If you wish to use POST to completely replace the current NodeMaps collection (rather than modifying it), first delete it using the DELETE method on the collection. Otherwise the current entries and the new ones will be merged if they are disjoint sets of nodes.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NodeMapArrayNodeMapArray**](NodeMapArrayNodeMapArray.md)|  | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoNodeMapPut**
> NodeMap100PostNodeMap DoNodeMapPut(ctx, body, xname)
Update definition for NodeMap ID {xname}

Update or create an entry for an individual node xname using PUT. Note the following points: * If the PUT operation contains an xname that already exists, the entry will be overwritten with the new entry (i.e. new NID, Role (if given), etc.). * The same NID cannot be used for more than one xname. If such a duplicate would be created, the operation will fail. * If the node has already been discovered for the first time (that is, it exists in /hsm/v2/State/Components and already has a previous/default NID), modifying the NodeMap entry will not automatically reassign the current NID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NodeMap100NodeMap**](NodeMap100NodeMap.md)|  | 
  **xname** | **string**| Locational xname of NodeMap record to create or update. | 

### Return type

[**NodeMap100PostNodeMap**](NodeMap.1.0.0_PostNodeMap.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoNodeMapsDeleteAll**
> Response100 DoNodeMapsDeleteAll(ctx, )
Delete all NodeMap entities

Delete all entries in the NodeMaps collection.

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

# **DoNodeMapsGet**
> NodeMapArrayNodeMapArray DoNodeMapsGet(ctx, )
Retrieve all NodeMaps, returning NodeMapArray

Retrieve all Node map entries as a named array, or an empty array if the collection is empty.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**NodeMapArrayNodeMapArray**](NodeMapArray_NodeMapArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

