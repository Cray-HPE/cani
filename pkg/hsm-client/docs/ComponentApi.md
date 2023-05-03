# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoCompArrayNIDPatch**](ComponentApi.md#DoCompArrayNIDPatch) | **Patch** /State/Components/BulkNID | Update multiple components&#x27; NIDs via ComponentArray
[**DoCompBulkEnabledPatch**](ComponentApi.md#DoCompBulkEnabledPatch) | **Patch** /State/Components/BulkEnabled | Update multiple components&#x27; Enabled values via a list of xnames
[**DoCompBulkFlagOnlyPatch**](ComponentApi.md#DoCompBulkFlagOnlyPatch) | **Patch** /State/Components/BulkFlagOnly | Update multiple components&#x27; Flag values via a list of xnames
[**DoCompBulkRolePatch**](ComponentApi.md#DoCompBulkRolePatch) | **Patch** /State/Components/BulkRole | Update multiple components&#x27; Role values via a list of xnames
[**DoCompBulkStateDataPatch**](ComponentApi.md#DoCompBulkStateDataPatch) | **Patch** /State/Components/BulkStateData | Update multiple components&#x27; state data via a list of xnames
[**DoCompBulkSwStatusPatch**](ComponentApi.md#DoCompBulkSwStatusPatch) | **Patch** /State/Components/BulkSoftwareStatus | Update multiple components&#x27; SoftwareStatus values via a list of xnames
[**DoCompEnabledPatch**](ComponentApi.md#DoCompEnabledPatch) | **Patch** /State/Components/{xname}/Enabled | Update component Enabled value at {xname}
[**DoCompFlagOnlyPatch**](ComponentApi.md#DoCompFlagOnlyPatch) | **Patch** /State/Components/{xname}/FlagOnly | Update component Flag value at {xname}
[**DoCompNIDPatch**](ComponentApi.md#DoCompNIDPatch) | **Patch** /State/Components/{xname}/NID | Update component NID value at {xname}
[**DoCompRolePatch**](ComponentApi.md#DoCompRolePatch) | **Patch** /State/Components/{xname}/Role | Update component Role and SubRole values at {xname}
[**DoCompStatePatch**](ComponentApi.md#DoCompStatePatch) | **Patch** /State/Components/{xname}/StateData | Update component state data at {xname}
[**DoCompSwStatusPatch**](ComponentApi.md#DoCompSwStatusPatch) | **Patch** /State/Components/{xname}/SoftwareStatus | Update component SoftwareStatus value at {xname}
[**DoComponentByNIDGet**](ComponentApi.md#DoComponentByNIDGet) | **Get** /State/Components/ByNID/{nid} | Retrieve component with NID&#x3D;{nid}
[**DoComponentByNIDQueryPost**](ComponentApi.md#DoComponentByNIDQueryPost) | **Post** /State/Components/ByNID/Query | Create component query (by NID ranges), returning ComponentArray
[**DoComponentDelete**](ComponentApi.md#DoComponentDelete) | **Delete** /State/Components/{xname} | Delete component with ID {xname}
[**DoComponentGet**](ComponentApi.md#DoComponentGet) | **Get** /State/Components/{xname} | Retrieve component at {xname}
[**DoComponentPut**](ComponentApi.md#DoComponentPut) | **Put** /State/Components/{xname} | Create/Update an HMS Component
[**DoComponentQueryGet**](ComponentApi.md#DoComponentQueryGet) | **Get** /State/Components/Query/{xname} | Retrieve component query for {xname}, returning ComponentArray
[**DoComponentsDeleteAll**](ComponentApi.md#DoComponentsDeleteAll) | **Delete** /State/Components | Delete all components
[**DoComponentsGet**](ComponentApi.md#DoComponentsGet) | **Get** /State/Components | Retrieve collection of HMS Components
[**DoComponentsPost**](ComponentApi.md#DoComponentsPost) | **Post** /State/Components | Create/Update a collection of HMS Components
[**DoComponentsQueryPost**](ComponentApi.md#DoComponentsQueryPost) | **Post** /State/Components/Query | Create component query (by xname list), returning ComponentArray

# **DoCompArrayNIDPatch**
> DoCompArrayNIDPatch(ctx, body)
Update multiple components' NIDs via ComponentArray

Modify the submitted ComponentArray and update the corresponding NID value for each entry. Other fields are ignored and not changed. ID field is required for all entries.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArrayNid**](ComponentArrayPatchArrayNid.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompBulkEnabledPatch**
> DoCompBulkEnabledPatch(ctx, body)
Update multiple components' Enabled values via a list of xnames

Update the Enabled field for a list of xnames. Specify a single value for Enabled and also the list of xnames. Note that Enabled is a boolean field and a value of false sets the component(s) to disabled.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArrayEnabled**](ComponentArrayPatchArrayEnabled.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompBulkFlagOnlyPatch**
> DoCompBulkFlagOnlyPatch(ctx, body)
Update multiple components' Flag values via a list of xnames

Specify a list of xnames to update the Flag field and specify the value. The list of IDs and the new Flag are required.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArrayFlagOnly**](ComponentArrayPatchArrayFlagOnly.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompBulkRolePatch**
> DoCompBulkRolePatch(ctx, body)
Update multiple components' Role values via a list of xnames

Update the Role and SubRole field for a list of xnames. Specify the Role and Subrole values and the list of xnames. The list of IDs and the new Role are required.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArrayRole**](ComponentArrayPatchArrayRole.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompBulkStateDataPatch**
> DoCompBulkStateDataPatch(ctx, body)
Update multiple components' state data via a list of xnames

Specify a list of xnames to update the State and Flag fields. If the Flag field is omitted, Flag is reverted to 'OK'. Other fields are ignored. The list of IDs and the new State are required.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArrayStateData**](ComponentArrayPatchArrayStateData.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompBulkSwStatusPatch**
> DoCompBulkSwStatusPatch(ctx, body)
Update multiple components' SoftwareStatus values via a list of xnames

Update the SoftwareStatus field for a list of xnames. Specify a single new value of SoftwareStatus like admindown and the list of xnames.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPatchArraySoftwareStatus**](ComponentArrayPatchArraySoftwareStatus.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompEnabledPatch**
> DoCompEnabledPatch(ctx, body, xname)
Update component Enabled value at {xname}

Update the component's Enabled field only. The State and other fields are not modified. Note that this is a boolean field, a value of false sets the component to disabled.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchEnabled**](Component100PatchEnabled.md)|  | 
  **xname** | **string**| Locational xname of component to set Enabled to true or false. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompFlagOnlyPatch**
> DoCompFlagOnlyPatch(ctx, body, xname)
Update component Flag value at {xname}

The State is not modified. Only the Flag is updated.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchFlagOnly**](Component100PatchFlagOnly.md)|  | 
  **xname** | **string**| Locational xname of component to modify flag on. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompNIDPatch**
> DoCompNIDPatch(ctx, body, xname)
Update component NID value at {xname}

Update the component's NID field only. Valid only for nodes. State and other fields are not modified.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchNid**](Component100PatchNid.md)|  | 
  **xname** | **string**| Locational xname of component to modify NID on. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompRolePatch**
> DoCompRolePatch(ctx, body, xname)
Update component Role and SubRole values at {xname}

Update the component's Role and SubRole fields only. Valid only for nodes. The State and other fields are not modified.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchRole**](Component100PatchRole.md)|  | 
  **xname** | **string**| Locational xname of component to modify Role on. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompStatePatch**
> DoCompStatePatch(ctx, body, xname)
Update component state data at {xname}

Update the component's state and flag fields only. If Flag field is omitted, the Flag value is reverted to 'OK'.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchStateData**](Component100PatchStateData.md)|  | 
  **xname** | **string**| Locational xname of component to set state/flag on. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoCompSwStatusPatch**
> DoCompSwStatusPatch(ctx, body, xname)
Update component SoftwareStatus value at {xname}

Update the component's SoftwareStatus field only. The State and other fields are not modified.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100PatchSoftwareStatus**](Component100PatchSoftwareStatus.md)|  | 
  **xname** | **string**| Locational xname of component to set new SoftwareStatus value. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentByNIDGet**
> Component100Component DoComponentByNIDGet(ctx, nid)
Retrieve component with NID={nid}

Retrieve a component by NID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **nid** | **string**| NID of component to return. | 

### Return type

[**Component100Component**](Component.1.0.0_Component.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentByNIDQueryPost**
> ComponentArrayComponentArray DoComponentByNIDQueryPost(ctx, body)
Create component query (by NID ranges), returning ComponentArray

Retrieve the targeted entries in the form of a ComponentArray by providing a payload of NID ranges.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPostByNidQuery**](ComponentArrayPostByNidQuery.md)|  | 

### Return type

[**ComponentArrayComponentArray**](ComponentArray_ComponentArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentDelete**
> Response100 DoComponentDelete(ctx, xname)
Delete component with ID {xname}

Delete a component by xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of component record to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentGet**
> Component100Component DoComponentGet(ctx, xname)
Retrieve component at {xname}

Retrieve state or components by xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of component to return. | 

### Return type

[**Component100Component**](Component.1.0.0_Component.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentPut**
> DoComponentPut(ctx, body, xname)
Create/Update an HMS Component

Create/Update a state/component. If the component already exists it will not be overwritten unless force=true in which case State, Flag, Subtype, NetType, Arch, and Class will get overwritten.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Component100Put**](Component100Put.md)|  | 
  **xname** | **string**| Locational xname of the component to create or update. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentQueryGet**
> ComponentArrayComponentArray DoComponentQueryGet(ctx, xname, optional)
Retrieve component query for {xname}, returning ComponentArray

Retrieve component entries in the form of a ComponentArray by providing xname and modifiers in the query string.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Locational xname of component to query. | 
 **optional** | ***ComponentApiDoComponentQueryGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ComponentApiDoComponentQueryGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **type_** | **optional.String**| Retrieve xname&#x27;s children of type&#x3D;{type} instead of {xname} for example NodeBMC, NodeEnclosure etc. | 
 **state** | **optional.String**| Filter the results based on HMS state like Ready, On etc. Can be specified multiple times for selecting entries in different states. | 
 **flag** | **optional.String**| Filter the results based on HMS flag value like OK, Alert etc. Can be specified multiple times for selecting entries with different flags. | 
 **role** | **optional.String**| Filter the results based on HMS role. Can be specified multiple times for selecting entries with different roles. Valid values are: - Compute - Service - System - Application - Storage - Management Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/role&#x27; for the complete list. | 
 **subrole** | **optional.String**| Filter the results based on HMS subrole. Can be specified multiple times for selecting entries with different subroles. Valid values are: - Master - Worker - Storage Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/subrole&#x27; for the complete list. | 
 **enabled** | **optional.String**| Filter the results based on enabled status (true or false). | 
 **softwarestatus** | **optional.String**| Filter the results based on software status. Software status is a free form string. Matching is case-insensitive. Can be specified multiple times for selecting entries with different software statuses. | 
 **subtype** | **optional.String**| Filter the results based on HMS subtype. Can be specified multiple times for selecting entries with different subtypes. | 
 **arch** | **optional.String**| Filter the results based on architecture. Can be specified multiple times for selecting components with different architectures. | 
 **class** | **optional.String**| Filter the results based on HMS hardware class. Can be specified multiple times for selecting entries with different classes. | 
 **nid** | **optional.String**| Filter the results based on NID. Can be specified multiple times for selecting entries with multiple specific NIDs. | 
 **nidStart** | **optional.String**| Filter the results based on NIDs equal to or greater than the provided integer. | 
 **nidEnd** | **optional.String**| Filter the results based on NIDs less than or equal to the provided integer. | 
 **partition** | **optional.String**| Restrict search to the given partition (p#.#). One partition can be combined with at most one group argument which will be treated as a logical AND. NULL will return components in NO partition. | 
 **group** | **optional.String**| Restrict search to the given group label. One group can be combined with at most one partition argument which will be treated as a logical AND. NULL will return components in NO groups. | 
 **stateonly** | **optional.Bool**| Return only component state and flag fields (plus xname/ID and type). Results can be modified and used for bulk state/flag- only patch operations. | 
 **flagonly** | **optional.Bool**| Return only component flag field (plus xname/ID and type). Results can be modified and used for bulk flag-only patch operations. | 
 **roleonly** | **optional.Bool**| Return only component role and subrole fields (plus xname/ID and type). Results can be modified and used for bulk role-only patches. | 
 **nidonly** | **optional.Bool**| Return only component NID field (plus xname/ID and type). Results can be modified and used for bulk NID-only patches. | 

### Return type

[**ComponentArrayComponentArray**](ComponentArray_ComponentArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentsDeleteAll**
> Response100 DoComponentsDeleteAll(ctx, )
Delete all components

Delete all entries in the components collection.

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

# **DoComponentsGet**
> ComponentArrayComponentArray DoComponentsGet(ctx, optional)
Retrieve collection of HMS Components

Retrieve the full collection of state/components in the form of a ComponentArray. Full results can also be filtered by query parameters. When multiple parameters are specified, they are applied in an AND fashion (e.g. type AND state). When a parameter is specified multiple times, they are applied in an OR fashion (e.g. type AND state1 OR state2). If the collection is empty or the filters have no match, an empty array is returned.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***ComponentApiDoComponentsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a ComponentApiDoComponentsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **optional.String**| Filter the results based on xname ID(s). Can be specified multiple times for selecting entries with multiple specific xnames. | 
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **state** | **optional.String**| Filter the results based on HMS state like Ready, On etc. Can be specified multiple times for selecting entries in different states. | 
 **flag** | **optional.String**| Filter the results based on HMS flag value like OK, Alert etc. Can be specified multiple times for selecting entries with different flags. | 
 **role** | **optional.String**| Filter the results based on HMS role. Can be specified multiple times for selecting entries with different roles. Valid values are: - Compute - Service - System - Application - Storage - Management Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/role&#x27; for the complete list. | 
 **subrole** | **optional.String**| Filter the results based on HMS subrole. Can be specified multiple times for selecting entries with different subroles. Valid values are: - Master - Worker - Storage Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/subrole&#x27; for the complete list. | 
 **enabled** | **optional.String**| Filter the results based on enabled status (true or false). | 
 **softwarestatus** | **optional.String**| Filter the results based on software status. Software status is a free form string. Matching is case-insensitive. Can be specified multiple times for selecting entries with different software statuses. | 
 **subtype** | **optional.String**| Filter the results based on HMS subtype. Can be specified multiple times for selecting entries with different subtypes. | 
 **arch** | **optional.String**| Filter the results based on architecture. Can be specified multiple times for selecting components with different architectures. | 
 **class** | **optional.String**| Filter the results based on HMS hardware class. Can be specified multiple times for selecting entries with different classes. | 
 **nid** | **optional.String**| Filter the results based on NID. Can be specified multiple times for selecting entries with multiple specific NIDs. | 
 **nidStart** | **optional.String**| Filter the results based on NIDs equal to or greater than the provided integer. | 
 **nidEnd** | **optional.String**| Filter the results based on NIDs less than or equal to the provided integer. | 
 **partition** | **optional.String**| Restrict search to the given partition (p#.#). One partition can be combined with at most one group argument which will be treated as a logical AND. NULL will return components in NO partition. | 
 **group** | **optional.String**| Restrict search to the given group label. One group can be combined with at most one partition argument which will be treated as a logical AND. NULL will return components in NO groups. | 
 **stateonly** | **optional.Bool**| Return only component state and flag fields (plus xname/ID and type). Results can be modified and used for bulk state/flag- only patch operations. | 
 **flagonly** | **optional.Bool**| Return only component flag field (plus xname/ID and type). Results can be modified and used for bulk flag-only patch operations. | 
 **roleonly** | **optional.Bool**| Return only component role and subrole fields (plus xname/ID and type). Results can be modified and used for bulk role-only patches. | 
 **nidonly** | **optional.Bool**| Return only component NID field (plus xname/ID and type). Results can be modified and used for bulk NID-only patches. | 

### Return type

[**ComponentArrayComponentArray**](ComponentArray_ComponentArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentsPost**
> DoComponentsPost(ctx, body)
Create/Update a collection of HMS Components

Create/Update a collection of state/components. If the component already exists it will not be overwritten unless force=true in which case State, Flag, Subtype, NetType, Arch, and Class will get overwritten.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPostArray**](ComponentArrayPostArray.md)|  | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoComponentsQueryPost**
> ComponentArrayComponentArray DoComponentsQueryPost(ctx, body)
Create component query (by xname list), returning ComponentArray

Retrieve the targeted entries in the form of a ComponentArray by providing a payload of component IDs.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ComponentArrayPostQuery**](ComponentArrayPostQuery.md)|  | 

### Return type

[**ComponentArrayComponentArray**](ComponentArray_ComponentArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

