# HwInventory100RedfishManagerLocationInfo

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DateTime** | **string** | This is a pass-through of the Redfish value of the same name. The current date and time with UTC offset that the manager uses to set or read time. | [optional] [default to null]
**DateTimeLocalOffset** | **string** | This is a pass-through of the Redfish value of the same name. The time offset from UTC that the DateTime property is in +HH:MM format. | [optional] [default to null]
**Description** | **string** | This is a pass-through of the Redfish value of the same name. This is an informational description set by the implementation. | [optional] [default to null]
**FirmwareVersion** | **string** | This is a pass-through of the Redfish value of the same name. | [optional] [default to null]
**Id** | **string** | This is a pass-through of the Redfish value of the same name. The Id is included for informational purposes.  The RedfishEndpoint objects are intended to help locate and interact with HMS components via the Redfish endpoint, so this is mostly needed in case servicing the component requires its ID/name according to a particular COTS manufacturers naming scheme within, for example, a particular server enclosure. | [optional] [default to null]
**Name** | **string** | This is a pass-through of the Redfish value of the same name. This is included for informational purposes as the naming will likely vary from manufacturer-to-manufacturer, but should help match items up to manufacturer&#x27;s documentation if the normalized HMS naming scheme is too vague for some COTS systems. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

