# {{classname}}

All URIs are relative to *https://localhost:8080/cmu/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddAll**](SystemGroupOperationsApi.md#AddAll) | **Post** /systemgroups | Creates one or multiple new system group(s)
[**AddNodes**](SystemGroupOperationsApi.md#AddNodes) | **Post** /systemgroups/{identifier}/nodes | Adds nodes to an existing group
[**Delete**](SystemGroupOperationsApi.md#Delete) | **Delete** /systemgroups/{identifier} | Deletes an existing system group
[**DeleteAll**](SystemGroupOperationsApi.md#DeleteAll) | **Delete** /systemgroups | Deletes a set of existing system groups
[**DeleteAttributes**](SystemGroupOperationsApi.md#DeleteAttributes) | **Delete** /systemgroups/{identifier}/attributes | Removes all attributes of an existing group
[**DeleteGlobalAttribute**](SystemGroupOperationsApi.md#DeleteGlobalAttribute) | **Delete** /systemgroups/attributes/{label} | Deletes a global attribute defined for groups
[**Get**](SystemGroupOperationsApi.md#Get) | **Get** /systemgroups/{identifier} | Gets one or more group(s)
[**GetAll**](SystemGroupOperationsApi.md#GetAll) | **Get** /systemgroups | Lists all groups
[**GetAttribute**](SystemGroupOperationsApi.md#GetAttribute) | **Get** /systemgroups/attributes/{label} | Gets a global attribute defined for groups
[**GetAttributes**](SystemGroupOperationsApi.md#GetAttributes) | **Get** /systemgroups/{identifier}/attributes | Gets all attributes of a single group
[**GetAvailableAction**](SystemGroupOperationsApi.md#GetAvailableAction) | **Get** /systemgroups/{identifier}/actions | Gets list of available actions on an existing group
[**GetGlobalAttributes**](SystemGroupOperationsApi.md#GetGlobalAttributes) | **Get** /systemgroups/attributes | Gets all global attributes defined for groups
[**GetNode**](SystemGroupOperationsApi.md#GetNode) | **Get** /systemgroups/{identifier}/nodes/{node_id} | Gets one node of an existing group
[**GetNodes**](SystemGroupOperationsApi.md#GetNodes) | **Get** /systemgroups/{identifier}/nodes | Gets all nodes of an existing group
[**Put**](SystemGroupOperationsApi.md#Put) | **Put** /systemgroups/{identifier} | Updates a existing system group
[**PutAll**](SystemGroupOperationsApi.md#PutAll) | **Put** /systemgroups | Updates a set of existing system groups
[**PutAttributes**](SystemGroupOperationsApi.md#PutAttributes) | **Put** /systemgroups/{identifier}/attributes | Adds or modifies attributes of an existing group
[**PutGlobalAttributes**](SystemGroupOperationsApi.md#PutGlobalAttributes) | **Put** /systemgroups/attributes | Adds or modifies global attributes for groups
[**RemoveNode**](SystemGroupOperationsApi.md#RemoveNode) | **Delete** /systemgroups/{identifier}/nodes/{node_id} | Removes one node from an existing group
[**RemoveNodes**](SystemGroupOperationsApi.md#RemoveNodes) | **Delete** /systemgroups/{identifier}/nodes | Removes some or all nodes from an existing group
[**RunAction**](SystemGroupOperationsApi.md#RunAction) | **Post** /systemgroups/{identifier}/actions/{action} | Runs an action on a set of existing groups

# **AddAll**
> AddAll(ctx, body)
Creates one or multiple new system group(s)

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]SystemGroup**](SystemGroup.md)| System group(s) definition | 

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
Deletes an existing system group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| System group identifier | 

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
Deletes a set of existing system groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***SystemGroupOperationsApiDeleteAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiDeleteAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of MultipleIdentifierDto**](MultipleIdentifierDto.md)| System groups identifier | 
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
 **optional** | ***SystemGroupOperationsApiGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetAllOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetAttributeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetAttributeOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetAttributesOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetAvailableActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetAvailableActionOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetGlobalAttributesOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetNodeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetNodeOpts struct
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
 **optional** | ***SystemGroupOperationsApiGetNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiGetNodesOpts struct
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
> SystemGroup Put(ctx, body, identifier, optional)
Updates a existing system group

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SystemGroup**](SystemGroup.md)| Updated system group definition | 
  **identifier** | **string**| System group identifier | 
 **optional** | ***SystemGroupOperationsApiPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**SystemGroup**](SystemGroup.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAll**
> []SystemGroup PutAll(ctx, body, optional)
Updates a set of existing system groups

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]SystemGroup**](SystemGroup.md)| Updated system groups definition | 
 **optional** | ***SystemGroupOperationsApiPutAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiPutAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nameAsId** | **optional.**| Use name instead of UUID as identifier | [default to false]
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**[]SystemGroup**](SystemGroup.md)

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
 **optional** | ***SystemGroupOperationsApiPutGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiPutGlobalAttributesOpts struct
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
 **optional** | ***SystemGroupOperationsApiRemoveNodesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiRemoveNodesOpts struct
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
 **optional** | ***SystemGroupOperationsApiRunActionOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a SystemGroupOperationsApiRunActionOpts struct
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

