# {{classname}}

All URIs are relative to *https://localhost:8080/cmu/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**AddAll**](NicOperationsApi.md#AddAll) | **Post** /nics | Creates one or multiple new nics
[**Delete**](NicOperationsApi.md#Delete) | **Delete** /nics/{identifier} | Deletes an existing nic
[**DeleteAll**](NicOperationsApi.md#DeleteAll) | **Delete** /nics | Deletes a set of existing nics
[**DeleteAttributes**](NicOperationsApi.md#DeleteAttributes) | **Delete** /nics/{identifier}/attributes | Removes all attributes of an existing nic
[**DeleteGlobalAttribute**](NicOperationsApi.md#DeleteGlobalAttribute) | **Delete** /nics/attributes/{label} | Deletes a global attribute defined for nics
[**Get**](NicOperationsApi.md#Get) | **Get** /nics/{identifier} | Gets one or more nic(s)
[**GetAll**](NicOperationsApi.md#GetAll) | **Get** /nics | Lists all nics
[**GetAttributes**](NicOperationsApi.md#GetAttributes) | **Get** /nics/{identifier}/attributes | Gets all attributes of a single nic
[**GetGlobalAttribute**](NicOperationsApi.md#GetGlobalAttribute) | **Get** /nics/attributes/{label} | Gets a global attribute defined for nics
[**GetGlobalAttributes**](NicOperationsApi.md#GetGlobalAttributes) | **Get** /nics/attributes | Gets all global attributes defined for nics
[**Put**](NicOperationsApi.md#Put) | **Put** /nics/{identifier} | Updates an existing nic
[**PutAll**](NicOperationsApi.md#PutAll) | **Put** /nics | Updates a set of existing nics
[**PutAttributes**](NicOperationsApi.md#PutAttributes) | **Put** /nics/{identifier}/attributes | Adds or modifies attributes of an existing nic
[**PutGlobalAttributes**](NicOperationsApi.md#PutGlobalAttributes) | **Put** /nics/attributes | Adds or modifies global attributes for nics

# **AddAll**
> AddAll(ctx, body, optional)
Creates one or multiple new nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]Nic**](Nic.md)| Nic(s) definition | 
 **optional** | ***NicOperationsApiAddAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiAddAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **networkId** | **optional.**| Attach nic(s) to the given network | 
 **nodeId** | **optional.**| Attach nic(s) to the given node | 
 **computeNextIp** | **optional.**| Compute and set available IP when not defined | [default to false]
 **checkDuplicateIp** | **optional.**| Check for IP duplication | [default to false]
 **checkDuplicateMac** | **optional.**| Check for MAC duplication | [default to false]

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
Deletes an existing nic

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Nic identifier | 

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
Deletes a set of existing nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NicOperationsApiDeleteAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiDeleteAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **body** | [**optional.Interface of MultipleIdentifierDto**](MultipleIdentifierDto.md)| Nics identifier | 
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
Removes all attributes of an existing nic

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Nic identifier | 

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
Deletes a global attribute defined for nics

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
> Nic Get(ctx, identifier, optional)
Gets one or more nic(s)

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**| Nic identifier | 
 **optional** | ***NicOperationsApiGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**Nic**](Nic.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAll**
> []Nic GetAll(ctx, optional)
Lists all nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NicOperationsApiGetAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiGetAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **fields** | [**optional.Interface of []string**](string.md)| Fields to display | 
 **where** | **optional.String**| Filter resources matching provided where clause | 
 **allowsEmpty** | **optional.Bool**| Do not fail when no resource is matched | 

### Return type

[**[]Nic**](Nic.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **GetAttributes**
> map[string]interface{} GetAttributes(ctx, identifier, optional)
Gets all attributes of a single nic

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **identifier** | **string**|  | 
 **optional** | ***NicOperationsApiGetAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiGetAttributesOpts struct
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

# **GetGlobalAttribute**
> string GetGlobalAttribute(ctx, label, optional)
Gets a global attribute defined for nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **label** | **string**|  | 
 **optional** | ***NicOperationsApiGetGlobalAttributeOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiGetGlobalAttributeOpts struct
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

# **GetGlobalAttributes**
> map[string]interface{} GetGlobalAttributes(ctx, optional)
Gets all global attributes defined for nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NicOperationsApiGetGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiGetGlobalAttributesOpts struct
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

# **Put**
> Nic Put(ctx, body, identifier, optional)
Updates an existing nic

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Nic**](Nic.md)| Updated nic definition | 
  **identifier** | **string**| Nic identifier | 
 **optional** | ***NicOperationsApiPutOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiPutOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------


 **computeNextIp** | **optional.**| Compute next free IP in this network and assign it to this nic | [default to false]
 **checkDuplicateIp** | **optional.**| Check for IP duplication | [default to false]
 **checkDuplicateMac** | **optional.**| Check for MAC duplication | [default to false]
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**Nic**](Nic.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAll**
> []Nic PutAll(ctx, body, optional)
Updates a set of existing nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**[]Nic**](Nic.md)| Updated nics definition | 
 **optional** | ***NicOperationsApiPutAllOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiPutAllOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **nameAsId** | **optional.**| Use name instead of UUID as identifier | [default to false]
 **computeNextIp** | **optional.**| Compute next free IP in this network and assign it to this nic | [default to false]
 **checkDuplicateIp** | **optional.**| Check for IP duplication | [default to false]
 **checkDuplicateMac** | **optional.**| Check for MAC duplication | [default to false]
 **where** | **optional.**| Filter resources matching provided where clause | 
 **force** | **optional.**| Force operation when more than one resource is matched | 

### Return type

[**[]Nic**](Nic.md)

### Authorization

[X-Auth-Token](../README.md#X-Auth-Token)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **PutAttributes**
> map[string]interface{} PutAttributes(ctx, body, identifier)
Adds or modifies attributes of an existing nic

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AttributesDto**](AttributesDto.md)| Attributes to be added/modified | 
  **identifier** | **string**| Nic identifier | 

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
Adds or modifies global attributes for nics

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***NicOperationsApiPutGlobalAttributesOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a NicOperationsApiPutGlobalAttributesOpts struct
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

