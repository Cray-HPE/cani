# ComponentArrayPostByNidQuery

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NIDRanges** | **[]string** | NID range values to query, producing a ComponentArray with the matching components, e.g. \&quot;0-24\&quot; or \&quot;2\&quot;.  Add each multiple ranges as a separate array item. | [default to null]
**Partition** | **string** |  | [optional] [default to null]
**Stateonly** | **bool** | Return only component state and flag fields (plus xname/ID and type).  Results can be modified and used for bulk state/flag- only patch operations. | [optional] [default to null]
**Flagonly** | **bool** | Return only component flag field (plus xname/ID and type). Results can be modified and used for bulk flag-only patch operations. | [optional] [default to null]
**Roleonly** | **bool** | Return only component role and subrole fields (plus xname/ID and type). Results can be modified and used for bulk role-only patches. | [optional] [default to null]
**Nidonly** | **bool** | Return only component NID field (plus xname/ID and type). Results can be modified and used for bulk NID-only patches. | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

