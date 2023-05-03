# AdminReservationRemove100

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ComponentIDs** | **[]string** | An array of XName/ID values for the components to query. | [optional] [default to null]
**Partition** | **[]string** | Partition name to filter on, as per current /partitions/names | [optional] [default to null]
**Group** | **[]string** | Group label to filter on, as per current /groups/labels | [optional] [default to null]
**Type_** | **[]string** | Retrieve all components with the given HMS type. | [optional] [default to null]
**State** | [**[]HmsState100**](HMSState.1.0.0.md) | Retrieve all components with the given HMS state. | [optional] [default to null]
**Flag** | [**[]HmsFlag100**](HMSFlag.1.0.0.md) | Retrieve all components with the given HMS flag value. | [optional] [default to null]
**Enabled** | **[]string** | Retrieve all components with the given enabled status (true or false). | [optional] [default to null]
**Softwarestatus** | **[]string** | Retrieve all components with the given software status. Software status is a free form string. Matching is case-insensitive. | [optional] [default to null]
**Role** | **[]string** | Retrieve all components (i.e. nodes) with the given HMS role | [optional] [default to null]
**Subrole** | **[]string** | Retrieve all components (i.e. nodes) with the given HMS subrole | [optional] [default to null]
**Subtype** | **[]string** | Retrieve all components with the given HMS subtype. | [optional] [default to null]
**Arch** | [**[]HmsArch100**](HMSArch.1.0.0.md) | Retrieve all components with the given architecture. | [optional] [default to null]
**Class** | [**[]HmsClass100**](HMSClass.1.0.0.md) | Retrieve all components (i.e. nodes) with the given HMS hardware class. Class can be River, Mountain, etc. | [optional] [default to null]
**NID** | **[]string** | Retrieve all components (i.e. one node) with the given integer NID | [optional] [default to null]
**ProcessingModel** | **string** | Rigid is all or nothing, flexible is best attempt. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

