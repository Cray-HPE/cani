/*
 * HPCM cmdb REST API Documentation
 *
 * HPE Performance Cluster Manager 'cmdb' service features a REST API. This section describes its implementation.  Standard REST API concepts (such as HTTP verbs, return codes, JSON, etc.) are not covered here.
 *
 * API version: v1
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package hpcm_client

import (
	"time"
)

type Architecture struct {
	Name             string                 `json:"name,omitempty"`
	Id               int64                  `json:"id,omitempty"`
	Uuid             string                 `json:"uuid,omitempty"`
	Etag             string                 `json:"etag,omitempty"`
	CreationTime     time.Time              `json:"creationTime,omitempty"`
	ModificationTime time.Time              `json:"modificationTime,omitempty"`
	DeletionTime     time.Time              `json:"deletionTime,omitempty"`
	Links            map[string]string      `json:"links,omitempty"`
	Platforms        []Platform             `json:"platforms,omitempty"`
	Attributes       map[string]interface{} `json:"attributes,omitempty"`
}
