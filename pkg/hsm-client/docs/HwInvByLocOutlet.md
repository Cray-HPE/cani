# HwInvByLocOutlet

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ID** | **string** |  | [default to null]
**Type_** | [***HmsType100**](HMSType.1.0.0.md) |  | [optional] [default to null]
**Ordinal** | **int32** | This is the normalized (from zero) index of the component location (e.g. slot number) when there are more than one.  This should match the last number in the xname in most cases (e.g. Ordinal 0 for node x0c0s0b0n0).  Note that Redfish may use a different value or naming scheme, but this is passed through via the *LocationInfo for the type of component. | [optional] [default to null]
**Status** | **string** | Populated or Empty - whether location is populated. | [optional] [default to null]
**HWInventoryByLocationType** | **string** | This is used as a discriminator to determine the additional HMS-type specific subtype that is returned. | [default to null]
**PopulatedFRU** | [***HwInventory100HwInventoryByFru**](HWInventory.1.0.0_HWInventoryByFRU.md) |  | [optional] [default to null]
**OutletLocationInfo** | [***HwInventory100RedfishOutletLocationInfo**](HWInventory.1.0.0_RedfishOutletLocationInfo.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

