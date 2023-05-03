# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LocksDisablePost**](LockingApi.md#LocksDisablePost) | **Post** /locks/disable | Disables the ability to create a reservation on components.
[**LocksLockPost**](LockingApi.md#LocksLockPost) | **Post** /locks/lock | Locks components.
[**LocksRepairPost**](LockingApi.md#LocksRepairPost) | **Post** /locks/repair | Repair components lock and reservation ability.
[**LocksReservationsPost**](LockingApi.md#LocksReservationsPost) | **Post** /locks/reservations | Create reservations
[**LocksReservationsReleasePost**](LockingApi.md#LocksReservationsReleasePost) | **Post** /locks/reservations/release | Releases existing reservations.
[**LocksReservationsRemovePost**](LockingApi.md#LocksReservationsRemovePost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.
[**LocksServiceReservationsCheckPost**](LockingApi.md#LocksServiceReservationsCheckPost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
[**LocksServiceReservationsPost**](LockingApi.md#LocksServiceReservationsPost) | **Post** /locks/service/reservations | Create reservations
[**LocksServiceReservationsReleasePost**](LockingApi.md#LocksServiceReservationsReleasePost) | **Post** /locks/service/reservations/release | Releases existing reservations.
[**LocksServiceReservationsRenewPost**](LockingApi.md#LocksServiceReservationsRenewPost) | **Post** /locks/service/reservations/renew | Renew existing reservations.
[**LocksStatusGet**](LockingApi.md#LocksStatusGet) | **Get** /locks/status | Retrieve lock status for all components or a filtered subset of components.
[**LocksStatusPost**](LockingApi.md#LocksStatusPost) | **Post** /locks/status | Retrieve lock status for component IDs.
[**LocksUnlockPost**](LockingApi.md#LocksUnlockPost) | **Post** /locks/unlock | Unlocks components.

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

# **LocksReservationsPost**
> AdminReservationCreateResponse100 LocksReservationsPost(ctx, body)
Create reservations

Creates reservations on a set of xnames of infinite duration.  Component must be locked to create a reservation.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminReservationCreate100**](AdminReservationCreate100.md)| List of components to create reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having reservations created if an xname doesn&#x27;t exist, or isn&#x27;t locked, or if already reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

### Return type

[**AdminReservationCreateResponse100**](AdminReservationCreate_Response.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksReservationsReleasePost**
> XnameResponse100 LocksReservationsReleasePost(ctx, body)
Releases existing reservations.

Given a list of {xname & reservation key}, releases the associated reservations.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ReservedKeys100**](ReservedKeys100.md)| List of {xname and reservation key} to release reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having their reservation released if an xname doesn&#x27;t exist, or isn&#x27;t reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksReservationsRemovePost**
> XnameResponse100 LocksReservationsRemovePost(ctx, body)
Forcibly deletes existing reservations.

Given a list of components, forcibly deletes any existing reservation. Does not change lock state; does not disable the reservation ability of the component. An empty set of xnames will delete reservations on all xnames. This functionality should be used sparingly, the normal flow should be to release reservations, versus removing them.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**AdminReservationRemove100**](AdminReservationRemove100.md)| List of xnames to remove reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having their reservation removed if an xname doesn&#x27;t exist, or isn&#x27;t reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksServiceReservationsCheckPost**
> ServiceReservationCheckResponse100 LocksServiceReservationsCheckPost(ctx, body)
Check the validity of reservations.

Using xname + reservation key check on the validity of reservations.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**DeputyKeys100**](DeputyKeys100.md)| List of components &amp; deputy keys to check on validity of reservations. | 

### Return type

[**ServiceReservationCheckResponse100**](ServiceReservationCheck_Response.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksServiceReservationsPost**
> ServiceReservationCreateResponse100 LocksServiceReservationsPost(ctx, body)
Create reservations

Creates reservations on a set of xnames of finite duration.  Component must be unlocked to create a reservation.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ServiceReservationCreate100**](ServiceReservationCreate100.md)| List of components to create reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having reservations created if an xname doesn&#x27;t exist, or isn&#x27;t locked, or if already reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

### Return type

[**ServiceReservationCreateResponse100**](ServiceReservationCreate_Response.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksServiceReservationsReleasePost**
> XnameResponse100 LocksServiceReservationsReleasePost(ctx, body)
Releases existing reservations.

Given a list of {xname & reservation key}, releases the associated reservations.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ReservedKeys100**](ReservedKeys100.md)| List of {xname and reservation key} to release reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having their reservation released if an xname doesn&#x27;t exist, or isn&#x27;t reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

### Return type

[**XnameResponse100**](XnameResponse_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **LocksServiceReservationsRenewPost**
> XnameResponse100 LocksServiceReservationsRenewPost(ctx, body)
Renew existing reservations.

Given a list of {xname & reservation key}, renews the associated reservations.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**ReservedKeysWithRenewal100**](ReservedKeysWithRenewal100.md)| List of {xname and reservation key} to renew reservations. A &#x60;rigid&#x60; processing model will result in the entire set of xnames not having their reservation renewed if an xname doesn&#x27;t exist, or isn&#x27;t reserved. A &#x60;flexible&#x60; processing model will perform all actions possible. | 

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
 **optional** | ***LockingApiLocksStatusGetOpts** | optional parameters | nil if no parameters

### Optional Parameters
Optional parameters are passed through a pointer to a LockingApiLocksStatusGetOpts struct
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

