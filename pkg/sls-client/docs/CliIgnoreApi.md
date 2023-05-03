# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**LivenessGet**](CliIgnoreApi.md#LivenessGet) | **Get** /liveness | Kubernetes liveness endpoint to monitor service health
[**ReadinessGet**](CliIgnoreApi.md#ReadinessGet) | **Get** /readiness | Kubernetes readiness endpoint to monitor service health

# **LivenessGet**
> LivenessGet(ctx, )
Kubernetes liveness endpoint to monitor service health

The `liveness` resource works in conjunction with the Kubernetes liveness probe to determine when the service is no longer responding to requests.  Too many failures of the liveness probe will result in the service being shut down and restarted.    This is primarily an endpoint for the automated Kubernetes system.

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

# **ReadinessGet**
> ReadinessGet(ctx, )
Kubernetes readiness endpoint to monitor service health

The `readiness` resource works in conjunction with the Kubernetes readiness probe to determine when the service is no longer healthy and able to respond correctly to requests.  Too many failures of the readiness probe will result in the traffic being routed away from this service and eventually the service will be shut down and restarted if in an unready state for too long.  This is primarily an endpoint for the automated Kubernetes system.

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

