# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoMembershipGet**](MembershipApi.md#DoMembershipGet) | **Get** /memberships/{xname} | Retrieve membership for component {xname}
[**DoMembershipsGet**](MembershipApi.md#DoMembershipsGet) | **Get** /memberships | Retrieve all memberships for components

# **DoMembershipGet**
> Membership100 DoMembershipGet(ctx, xname)
Retrieve membership for component {xname}

Display group labels and partition names for a given component xname ID.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **xname** | **string**| Component xname ID (i.e. locational identifier) | 

### Return type

[**Membership100**](Membership.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoMembershipsGet**
> []Membership100 DoMembershipsGet(ctx, optional)
Retrieve all memberships for components

Display group labels and partition names for each component xname ID (where applicable).

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***MembershipApiDoMembershipsGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a MembershipApiDoMembershipsGetOpts struct
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

### Return type

[**[]Membership100**](Membership.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

