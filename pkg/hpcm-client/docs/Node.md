# Node

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** |  | [optional] [default to null]
**Aliases** | **map[string]string** |  | [optional] [default to null]
**Id** | **int64** |  | [optional] [default to null]
**Uuid** | **string** |  | [optional] [default to null]
**Etag** | **string** |  | [optional] [default to null]
**CreationTime** | [**time.Time**](time.Time.md) |  | [optional] [default to null]
**ModificationTime** | [**time.Time**](time.Time.md) |  | [optional] [default to null]
**DeletionTime** | [**time.Time**](time.Time.md) |  | [optional] [default to null]
**Links** | **map[string]string** |  | [optional] [default to null]
**Network** | [***NetworkSettings**](NetworkSettings.md) |  | [optional] [default to null]
**Image** | [***ImageSettings**](ImageSettings.md) |  | [optional] [default to null]
**Platform** | [***PlatformSettings**](PlatformSettings.md) |  | [optional] [default to null]
**Management** | [***ManagementSettings**](ManagementSettings.md) |  | [optional] [default to null]
**Controller** | [***ControllerSettings**](ControllerSettings.md) |  | [optional] [default to null]
**Location** | [***LocationSettings**](LocationSettings.md) |  | [optional] [default to null]
**InternalName** | **string** |  | [optional] [default to null]
**Type_** | **string** |  | [optional] [default to null]
**ImageTransport** | **string** |  | [optional] [default to null]
**ImagePending** | **bool** |  | [optional] [default to null]
**TemplateName** | **string** |  | [optional] [default to null]
**RootFs** | **string** |  | [optional] [default to null]
**OperationalStatus** | **int32** |  | [optional] [default to null]
**AdministrativeStatus** | **int32** |  | [optional] [default to null]
**Managed** | **bool** |  | [optional] [default to null]
**Monitoring** | **string** |  | [optional] [default to null]
**RootSlot** | **int32** |  | [optional] [default to null]
**BiosBootMode** | **string** |  | [optional] [default to null]
**BootOrder** | **int32** |  | [optional] [default to null]
**IscsiRoot** | **string** |  | [optional] [default to null]
**Inventory** | [***interface{}**](interface{}.md) |  | [optional] [default to null]
**NodeController** | **string** | Write-only field to configure the controller this node is attached to at creation time | [optional] [default to null]
**Attributes** | [**map[string]interface{}**](interface{}.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

