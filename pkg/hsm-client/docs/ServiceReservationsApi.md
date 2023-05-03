# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LocksServiceReservationsCheckPost**](ServiceReservationsApi.md#LocksServiceReservationsCheckPost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
[**LocksServiceReservationsPost**](ServiceReservationsApi.md#LocksServiceReservationsPost) | **Post** /locks/service/reservations | Create reservations
[**LocksServiceReservationsReleasePost**](ServiceReservationsApi.md#LocksServiceReservationsReleasePost) | **Post** /locks/service/reservations/release | Releases existing reservations.
[**LocksServiceReservationsRenewPost**](ServiceReservationsApi.md#LocksServiceReservationsRenewPost) | **Post** /locks/service/reservations/renew | Renew existing reservations.

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

