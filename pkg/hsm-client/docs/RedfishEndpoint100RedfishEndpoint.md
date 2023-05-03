# RedfishEndpoint100RedfishEndpoint

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ID** | **string** |  | [default to null]
**Type_** | [***HmsType100**](HMSType.1.0.0.md) |  | [optional] [default to null]
**Name** | **string** | This is an arbitrary, user-provided name for the endpoint.  It can describe anything that is not captured by the ID/xname. | [optional] [default to null]
**Hostname** | **string** | Hostname of the endpoint&#x27;s FQDN, will always be the host portion of the fully-qualified domain name. Note that the hostname should normally always be the same as the ID field (i.e. xname) of the endpoint. | [optional] [default to null]
**Domain** | **string** | Domain of the endpoint&#x27;s FQDN.  Will always match remaining non-hostname portion of fully-qualified domain name (FQDN). | [optional] [default to null]
**FQDN** | **string** | Fully-qualified domain name of RF endpoint on management network. This is not writable because it is made up of the Hostname and Domain. | [optional] [default to null]
**Enabled** | **bool** | To disable a component without deleting its data from the database, can be set to false | [optional] [default to null]
**UUID** | **string** |  | [optional] [default to null]
**User** | **string** | Username to use when interrogating endpoint | [optional] [default to null]
**Password** | **string** | Password to use when interrogating endpoint, normally suppressed in output. | [optional] [default to null]
**UseSSDP** | **bool** | Whether to use SSDP for discovery if the EP supports it. | [optional] [default to null]
**MacRequired** | **bool** | Whether the MAC must be used (e.g. in River) in setting up geolocation info so the endpoint&#x27;s location in the system can be determined.  The MAC does not need to be provided when creating the endpoint if the endpoint type can arrive at a geolocated hostname on its own. | [optional] [default to null]
**MACAddr** | **string** | This is the MAC on the of the Redfish Endpoint on the management network, i.e. corresponding to the FQDN field&#x27;s Ethernet interface where the root service is running. Not the HSN MAC. This is a MAC address in the standard colon-separated 12 byte hex format. | [optional] [default to null]
**IPAddress** | **string** | This is the IP of the Redfish Endpoint on the management network, i.e. corresponding to the FQDN field&#x27;s Ethernet interface where the root service is running. This may be IPv4 or IPv6 | [optional] [default to null]
**RediscoverOnUpdate** | **bool** | Trigger a rediscovery when endpoint info is updated. | [optional] [default to null]
**TemplateID** | **string** | Links to a discovery template defining how the endpoint should be discovered. | [optional] [default to null]
**DiscoveryInfo** | [***RedfishEndpoint100RedfishEndpointDiscoveryInfo**](RedfishEndpoint.1.0.0_RedfishEndpoint_DiscoveryInfo.md) |  | [optional] [default to null]

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)

