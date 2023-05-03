# Lock100

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Id** | **string** | The ID number of the lock. | [optional] [default to null]
**Created** | [**time.Time**](time.Time.md) | A timestamp for when the lock was created. | [optional] [default to null]
**Reason** | **string** | A one-line, user-provided reason for the lock. | [optional] [default to null]
**Owner** | **string** | A user-provided self identifier for the lock | [default to null]
**Lifetime** | **int32** | The length of time in seconds the component lock should exist before it is automatically deleted by HSM. | [default to null]
**Xnames** | **[]string** | An array of XName/ID values for the components managed by the lock. These components will have their component flag set to \&quot;Locked\&quot; upon lock creation and set to \&quot;OK\&quot; upon lock deletion. | [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

