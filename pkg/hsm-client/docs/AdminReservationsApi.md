# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LocksReservationsPost**](AdminReservationsApi.md#LocksReservationsPost) | **Post** /locks/reservations | Create reservations
[**LocksReservationsReleasePost**](AdminReservationsApi.md#LocksReservationsReleasePost) | **Post** /locks/reservations/release | Releases existing reservations.
[**LocksReservationsRemovePost**](AdminReservationsApi.md#LocksReservationsRemovePost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.

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

