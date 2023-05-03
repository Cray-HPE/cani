# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LocksDisablePost**](AdminLocksApi.md#LocksDisablePost) | **Post** /locks/disable | Disables the ability to create a reservation on components.
[**LocksLockPost**](AdminLocksApi.md#LocksLockPost) | **Post** /locks/lock | Locks components.
[**LocksRepairPost**](AdminLocksApi.md#LocksRepairPost) | **Post** /locks/repair | Repair components lock and reservation ability.
[**LocksStatusGet**](AdminLocksApi.md#LocksStatusGet) | **Get** /locks/status | Retrieve lock status for all components or a filtered subset of components.
[**LocksStatusPost**](AdminLocksApi.md#LocksStatusPost) | **Post** /locks/status | Retrieve lock status for component IDs.
[**LocksUnlockPost**](AdminLocksApi.md#LocksUnlockPost) | **Post** /locks/unlock | Unlocks components.

# **LocksDisablePost**
> XnameResponse100 LocksDisablePost(ctx, body)
Disables the ability to create a reservation on components.

Disables the ability to create a reservation on components, deletes any existing reservations. Does not change lock state. Attempting to disable an already-disabled component will not result in an error.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminLock100**](AdminLock100.md)| List of xnames to disable. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksLockPost**
> XnameResponse100 LocksLockPost(ctx, body)
Locks components.

Using a component create a lock.  Cannot be locked if already locked, or if there is a current reservation.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminLock100**](AdminLock100.md)| List of xnames to lock. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksRepairPost**
> XnameResponse100 LocksRepairPost(ctx, body)
Repair components lock and reservation ability.

Repairs the disabled status of an xname allowing new reservations to be created.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminLock100**](AdminLock100.md)| List of xnames to repair. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksStatusGet**
> AdminStatusCheckResponse100 LocksStatusGet(ctx, optional)
Retrieve lock status for all components or a filtered subset of components.

Retrieve the status of all component locks and/or reservations. Results can be filtered by query parameters.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
 **optional** | ***AdminLocksApiLocksStatusGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a AdminLocksApiLocksStatusGetOpts struct
Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **type_** | **optional.String**| Filter the results based on HMS type like Node, NodeEnclosure, NodeBMC etc. Can be specified multiple times for selecting entries of multiple types. | 
 **state** | **optional.String**| Filter the results based on HMS state like Ready, On etc. Can be specified multiple times for selecting entries in different states. | 
 **role** | **optional.String**| Filter the results based on HMS role. Can be specified multiple times for selecting entries with different roles. Valid values are: - Compute - Service - System - Application - Storage - Management Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/role&#x27; for the complete list. | 
 **subrole** | **optional.String**| Filter the results based on HMS subrole. Can be specified multiple times for selecting entries with different subroles. Valid values are: - Master - Worker - Storage Additional valid values may be added via configuration file. See the results of &#x27;GET /service/values/subrole&#x27; for the complete list. | 
 **locked** | **optional.Bool**| Return components based on the &#x27;Locked&#x27; field of their lock status. | 
 **reserved** | **optional.Bool**| Return components based on the &#x27;Reserved&#x27; field of their lock status. | 
 **reservationDisabled** | **optional.Bool**| Return components based on the &#x27;ReservationDisabled&#x27; field of their lock status. | 

### Return type

[**AdminStatusCheckResponse100**](AdminStatusCheck_Response.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksStatusPost**
> AdminStatusCheckResponse100 LocksStatusPost(ctx, body)
Retrieve lock status for component IDs.

Using component ID retrieve the status of any lock and/or reservation.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**Xnames**](Xnames.md)| List of components to retrieve status. | 

### Return type

[**AdminStatusCheckResponse100**](AdminStatusCheck_Response.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksUnlockPost**
> XnameResponse100 LocksUnlockPost(ctx, body)
Unlocks components.

Using a component unlock a lock.  Cannot be unlocked if already unlocked.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminLock100**](AdminLock100.md)| List of xnames to unlock. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

