# HwInventory100RedfishProcessorLocationInfo

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | This is a pass-through of the Redfish value of the same name. The Id is included for informational purposes.  The RedfishEndpoint objects are intended to help locate and interact with HMS components via the Redfish endpoint, so this is mostly needed in case servicing the component requires its ID/name according to a particular COTS manufacturer&#x27;s naming scheme within, for example, a particular server enclosure. | [optional] [default to null]
**Name** | **string** | This is a pass-through of the Redfish value of the same name. This is included for informational purposes as the naming will likely vary from manufacturer-to-manufacturer, but should help match items up to manufacturer&#x27;s documentation if the normalized HMS naming scheme is too vague for some COTS systems. | [optional] [default to null]
**Description** | **string** | This is a pass-through of the Redfish value of the same name. This is an informational description set by the BMC implementation. | [optional] [default to null]
**Socket** | **string** | This is a pass-through of the Redfish value of the same name. It represents the socket or location of the processor, and may differ from the normalized HMS Ordinal value (or xname) that is always indexed from 0.  Manufacturers may or may not use zero indexing (or may have some other naming scheme for sockets) and so we retain this information to resolve any ambiguity when servicing the component. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

