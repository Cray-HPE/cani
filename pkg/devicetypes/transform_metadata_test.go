package devicetypes

import (
	"reflect"
	"strings"
	"testing"
)

// TestMergeTransformMetadataPreservesCustomFields verifies complete definitions
// survive a transform and repeated imports keep existing definitions by key.
//
// Why it matters: export needs the field schema before sending custom values.
// Inputs: a local definition, a conflicting reimport, and a new select field.
// Outputs: the local definition and every field of the new definition survive.
// Data choice: choices and a typed default expose partial or lossy copying.
func TestMergeTransformMetadataPreservesCustomFields(t *testing.T) {
	inventory := NewInventory()
	local := CustomFieldDefinition{Key: "owner", Label: "Local owner", Type: "text", ContentTypes: []string{"dcim.device"}}
	inventory.Metadata.CustomFields = []CustomFieldDefinition{local}
	definition := CustomFieldDefinition{
		Key: "tier", Label: "Service tier", Type: "select",
		ContentTypes: []string{"dcim.device", "dcim.module"},
		Description:  "Service classification", Required: true,
		Default: "standard", Choices: []string{"standard", "critical"}, Weight: 42,
	}
	result := &TransformResult{Metadata: &InventoryMetadata{CustomFields: []CustomFieldDefinition{
		{Key: "owner", Label: "Remote owner", Type: "text", ContentTypes: []string{"dcim.rack"}},
		definition,
	}}}

	for range 2 {
		if _, err := inventory.MergeTransformResult(result); err != nil {
			t.Fatalf("MergeTransformResult() error = %v", err)
		}
	}

	want := []CustomFieldDefinition{local, definition}
	if !reflect.DeepEqual(inventory.Metadata.CustomFields, want) {
		t.Fatalf("custom fields = %#v, want %#v", inventory.Metadata.CustomFields, want)
	}
	result.Metadata.CustomFields[1].Choices[0] = "changed"
	if inventory.Metadata.CustomFields[1].Choices[0] != "standard" {
		t.Error("merged choices alias the transform input")
	}
}

// TestMergeTransformMetadataRejectsInvalidCustomField verifies invalid new
// definitions fail the complete-result transaction instead of disappearing.
//
// Why it matters: a reported successful import must not silently lose schema.
// Inputs: a new role followed by a custom field with an invalid content type.
// Outputs: a field-specific error and an unchanged inventory metadata pointer.
// Data choice: a preceding valid role detects partial application on failure.
func TestMergeTransformMetadataRejectsInvalidCustomField(t *testing.T) {
	inventory := NewInventory()
	originalMetadata := inventory.Metadata
	result := &TransformResult{Metadata: &InventoryMetadata{
		Roles: []MetadataEntry{{Name: "compute"}},
		CustomFields: []CustomFieldDefinition{{
			Key: "owner", Type: "text", ContentTypes: []string{"missing.object"},
		}},
	}}

	_, err := inventory.MergeTransformResult(result)

	if err == nil || !strings.Contains(err.Error(), "owner") {
		t.Fatalf("MergeTransformResult() error = %v, want invalid owner definition", err)
	}
	if inventory.Metadata != originalMetadata || len(inventory.Metadata.Roles) != 0 {
		t.Error("failed metadata merge changed the live inventory")
	}
}
