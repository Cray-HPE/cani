# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoDeleteSCNSubscription**](CliIgnoreApi.md#DoDeleteSCNSubscription) | **Delete** /Subscriptions/SCN/{id} | Delete a state change notification subscription
[**DoDeleteSCNSubscriptionsAll**](CliIgnoreApi.md#DoDeleteSCNSubscriptionsAll) | **Delete** /Subscriptions/SCN | Delete all state change notification subscriptions
[**DoGetSCNSubscription**](CliIgnoreApi.md#DoGetSCNSubscription) | **Get** /Subscriptions/SCN/{id} | Retrieve a currently-held state change notification subscription
[**DoGetSCNSubscriptionsAll**](CliIgnoreApi.md#DoGetSCNSubscriptionsAll) | **Get** /Subscriptions/SCN | Retrieve currently-held state change notification subscriptions
[**DoHWInvByLocationPost**](CliIgnoreApi.md#DoHWInvByLocationPost) | **Post** /Inventory/Hardware | Create/Update hardware inventory entries
[**DoPatchSCNSubscription**](CliIgnoreApi.md#DoPatchSCNSubscription) | **Patch** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
[**DoPostSCNSubscription**](CliIgnoreApi.md#DoPostSCNSubscription) | **Post** /Subscriptions/SCN | Create a subscription for state change notifications
[**DoPowerMapsDeleteAll**](CliIgnoreApi.md#DoPowerMapsDeleteAll) | **Delete** /sysinfo/powermaps | Delete all PowerMap entities
[**DoPutSCNSubscription**](CliIgnoreApi.md#DoPutSCNSubscription) | **Put** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
[**DoRedfishEndpointPut**](CliIgnoreApi.md#DoRedfishEndpointPut) | **Put** /Inventory/RedfishEndpoints/{xname} | Update definition for RedfishEndpoint ID {xname}
[**LocksReservationsPost**](CliIgnoreApi.md#LocksReservationsPost) | **Post** /locks/reservations | Create reservations
[**LocksReservationsReleasePost**](CliIgnoreApi.md#LocksReservationsReleasePost) | **Post** /locks/reservations/release | Releases existing reservations.
[**LocksReservationsRemovePost**](CliIgnoreApi.md#LocksReservationsRemovePost) | **Post** /locks/reservations/remove | Forcibly deletes existing reservations.
[**LocksServiceReservationsCheckPost**](CliIgnoreApi.md#LocksServiceReservationsCheckPost) | **Post** /locks/service/reservations/check | Check the validity of reservations.
[**LocksServiceReservationsPost**](CliIgnoreApi.md#LocksServiceReservationsPost) | **Post** /locks/service/reservations | Create reservations
[**LocksServiceReservationsReleasePost**](CliIgnoreApi.md#LocksServiceReservationsReleasePost) | **Post** /locks/service/reservations/release | Releases existing reservations.
[**LocksServiceReservationsRenewPost**](CliIgnoreApi.md#LocksServiceReservationsRenewPost) | **Post** /locks/service/reservations/renew | Renew existing reservations.

# **DoDeleteSCNSubscription**
> DoDeleteSCNSubscription(ctx, id)
Delete a state change notification subscription

Delete a state change notification subscription.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| This is the ID associated with the subscription that was generated at its creation. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoDeleteSCNSubscriptionsAll**
> DoDeleteSCNSubscriptionsAll(ctx, )
Delete all state change notification subscriptions

Delete all subscriptions.

### Required Parameters
This endpoint does not need any parameter.

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGetSCNSubscription**
> SubscriptionsScnPostSubscription DoGetSCNSubscription(ctx, id)
Retrieve a currently-held state change notification subscription

Return the information on a currently held state change notification subscription

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **id** | **string**| This is the ID associated with the subscription that was generated at its creation. | 

### Return type

[**SubscriptionsScnPostSubscription**](Subscriptions_SCNPostSubscription.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoGetSCNSubscriptionsAll**
> SubscriptionsScnSubscriptionArray DoGetSCNSubscriptionsAll(ctx, )
Retrieve currently-held state change notification subscriptions

Retrieve all information on currently held state change notification subscriptions.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**SubscriptionsScnSubscriptionArray**](Subscriptions_SCNSubscriptionArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoHWInvByLocationPost**
> Response100 DoHWInvByLocationPost(ctx, body)
Create/Update hardware inventory entries

Create/Update hardware inventory entries

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**InventoryHardwareBody**](InventoryHardwareBody.md)|  | 

### Return type

[**Response100**](Response_1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPatchSCNSubscription**
> DoPatchSCNSubscription(ctx, body, id)
Update a subscription for state change notifications

Update a subscription for state change notifications to add or remove triggers.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SubscriptionsScnPatchSubscription**](SubscriptionsScnPatchSubscription.md)|  | 
  **id** | **string**| This is the ID associated with the subscription that was generated at its creation. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPostSCNSubscription**
> SubscriptionsScnSubscriptionArrayItem100 DoPostSCNSubscription(ctx, body)
Create a subscription for state change notifications

Request a subscription for state change notifications for a set of component states. This will create a new subscription and produce a unique ID for the subscription. This will not affect the existing subscriptions.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SubscriptionsScnPostSubscription**](SubscriptionsScnPostSubscription.md)|  | 

### Return type

[**SubscriptionsScnSubscriptionArrayItem100**](Subscriptions_SCNSubscriptionArrayItem.1.0.0.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoPowerMapsDeleteAll**
> Response100 DoPowerMapsDeleteAll(ctx, )
Delete all PowerMap entities

Delete all entries in the PowerMaps collection.

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

# **DoPutSCNSubscription**
> DoPutSCNSubscription(ctx, body, id)
Update a subscription for state change notifications

Update an existing state change notification subscription in whole. This will overwrite the specified subscription.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**SubscriptionsScnPostSubscription**](SubscriptionsScnPostSubscription.md)|  | 
  **id** | **string**| This is the ID associated with the subscription that was generated at its creation. | 

### Return type

 (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoRedfishEndpointPut**
> RedfishEndpoint100RedfishEndpoint DoRedfishEndpointPut(ctx, body, xname)
Update definition for RedfishEndpoint ID {xname}

Create or update RedfishEndpoint record for a specific xname.

### Required Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **ctx** | **context.Context** | context for authentication, logging, cancellation, deadlines, tracing, etc.
  **body** | [**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint100RedfishEndpoint.md)|  | 
  **xname** | **string**| Locational xname of RedfishEndpoint record to create or update. | 

### Return type

[**RedfishEndpoint100RedfishEndpoint**](RedfishEndpoint.1.0.0_RedfishEndpoint.md)

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

