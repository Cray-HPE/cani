package runtime

import "testing"

// TestJSONMerge verifies the dependency-free JSONMerge reproduces the
// CopyNonexistent merge semantics the generated client previously relied on:
// patch-only keys are added, shared scalar keys are overwritten by the patch,
// nested objects merge recursively, an object value is preserved when the patch
// offers a scalar (type mismatch), a scalar value yields to an object patch, and
// numbers keep their exact source formatting.
//
// Why it matters: every generated union type marshals itself through JSONMerge,
// so a behavioral drift here would silently corrupt Nautobot request/response
// JSON (for example losing a discriminator field or truncating a 64-bit ID).
// Inputs: pairs of raw JSON documents. Output: the merged JSON, whose bytes are
// deterministic because map keys marshal in sorted order.
// Data choice: each case isolates one merge rule, plus a large integer that a
// float64 decode would corrupt and empty/nil inputs that exercise the guards.
func TestJSONMerge(t *testing.T) {
	tests := []struct {
		name  string
		data  string
		patch string
		want  string
	}{
		{
			name:  "adds patch-only keys and overrides shared scalars",
			data:  `{"a":1,"b":2}`,
			patch: `{"b":3,"c":4}`,
			want:  `{"a":1,"b":3,"c":4}`,
		},
		{
			name:  "merges nested objects recursively",
			data:  `{"obj":{"x":1,"y":2}}`,
			patch: `{"obj":{"y":3,"z":4}}`,
			want:  `{"obj":{"x":1,"y":3,"z":4}}`,
		},
		{
			name:  "object value wins over scalar patch (type mismatch)",
			data:  `{"k":{"x":1}}`,
			patch: `{"k":5}`,
			want:  `{"k":{"x":1}}`,
		},
		{
			name:  "scalar value yields to object patch",
			data:  `{"k":1}`,
			patch: `{"k":{"x":2}}`,
			want:  `{"k":{"x":2}}`,
		},
		{
			name:  "preserves large integer formatting",
			data:  `{"id":123456789012345678}`,
			patch: `{}`,
			want:  `{"id":123456789012345678}`,
		},
		{
			name:  "nil data is treated as empty object",
			data:  ``,
			patch: `{"a":1}`,
			want:  `{"a":1}`,
		},
		{
			name:  "nil patch leaves data unchanged",
			data:  `{"a":1}`,
			patch: ``,
			want:  `{"a":1}`,
		},
		{
			name:  "union discriminator plus extra field",
			data:  `{"object_type":"dcim.interface","id":"abc"}`,
			patch: `{"url":"http://x"}`,
			want:  `{"id":"abc","object_type":"dcim.interface","url":"http://x"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JSONMerge([]byte(tt.data), []byte(tt.patch))
			if err != nil {
				t.Fatalf("JSONMerge(%q, %q) returned error: %v", tt.data, tt.patch, err)
			}
			if string(got) != tt.want {
				t.Errorf("JSONMerge(%q, %q) = %q, want %q", tt.data, tt.patch, got, tt.want)
			}
		})
	}
}

// TestJSONMergeInvalidJSON verifies JSONMerge surfaces a decode error instead of
// panicking or silently returning malformed output when either argument is not
// valid JSON.
//
// Why it matters: the generated marshaler propagates this error up through
// MarshalJSON, so it must be a clean error return rather than a crash.
// Inputs: a malformed data document and a malformed patch document. Output: a
// non-nil error in both cases.
// Data choice: a bare unquoted token is the minimal input that fails json
// decoding for each argument position.
func TestJSONMergeInvalidJSON(t *testing.T) {
	if _, err := JSONMerge([]byte(`{not json}`), []byte(`{}`)); err == nil {
		t.Error("expected error for invalid data JSON, got nil")
	}
	if _, err := JSONMerge([]byte(`{}`), []byte(`{not json}`)); err == nil {
		t.Error("expected error for invalid patch JSON, got nil")
	}
}
