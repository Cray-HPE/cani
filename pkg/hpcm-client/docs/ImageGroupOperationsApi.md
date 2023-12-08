# {{classname}}

All URIs are relative to *https://localhost:8080/cmu/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddAll**](ImageGroupOperationsApi.md#AddAll) | **Post** /imagegroups | Creates one or multiple new image group(s)
[**AddNodes**](ImageGroupOperationsApi.md#AddNodes) | **Post** /imagegroups/{identifier}/nodes | Adds nodes to an existing group
[**Delete**](ImageGroupOperationsApi.md#Delete) | **Delete** /imagegroups/{identifier} | Deletes an existing image group
[**DeleteAll**](ImageGroupOperationsApi.md#DeleteAll) | **Delete** /imagegroups | Deletes a set of existing image groups
[**DeleteAttributes**](ImageGroupOperationsApi.md#DeleteAttributes) | **Delete** /imagegroups/{identifier}/attributes | Removes all attributes of an existing group
[**DeleteGlobalAttribute**](ImageGroupOperationsApi.md#DeleteGlobalAttribute) | **Delete** /imagegroups/attributes/{label} | Deletes a global attribute defined for groups
[**Get**](ImageGroupOperationsApi.md#Get) | **Get** /imagegroups/{identifier} | Gets one or more group(s)
[**GetAll**](ImageGroupOperationsApi.md#GetAll) | **Get** /imagegroups | Lists all groups
[**GetAttribute**](ImageGroupOperationsApi.md#GetAttribute) | **Get** /imagegroups/attributes/{label} | Gets a global attribute defined for groups
[**GetAttributes**](ImageGroupOperationsApi.md#GetAttributes) | **Get** /imagegroups/{identifier}/attributes | Gets all attributes of a single group
[**GetAvailableAction**](ImageGroupOperationsApi.md#GetAvailableAction) | **Get** /imagegroups/{identifier}/actions | Gets list of available actions on an existing group
[**GetGlobalAttributes**](ImageGroupOperationsApi.md#GetGlobalAttributes) | **Get** /imagegroups/attributes | Gets all global attributes defined for groups
[**GetNode**](ImageGroupOperationsApi.md#GetNode) | **Get** /imagegroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
[**GetNodes**](ImageGroupOperationsApi.md#GetNodes) | **Get** /imagegroups/{identifier}/nodes | Gets all nodes of an existing group
[**Put**](ImageGroupOperationsApi.md#Put) | **Put** /imagegroups/{identifier} | 
[**PutAll**](ImageGroupOperationsApi.md#PutAll) | **Put** /imagegroups | 
[**PutAttributes**](ImageGroupOperationsApi.md#PutAttributes) | **Put** /imagegroups/{identifier}/attributes | Adds or modifies attributes of an existing group
[**PutGlobalAttributes**](ImageGroupOperationsApi.md#PutGlobalAttributes) | **Put** /imagegroups/attributes | Adds or modifies global attributes for groups
[**RemoveNode**](ImageGroupOperationsApi.md#RemoveNode) | **Delete** /imagegroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
[**RemoveNodes**](ImageGroupOperationsApi.md#RemoveNodes) | **Delete** /imagegroups/{identifier}/nodes | Removes some or all nodes from an existing group
[**RunAction**](ImageGroupOperationsApi.md#RunAction) | **Post** /imagegroups/{identifier}/actions/{action} | Runs an action on a set of existing groups

# **AddAll**
> AddAll(ctx, body)
Creates one or multiple new image group(s)

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]ImageGroup**](ImageGroup.md)| Image group(s) definition | 

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
Deletes an existing image group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Image group identifier | 

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
Deletes a set of existing image groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ImageGroupOperationsApiDeleteAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiDeleteAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of MultipleIdentifierDto**](MultipleIdentifierDto.md)| Image groups identifier | 
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
 **optional** | ***ImageGroupOperationsApiGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetAllOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetAttributeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetAttributeOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetAttributesOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetAvailableActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetAvailableActionOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetGlobalAttributesOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetNodeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetNodeOpts struct
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
 **optional** | ***ImageGroupOperationsApiGetNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiGetNodesOpts struct
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
> ImageGroup Put(ctx, body, identifier, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ImageGroup**](ImageGroup.md)| Updated image group definition | 
  **identifier** | **string**| Image group identifier | 
 **optional** | ***ImageGroupOperationsApiPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**ImageGroup**](ImageGroup.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAll**
> []ImageGroup PutAll(ctx, body, optional)


### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]ImageGroup**](ImageGroup.md)| Updated image groups definition | 
 **optional** | ***ImageGroupOperationsApiPutAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiPutAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nameAsId** | **optional.**| Use name instead of UUID as identifier | [default to false]
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**[]ImageGroup**](ImageGroup.md)

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
 **optional** | ***ImageGroupOperationsApiPutGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiPutGlobalAttributesOpts struct
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
 **optional** | ***ImageGroupOperationsApiRemoveNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiRemoveNodesOpts struct
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
 **optional** | ***ImageGroupOperationsApiRunActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ImageGroupOperationsApiRunActionOpts struct
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

