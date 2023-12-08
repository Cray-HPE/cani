# {{classname}}

All URIs are relative to *https://localhost:8080/cmu/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddAll**](NetworkGroupOperationsApi.md#AddAll) | **Post** /networkgroups | Creates one or multiple new network group(s)
[**AddNodes**](NetworkGroupOperationsApi.md#AddNodes) | **Post** /networkgroups/{identifier}/nodes | Adds nodes to an existing group
[**Delete**](NetworkGroupOperationsApi.md#Delete) | **Delete** /networkgroups/{identifier} | Deletes an existing network group
[**DeleteAll**](NetworkGroupOperationsApi.md#DeleteAll) | **Delete** /networkgroups | Deletes a set of existing network groups
[**DeleteAttributes**](NetworkGroupOperationsApi.md#DeleteAttributes) | **Delete** /networkgroups/{identifier}/attributes | Removes all attributes of an existing group
[**DeleteGlobalAttribute**](NetworkGroupOperationsApi.md#DeleteGlobalAttribute) | **Delete** /networkgroups/attributes/{label} | Deletes a global attribute defined for groups
[**Get**](NetworkGroupOperationsApi.md#Get) | **Get** /networkgroups/{identifier} | Gets one or more group(s)
[**GetAll**](NetworkGroupOperationsApi.md#GetAll) | **Get** /networkgroups | Lists all groups
[**GetAttribute**](NetworkGroupOperationsApi.md#GetAttribute) | **Get** /networkgroups/attributes/{label} | Gets a global attribute defined for groups
[**GetAttributes**](NetworkGroupOperationsApi.md#GetAttributes) | **Get** /networkgroups/{identifier}/attributes | Gets all attributes of a single group
[**GetAvailableAction**](NetworkGroupOperationsApi.md#GetAvailableAction) | **Get** /networkgroups/{identifier}/actions | Gets list of available actions on an existing group
[**GetGlobalAttributes**](NetworkGroupOperationsApi.md#GetGlobalAttributes) | **Get** /networkgroups/attributes | Gets all global attributes defined for groups
[**GetNode**](NetworkGroupOperationsApi.md#GetNode) | **Get** /networkgroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
[**GetNodes**](NetworkGroupOperationsApi.md#GetNodes) | **Get** /networkgroups/{identifier}/nodes | Gets all nodes of an existing group
[**Put**](NetworkGroupOperationsApi.md#Put) | **Put** /networkgroups/{identifier} | 
[**PutAll**](NetworkGroupOperationsApi.md#PutAll) | **Put** /networkgroups | 
[**PutAttributes**](NetworkGroupOperationsApi.md#PutAttributes) | **Put** /networkgroups/{identifier}/attributes | Adds or modifies attributes of an existing group
[**PutGlobalAttributes**](NetworkGroupOperationsApi.md#PutGlobalAttributes) | **Put** /networkgroups/attributes | Adds or modifies global attributes for groups
[**RemoveNode**](NetworkGroupOperationsApi.md#RemoveNode) | **Delete** /networkgroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
[**RemoveNodes**](NetworkGroupOperationsApi.md#RemoveNodes) | **Delete** /networkgroups/{identifier}/nodes | Removes some or all nodes from an existing group
[**RunAction**](NetworkGroupOperationsApi.md#RunAction) | **Post** /networkgroups/{identifier}/actions/{action} | Runs an action on a set of existing groups

# **AddAll**
> AddAll(ctx, body)
Creates one or multiple new network group(s)

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]NetworkGroup**](NetworkGroup.md)| Network group(s) definition | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **AddNodes**
> AddNodes(ctx, body, identifier)
Adds nodes to an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**MultipleIdentifierDto**](MultipleIdentifierDto.md)| Nodes identifier | 
  **identifier** | **string**| Group identifier | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Delete**
> Delete(ctx, identifier)
Deletes an existing network group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Network group identifier | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteAll**
> DeleteAll(ctx, optional)
Deletes a set of existing network groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NetworkGroupOperationsApiDeleteAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiDeleteAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of MultipleIdentifierDto**](MultipleIdentifierDto.md)| Network groups identifier | 
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteAttributes**
> DeleteAttributes(ctx, identifier)
Removes all attributes of an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DeleteGlobalAttribute**
> DeleteGlobalAttribute(ctx, label)
Deletes a global attribute defined for groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **label** | **string**|  | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Get**
> Group Get(ctx, identifier, optional)
Gets one or more group(s)

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
 **optional** | ***NetworkGroupOperationsApiGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**Group**](Group.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAll**
> []Group GetAll(ctx, optional)
Lists all groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NetworkGroupOperationsApiGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**[]Group**](Group.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAttribute**
> string GetAttribute(ctx, label, optional)
Gets a global attribute defined for groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **label** | **string**|  | 
 **optional** | ***NetworkGroupOperationsApiGetAttributeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetAttributeOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

**string**

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAttributes**
> map[string]interface{} GetAttributes(ctx, identifier, optional)
Gets all attributes of a single group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
 **optional** | ***NetworkGroupOperationsApiGetAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetAttributesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

[**map[string]interface{}**](interface{}.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAvailableAction**
> GetAvailableAction(ctx, identifier, optional)
Gets list of available actions on an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
 **optional** | ***NetworkGroupOperationsApiGetAvailableActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetAvailableActionOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetGlobalAttributes**
> map[string]interface{} GetGlobalAttributes(ctx, optional)
Gets all global attributes defined for groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NetworkGroupOperationsApiGetGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetGlobalAttributesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

[**map[string]interface{}**](interface{}.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetNode**
> Node GetNode(ctx, identifier, nodeId, optional)
Gets one node of an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
  **nodeId** | **string**| Node identifier | 
 **optional** | ***NetworkGroupOperationsApiGetNodeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetNodeOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 

### Return type

[**Node**](Node.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetNodes**
> []Node GetNodes(ctx, identifier, optional)
Gets all nodes of an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
 **optional** | ***NetworkGroupOperationsApiGetNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiGetNodesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**[]Node**](Node.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **Put**
> NetworkGroup Put(ctx, body, identifier, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**NetworkGroup**](NetworkGroup.md)| Updated network group definition | 
  **identifier** | **string**| Network group identifier | 
 **optional** | ***NetworkGroupOperationsApiPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**NetworkGroup**](NetworkGroup.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAll**
> []NetworkGroup PutAll(ctx, body, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]NetworkGroup**](NetworkGroup.md)| Updated network groups definition | 
 **optional** | ***NetworkGroupOperationsApiPutAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiPutAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nameAsId** | **optional.**| Use name instead of UUID as identifier | [default to false]
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**[]NetworkGroup**](NetworkGroup.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAttributes**
> map[string]interface{} PutAttributes(ctx, body, identifier)
Adds or modifies attributes of an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AttributesDto**](AttributesDto.md)| Attributes to be added/modified | 
  **identifier** | **string**| Group identifier | 

### Return type

[**map[string]interface{}**](interface{}.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutGlobalAttributes**
> PutGlobalAttributes(ctx, optional)
Adds or modifies global attributes for groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NetworkGroupOperationsApiPutGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiPutGlobalAttributesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of AttributesDto**](AttributesDto.md)|  | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveNode**
> RemoveNode(ctx, identifier, nodeId)
Removes one node from an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
  **nodeId** | **string**| Node identifier | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RemoveNodes**
> RemoveNodes(ctx, identifier, optional)
Removes some or all nodes from an existing group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
 **optional** | ***NetworkGroupOperationsApiRemoveNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiRemoveNodesOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **body** | [**optional.Interface of MultipleIdentifierDto**](MultipleIdentifierDto.md)| Nodes identifier, or empty to remove all nodes | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **RunAction**
> RunAction(ctx, identifier, action, optional)
Runs an action on a set of existing groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Group identifier | 
  **action** | **string**| Action | 
 **optional** | ***NetworkGroupOperationsApiRunActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NetworkGroupOperationsApiRunActionOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **body** | [**optional.Interface of map[string]interface{}**](map.md)|  | 

### Return type

 (empty response body)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

