# HardwareExtraPropertiesMgmtSwitchConnector

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**CaniId** | **string** |  | [optional] [default to null]
**CaniLastModified** | **string** |  | [optional] [default to null]
**CaniSlsSchemaVersion** | **string** |  | [optional] [default to null]
**CaniStatus** | [***CaniStatus**](CANIStatus.md) |  | [optional] [default to null]
**NodeNics** | **[]string** | An array of xnames that the hardware_mgmt_switch_connector is connected to.  Excludes the parent. | [default to null]
**VendorName** | **string** | The vendor-assigned name for this port, as it appears in the switch management software.  Typically this is something like \&quot;GigabitEthernet 1/31\&quot; (Berkeley-style names), but may be any string. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

