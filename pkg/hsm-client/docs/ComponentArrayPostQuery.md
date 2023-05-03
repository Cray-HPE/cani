# ComponentArrayPostQuery

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ComponentIDs** | **[]string** | An array of XName/ID values for the components to query. | [optional] [default to null]
**Partition** | **string** | Partition name to filter on, as per current /partitions/names | [optional] [default to null]
**Group** | **string** | Group label to filter on, as per current /groups/labels | [optional] [default to null]
**Stateonly** | **bool** | Return only component state and flag fields (plus xname/ID and type).  Results can be modified and used for bulk state/flag- only patch operations. | [optional] [default to null]
**Flagonly** | **bool** | Return only component flag field (plus xname/ID and type). Results can be modified and used for bulk flag-only patch operations. | [optional] [default to null]
**Roleonly** | **bool** | Return only component role and subrole fields (plus xname/ID and type). Results can be modified and used for bulk role-only patches. | [optional] [default to null]
**Nidonly** | **bool** | Return only component NID field (plus xname/ID and type). Results can be modified and used for bulk NID-only patches. | [optional] [default to null]
**Type_** | **[]string** | Retrieve all components with the given HMS type. | [optional] [default to null]
**State** | **[]string** | Retrieve all components with the given HMS state. | [optional] [default to null]
**Flag** | **[]string** | Retrieve all components with the given HMS flag value. | [optional] [default to null]
**Enabled** | **[]string** | Retrieve all components with the given enabled status (true or false). | [optional] [default to null]
**Softwarestatus** | **[]string** | Retrieve all components with the given software status. Software status is a free form string. Matching is case-insensitive. | [optional] [default to null]
**Role** | **[]string** | Retrieve all components (i.e. nodes) with the given HMS role | [optional] [default to null]
**Subrole** | **[]string** | Retrieve all components (i.e. nodes) with the given HMS subrole | [optional] [default to null]
**Subtype** | **[]string** | Retrieve all components with the given HMS subtype. | [optional] [default to null]
**Arch** | **[]string** | Retrieve all components with the given architecture. | [optional] [default to null]
**Class** | **[]string** | Retrieve all components (i.e. nodes) with the given HMS hardware class. Class can be River, Mountain, etc. | [optional] [default to null]
**Nid** | **[]string** | Retrieve all components (i.e. one node) with the given integer NID | [optional] [default to null]
**NidStart** | **[]string** | Retrieve all components (i.e. nodes) with NIDs equal to or greater than the provided integer. | [optional] [default to null]
**NidEnd** | **[]string** | Retrieve all components (i.e. nodes) with NIDs less than or equal to the provided integer. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

