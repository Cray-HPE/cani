# PowerControl100OemCrayPowerLimit

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Min** | **float64** | The minimum allowed value for a PowerLimit&#x27;s LimitInWatts. This is the estimated lowest value (most restrictive) power cap that can be achieved by the associated PowerControl resource. | [optional] [default to null]
**Max** | **float64** | The maximum allowed value for a PowerLimit&#x27;s LimitInWatts. This is the estimated highest value (least restrictive) power cap that can be achieved by the associated PowerControl resource. Note that the actual maximum allowed LimitInWatts is the lesser of PowerLimit.Max or PowerControl.PowerAllocatedWatts. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

