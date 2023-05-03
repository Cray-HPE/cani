# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoDeleteSCNSubscription**](SCNApi.md#DoDeleteSCNSubscription) | **Delete** /Subscriptions/SCN/{id} | Delete a state change notification subscription
[**DoDeleteSCNSubscriptionsAll**](SCNApi.md#DoDeleteSCNSubscriptionsAll) | **Delete** /Subscriptions/SCN | Delete all state change notification subscriptions
[**DoGetSCNSubscription**](SCNApi.md#DoGetSCNSubscription) | **Get** /Subscriptions/SCN/{id} | Retrieve a currently-held state change notification subscription
[**DoGetSCNSubscriptionsAll**](SCNApi.md#DoGetSCNSubscriptionsAll) | **Get** /Subscriptions/SCN | Retrieve currently-held state change notification subscriptions
[**DoPatchSCNSubscription**](SCNApi.md#DoPatchSCNSubscription) | **Patch** /Subscriptions/SCN/{id} | Update a subscription for state change notifications
[**DoPostSCNSubscription**](SCNApi.md#DoPostSCNSubscription) | **Post** /Subscriptions/SCN | Create a subscription for state change notifications
[**DoPutSCNSubscription**](SCNApi.md#DoPutSCNSubscription) | **Put** /Subscriptions/SCN/{id} | Update a subscription for state change notifications

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

