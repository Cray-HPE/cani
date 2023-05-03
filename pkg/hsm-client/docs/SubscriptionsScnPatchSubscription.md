# SubscriptionsScnPatchSubscription

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Op** | **string** | The type of operation to be performed on the subscription | [optional] [default to null]
**Enabled** | **bool** | This value toggles subscriptions to state change notifications concerning components being disabled or enabled. &#x27;true&#x27; will cause the subscriber to be notified about components being enabled or disabled. &#x27;false&#x27; or empty will result in no such notifications. | [optional] [default to null]
**Roles** | **[]string** | This is an array containing component roles for which to be notified when role changes occur. | [optional] [default to null]
**SubRoles** | **[]string** | This is an array containing component subroles for which to be notified when subrole changes occur. | [optional] [default to null]
**SoftwareStatus** | **[]string** | This is an array containing component software statuses for which to be notified when software status changes occur. | [optional] [default to null]
**States** | [**[]HmsState100**](HMSState.1.0.0.md) | This is an array containing component states for which to be notified when state changes occur. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

