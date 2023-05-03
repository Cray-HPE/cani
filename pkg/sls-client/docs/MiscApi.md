# {{classname}}

All URIs are relative to *https://api-gw-service-nmn.local/apis/sls/v1*

Method | HTTP request | Description
------------- | ------------- | -------------
[**HealthGet**](MiscApi.md#HealthGet) | **Get** /health | Query the health of the service
[**LivenessGet**](MiscApi.md#LivenessGet) | **Get** /liveness | Kubernetes liveness endpoint to monitor service health
[**ReadinessGet**](MiscApi.md#ReadinessGet) | **Get** /readiness | Kubernetes readiness endpoint to monitor service health
[**VersionGet**](MiscApi.md#VersionGet) | **Get** /version | Retrieve versioning information on the information in SLS

# **HealthGet**
> InlineResponse200 HealthGet(ctx, )
Query the health of the service

The `health` resource returns health information about the SLS service and its dependencies.  This actively checks the connection between  SLS and the following:   * Vault   * Database   This is primarily intended as a diagnostic tool to investigate the functioning of the SLS service.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**InlineResponse200**](inline_response_200.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

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

# **VersionGet**
> VersionResponse VersionGet(ctx, )
Retrieve versioning information on the information in SLS

Retrieve the current version of the SLS mapping. Information returned is a JSON array with two keys: * Counter: A monotonically increasing counter. This counter is incremented every time   a change is made to the map stored in SLS. This shall be 0 if no data is uploaded to SLS * LastUpdated: An ISO 8601 datetime representing the time of the last change to SLS.    This shall be set to the Unix Epoch if no data has ever been stored in SLS.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**VersionResponse**](versionResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

