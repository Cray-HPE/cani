# {{classname}}

All URIs are relative to *https://sms/apis/smd/hsm/v2*

Method | HTTP request | Description
------------- | ------------- | -------------
[**DoArchValuesGet**](ServiceInfoApi.md#DoArchValuesGet) | **Get** /service/values/arch | Retrieve all valid values for use with the &#x27;arch&#x27; parameter
[**DoClassValuesGet**](ServiceInfoApi.md#DoClassValuesGet) | **Get** /service/values/class | Retrieve all valid values for use with the &#x27;class&#x27; parameter
[**DoFlagValuesGet**](ServiceInfoApi.md#DoFlagValuesGet) | **Get** /service/values/flag | Retrieve all valid values for use with the &#x27;flag&#x27; parameter
[**DoLivenessGet**](ServiceInfoApi.md#DoLivenessGet) | **Get** /service/liveness | Kubernetes liveness endpoint to monitor service health
[**DoNetTypeValuesGet**](ServiceInfoApi.md#DoNetTypeValuesGet) | **Get** /service/values/nettype | Retrieve all valid values for use with the &#x27;nettype&#x27; parameter
[**DoReadyGet**](ServiceInfoApi.md#DoReadyGet) | **Get** /service/ready | Kubernetes readiness endpoint to monitor service health
[**DoRoleValuesGet**](ServiceInfoApi.md#DoRoleValuesGet) | **Get** /service/values/role | Retrieve all valid values for use with the &#x27;role&#x27; parameter
[**DoStateValuesGet**](ServiceInfoApi.md#DoStateValuesGet) | **Get** /service/values/state | Retrieve all valid values for use with the &#x27;state&#x27; parameter
[**DoSubRoleValuesGet**](ServiceInfoApi.md#DoSubRoleValuesGet) | **Get** /service/values/subrole | Retrieve all valid values for use with the &#x27;subrole&#x27; parameter
[**DoTypeValuesGet**](ServiceInfoApi.md#DoTypeValuesGet) | **Get** /service/values/type | Retrieve all valid values for use with the &#x27;type&#x27; parameter
[**DoValuesGet**](ServiceInfoApi.md#DoValuesGet) | **Get** /service/values | Retrieve all valid values for use as parameters

# **DoArchValuesGet**
> Values100ArchArray DoArchValuesGet(ctx, )
Retrieve all valid values for use with the 'arch' parameter

Retrieve all valid values for use with the 'arch' (component architecture) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100ArchArray**](Values.1.0.0_ArchArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoClassValuesGet**
> Values100ClassArray DoClassValuesGet(ctx, )
Retrieve all valid values for use with the 'class' parameter

Retrieve all valid values for use with the 'class' (hardware class) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100ClassArray**](Values.1.0.0_ClassArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoFlagValuesGet**
> Values100FlagArray DoFlagValuesGet(ctx, )
Retrieve all valid values for use with the 'flag' parameter

Retrieve all valid values for use with the 'flag' (component flag) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100FlagArray**](Values.1.0.0_FlagArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoLivenessGet**
> DoLivenessGet(ctx, )
Kubernetes liveness endpoint to monitor service health

The `liveness` resource works in conjunction with the Kubernetes liveness probe to determine when the service is no longer responding to requests.  Too many failures of the liveness probe will result in the service being shut down and restarted.  This is primarily an endpoint for the automated Kubernetes system.

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

# **DoNetTypeValuesGet**
> Values100NetTypeArray DoNetTypeValuesGet(ctx, )
Retrieve all valid values for use with the 'nettype' parameter

Retrieve all valid values for use with the 'nettype' (component network type) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100NetTypeArray**](Values.1.0.0_NetTypeArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoReadyGet**
> Response100 DoReadyGet(ctx, )
Kubernetes readiness endpoint to monitor service health

The `readiness` resource works in conjunction with the Kubernetes readiness probe to determine when the service is no longer healthy and able to respond correctly to requests.  Too many failures of the readiness probe will result in the traffic being routed away from this service and eventually the service will be shut down and restarted if in an unready state for too long.  This is primarily an endpoint for the automated Kubernetes system.

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

# **DoRoleValuesGet**
> Values100RoleArray DoRoleValuesGet(ctx, )
Retrieve all valid values for use with the 'role' parameter

Retrieve all valid values for use with the 'role' (component role) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100RoleArray**](Values.1.0.0_RoleArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoStateValuesGet**
> Values100StateArray DoStateValuesGet(ctx, )
Retrieve all valid values for use with the 'state' parameter

Retrieve all valid values for use with the 'state' (component state) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100StateArray**](Values.1.0.0_StateArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoSubRoleValuesGet**
> Values100SubRoleArray DoSubRoleValuesGet(ctx, )
Retrieve all valid values for use with the 'subrole' parameter

Retrieve all valid values for use with the 'subrole' (component subrole) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100SubRoleArray**](Values.1.0.0_SubRoleArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoTypeValuesGet**
> Values100TypeArray DoTypeValuesGet(ctx, )
Retrieve all valid values for use with the 'type' parameter

Retrieve all valid values for use with the 'type' (component HMSType) parameter.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100TypeArray**](Values.1.0.0_TypeArray.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **DoValuesGet**
> Values100Values DoValuesGet(ctx, )
Retrieve all valid values for use as parameters

Retrieve all valid values for use as parameters.

### Required Parameters
This endpoint does not need any parameter.

### Return type

[**Values100Values**](Values.1.0.0_Values.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json, application/problem+json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

