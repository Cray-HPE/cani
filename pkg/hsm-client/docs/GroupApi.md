# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoGroupDelete**](GroupApi.md#DoGroupDelete) | **Delete** /groups/{group_label} | Delete existing group with {group_label}
[**DoGroupGet**](GroupApi.md#DoGroupGet) | **Get** /groups/{group_label} | Retrieve existing group {group_label}
[**DoGroupLabelsGet**](GroupApi.md#DoGroupLabelsGet) | **Get** /groups/labels | Retrieve all existing group labels
[**DoGroupMemberDelete**](GroupApi.md#DoGroupMemberDelete) | **Delete** /groups/{group_label}/members/{xname_id} | Delete member from existing group
[**DoGroupMembersGet**](GroupApi.md#DoGroupMembersGet) | **Get** /groups/{group_label}/members | Retrieve all members of existing group
[**DoGroupMembersPost**](GroupApi.md#DoGroupMembersPost) | **Post** /groups/{group_label}/members | Create new member of existing group (via POST)
[**DoGroupPatch**](GroupApi.md#DoGroupPatch) | **Patch** /groups/{group_label} | Update metadata for existing group {group_label}
[**DoGroupsGet**](GroupApi.md#DoGroupsGet) | **Get** /groups | Retrieve all existing groups
[**DoGroupsPost**](GroupApi.md#DoGroupsPost) | **Post** /groups | Create a new group

# **DoGroupDelete**
> Response100 DoGroupDelete(ctx, groupLabel)
Delete existing group with {group_label}

Delete the given group with {group_label}. Any members previously in the group will no longer have the deleted group label associated with them.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **groupLabel** | **string**| Label (i.e. name) of the group to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupGet**
> Group100 DoGroupGet(ctx, groupLabel, optional)
Retrieve existing group {group_label}

Retrieve the group which was created with the given {group_label}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **groupLabel** | **string**| Label name of the group to return. | 
 **optional** | ***GroupApiDoGroupGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a GroupApiDoGroupGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **partition** | **optional.String**| AND the members set by the given partition name (p#.#).  NULL will return the group members not in ANY partition. | 

### Return type

[**Group100**](Group.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupLabelsGet**
> []string DoGroupLabelsGet(ctx, )
Retrieve all existing group labels

Retrieve a string array of all group labels (i.e. group names) that currently exist in HSM.

### Required Parameters
This endpoint does not need any parameter.

### Return type

**[]string**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupMemberDelete**
> Response100 DoGroupMemberDelete(ctx, groupLabel, xnameId)
Delete member from existing group

Delete component {xname_id} from the members of group {group_label}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **groupLabel** | **string**| Specifies an existing group {group_label} to remove the member from. | 
  **xnameId** | **string**| Member of {group_label} to remove. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupMembersGet**
> Members100 DoGroupMembersGet(ctx, groupLabel, optional)
Retrieve all members of existing group

Retrieve members of an existing group {group_label}, optionally filtering the set, returning a members set containing the component xname IDs.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **groupLabel** | **string**| Specifies an existing group {group_label} to query the members of. | 
 **optional** | ***GroupApiDoGroupMembersGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a GroupApiDoGroupMembersGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------

 **partition** | **optional.String**| AND the members set by the given partition name (p#.#).  NULL will return the group members not in ANY partition. | 

### Return type

[**Members100**](Members.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupMembersPost**
> []ResourceUri100 DoGroupMembersPost(ctx, body, groupLabel)
Create new member of existing group (via POST)

Create a new member of group {group_label} with the component xname ID provided in the payload. New member should not already exist in the given group.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**MemberId**](MemberId.md)|  | 
  **groupLabel** | **string**| Specifies an existing group {group_label} to add the new member to. | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupPatch**
> DoGroupPatch(ctx, body, groupLabel)
Update metadata for existing group {group_label}

To update the tags array and/or description, a PATCH operation can be used.  Omitted fields are not updated. This cannot be used to completely replace the members list. Rather, individual members can be removed or added with the POST/DELETE {group_label}/members API below.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Group100Patch**](Group100Patch.md)|  | 
  **groupLabel** | **string**| Label (i.e. name) of the group to update. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupsGet**
> []Group100 DoGroupsGet(ctx, optional)
Retrieve all existing groups

Retrieve all groups that currently exist, optionally filtering the set, returning an array of groups.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***GroupApiDoGroupsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a GroupApiDoGroupsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **group** | **optional.String**| Retrieve the group with the provided group label. Can be repeated to select multiple groups. | 
 **tag** | **optional.String**| Retrieve all groups associated with the given free-form tag from the tags field. | 

### Return type

[**[]Group100**](Group.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGroupsPost**
> []ResourceUri100 DoGroupsPost(ctx, body)
Create a new group

Create a new group identified by the group_label field. Label should be given explicitly, and should not conflict with any existing group, or an error will occur.  Note that if the exclusiveGroup field is present, the group is not allowed to add a member that exists under a different group/label where the exclusiveGroup field is the same. This can be used to create groups of groups where a component may only be present in one of the set.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Group100**](Group100.md)|  | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

