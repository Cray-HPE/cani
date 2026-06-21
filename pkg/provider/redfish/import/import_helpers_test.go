package import_

import "testing"

// TestDeduplicationKeyUsesHostnameWhenFQDNMissing verifies the deduplication key
// falls back to BMC hostname when FQDN is absent.
//
// Why it matters: some ServiceRoot responses include only a short hostname, and
// the import must still distinguish endpoints that share a Redfish UUID.
// Inputs: a root with UUID "aaa" and an HPE manager HostName of "host-a".
// Outputs: the composite key "aaa|host-a".
// Data choice: leaving FQDN empty isolates the hostname fallback branch.
func TestDeduplicationKeyUsesHostnameWhenFQDNMissing(t *testing.T) {
	root := ServiceRoot{
		UUID: "aaa",
		Oem:  OemData{Hpe: &HpeOem{Manager: []HpeManager{{HostName: "host-a"}}}},
	}

	got := deduplicationKey(root)

	if got != "aaa|host-a" {
		t.Errorf("deduplicationKey() = %q, want %q", got, "aaa|host-a")
	}
}

// TestFormatParsedRootWithoutManager verifies parsed display text omits manager
// details when no BMC manager type is available.
//
// Why it matters: StepMode should present minimal non-HPE or sparse Redfish roots
// without adding empty parentheses or misleading firmware text.
// Inputs: a ServiceRoot with only Product set. Outputs: "server: Generic".
// Data choice: an OEM-free root drives the no-manager formatting branch.
func TestFormatParsedRootWithoutManager(t *testing.T) {
	got := formatParsedRoot(ServiceRoot{Product: "Generic"})

	if got != "server: Generic" {
		t.Errorf("formatParsedRoot() = %q, want %q", got, "server: Generic")
	}
}

// TestFormatIdentifierFallbackOrder verifies record identifiers prefer FQDN,
// then hostname, then UUID.
//
// Why it matters: StepMode uses this identifier to help operators recognize the
// raw record being reviewed, so the most specific BMC identity should win.
// Inputs: roots with FQDN+hostname+UUID, hostname+UUID, and UUID only. Outputs:
// the selected identifier for each case.
// Data choice: each case removes one higher-priority field to isolate the next
// fallback in the chain.
func TestFormatIdentifierFallbackOrder(t *testing.T) {
	cases := []struct {
		name string
		root ServiceRoot
		want string
	}{
		{
			name: "fqdn preferred",
			root: ServiceRoot{UUID: "uuid-a", Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{{
				FQDN: "fqdn.example.com", HostName: "host-a",
			}}}}},
			want: "fqdn.example.com",
		},
		{
			name: "hostname fallback",
			root: ServiceRoot{UUID: "uuid-b", Oem: OemData{Hpe: &HpeOem{Manager: []HpeManager{{
				HostName: "host-b",
			}}}}},
			want: "host-b",
		},
		{name: "uuid fallback", root: ServiceRoot{UUID: "uuid-c"}, want: "uuid-c"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatIdentifier(tt.root); got != tt.want {
				t.Errorf("formatIdentifier() = %q, want %q", got, tt.want)
			}
		})
	}
}
