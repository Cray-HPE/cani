# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoPartitionDelete**](PartitionApi.md#DoPartitionDelete) | **Delete** /partitions/{partition_name} | Delete existing partition with {partition_name}
[**DoPartitionGet**](PartitionApi.md#DoPartitionGet) | **Get** /partitions/{partition_name} | Retrieve existing partition {partition_name}
[**DoPartitionMemberDelete**](PartitionApi.md#DoPartitionMemberDelete) | **Delete** /partitions/{partition_name}/members/{xname_id} | Delete member from existing partition
[**DoPartitionMembersGet**](PartitionApi.md#DoPartitionMembersGet) | **Get** /partitions/{partition_name}/members | Retrieve all members of existing partition
[**DoPartitionMembersPost**](PartitionApi.md#DoPartitionMembersPost) | **Post** /partitions/{partition_name}/members | Create new member of existing partition (via POST)
[**DoPartitionNamesGet**](PartitionApi.md#DoPartitionNamesGet) | **Get** /partitions/names | Retrieve all existing partition names
[**DoPartitionPatch**](PartitionApi.md#DoPartitionPatch) | **Patch** /partitions/{partition_name} | Update metadata for existing partition {partition_name}
[**DoPartitionsGet**](PartitionApi.md#DoPartitionsGet) | **Get** /partitions | Retrieve all existing partitions
[**DoPartitionsPost**](PartitionApi.md#DoPartitionsPost) | **Post** /partitions | Create new partition (via POST)

# **DoPartitionDelete**
> Response100 DoPartitionDelete(ctx, partitionName)
Delete existing partition with {partition_name}

Delete partition {partition_name}. Any members previously in the partition will no longer have the deleted partition name associated with them.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **partitionName** | **string**| Partition name of the partition to delete. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionGet**
> Partition100 DoPartitionGet(ctx, partitionName)
Retrieve existing partition {partition_name}

Retrieve the partition which was created with the given {partition_name}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **partitionName** | **string**| Partition name to be retrieved | 

### Return type

[**Partition100**](Partition.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionMemberDelete**
> Response100 DoPartitionMemberDelete(ctx, partitionName, xnameId)
Delete member from existing partition

Delete component {xname_id} from the members of partition {partition_name}.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **partitionName** | **string**| Existing partition {partition_name} to remove the member from. | 
  **xnameId** | **string**| Member of {partition_name} to remove. | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionMembersGet**
> Members100 DoPartitionMembersGet(ctx, partitionName)
Retrieve all members of existing partition

Retrieve all members of existing partition {partition_name}, optionally filtering the set, returning a members set that includes the component xname IDs.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **partitionName** | **string**| Existing partition {partition_name} to query the members of. | 

### Return type

[**Members100**](Members.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionMembersPost**
> []ResourceUri100 DoPartitionMembersPost(ctx, body, partitionName)
Create new member of existing partition (via POST)

Create a new member of partition {partition_name} with the component xname ID provided in the payload. New member should not already exist in the given partition

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**MemberId**](MemberId.md)|  | 
  **partitionName** | **string**| Existing partition {partition_name} to add the new member to. | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionNamesGet**
> []string DoPartitionNamesGet(ctx, )
Retrieve all existing partition names

Retrieve a string array of all partition names that currently exist in HSM. These are just the names, not the complete partition records.

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

# **DoPartitionPatch**
> DoPartitionPatch(ctx, body, partitionName)
Update metadata for existing partition {partition_name}

Update the tags array and/or description by using PATCH. Omitted fields are not updated. This cannot be used to completely replace the members list. Rather, individual members can be removed or added with the POST/DELETE {partition_name}/members API.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Partition100Patch**](Partition100Patch.md)|  | 
  **partitionName** | **string**| Name of the partition to update. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionsGet**
> []Partition100 DoPartitionsGet(ctx, optional)
Retrieve all existing partitions

Retrieve all partitions that currently exist, optionally filtering the set, returning an array of partition records.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***PartitionApiDoPartitionsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a PartitionApiDoPartitionsGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **partition** | **optional.String**| Retrieve the partition with the provided partition name (p#.#). Can be repeated to select multiple partitions. | 
 **tag** | **optional.String**| Retrieve all partitions associated with the given free-form tag from the tags field. | 

### Return type

[**[]Partition100**](Partition.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPartitionsPost**
> []ResourceUri100 DoPartitionsPost(ctx, body)
Create new partition (via POST)

Create a new partition identified by the partition_name field. Partition names should be of the format p# or p#.# (hard_part.soft_part). Partition name should be given explicitly, and should not conflict with any existing partition, or an error will occur.  In addition, the member list must not overlap with any existing partition.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Partition100**](Partition100.md)|  | 

### Return type

[**[]ResourceUri100**](ResourceURI.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

