/*
 * HPCM cmdb REST API Documentation
 *
 * HPE Performance Cluster Manager 'cmdb' service features a REST API. This section describes its implementation.  Standard REST API concepts (such as HTTP verbs, return codes, JSON, etc.) are not covered here.
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package hpcm_client

type ManagementSettings struct {
	CardType       string `json:"cardType,omitempty"`
	CardIpAddress  string `json:"cardIpAddress,omitempty"`
	CardMacAddress string `json:"cardMacAddress,omitempty"`
	Protocol       string `json:"protocol,omitempty"`
	Channel        int32  `json:"channel,omitempty"`
	Username       string `json:"username,omitempty"`
	Password       string `json:"password,omitempty"`
}
