# Component100Component

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ID** | **string** |  | [optional] [default to null]
**Type_** | [***HmsType100**](HMSType.1.0.0.md) |  | [optional] [default to null]
**State** | [***HmsState100**](HMSState.1.0.0.md) |  | [optional] [default to null]
**Flag** | [***HmsFlag100**](HMSFlag.1.0.0.md) |  | [optional] [default to null]
**Enabled** | **bool** | Whether component is enabled. True when enabled, false when disabled. | [optional] [default to null]
**SoftwareStatus** | **string** | SoftwareStatus of a node, used by the managed plane for running nodes.  Will be missing for other component types or if not set by software. | [optional] [default to null]
**Role** | **string** |  | [optional] [default to null]
**SubRole** | **string** |  | [optional] [default to null]
**NID** | **int32** | This is the integer Node ID if the component is a node. | [optional] [default to null]
**Subtype** | **string** | Further distinguishes between components of same type. | [optional] [default to null]
**NetType** | [***NetType100**](NetType.1.0.0.md) |  | [optional] [default to null]
**Arch** | [***HmsArch100**](HMSArch.1.0.0.md) |  | [optional] [default to null]
**Class** | [***HmsClass100**](HMSClass.1.0.0.md) |  | [optional] [default to null]
**ReservationDisabled** | **bool** | Whether component can be reserved via the locking API. True when reservations are disabled, thus no new reservations can be created on this component. | [optional] [default to null]
**Locked** | **bool** | Whether a component is locked via the locking API. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

