/*
 *
 *  MIT License
 *
 *  (C) Copyright 2026 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package nautobot

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/commands"
	"github.com/Cray-HPE/cani/pkg/provider/nautobot/export"
	"gopkg.in/yaml.v3"
)

// setProviderConfigYAML installs the nautobot section of a config file,
// decoded from the text the file would hold so a test reads the values in
// the shapes the config loader produces.
func setProviderConfigYAML(t *testing.T, section string) {
	t.Helper()
	block := map[string]any{}
	if err := yaml.Unmarshal([]byte(section), &block); err != nil {
		t.Fatalf("config section does not parse: %v", err)
	}
	setProviderConfig(t, block)
}

// newExportCmd returns the provider's own export command, so a test reads
// the flags as they are registered.
func newExportCmd(t *testing.T) *cli.Command {
	t.Helper()
	cmd, err := commands.NewExportCommand(nil)
	if err != nil {
		t.Fatalf("NewExportCommand: %v", err)
	}
	return cmd
}

// TestExportStringsFromConfig_ReadsAListOrOneString verifies the option is
// read from a YAML list in either style and from one comma-separated
// string, and that a key with no value is no entries.
//
// Why it matters: the option is typed by hand; reading only one of the
// forms a list takes in YAML would leave the others exporting nothing.
// Inputs: the nautobot config section, one written form per case.
// Outputs: the entries in the order written, and no error.
// Data choice: the string form carries stray spaces and a trailing comma,
// which are the two ways a one-line list is usually untidy.
func TestExportStringsFromConfig_ReadsAListOrOneString(t *testing.T) {
	both := []string{"csm.xname", "csm.nid=node_id"}
	tests := []struct {
		name    string
		section string
		want    []string
	}{
		{"block list", "export:\n  provider_metadata_fields:\n    - csm.xname\n    - csm.nid=node_id\n", both},
		{"flow list", "export:\n  provider_metadata_fields: [csm.xname, csm.nid=node_id]\n", both},
		{"one comma-separated string", "export:\n  provider_metadata_fields: ' csm.xname ,csm.nid=node_id, '\n", both},
		{"empty list", "export:\n  provider_metadata_fields: []\n", []string{}},
		{"key with no value", "export:\n  provider_metadata_fields:\n", nil},
		{"no key", "export:\n  merge: true\n", nil},
		{"no export section", "url: http://nb.example/api\n", nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			setProviderConfigYAML(t, tc.section)

			// Act.
			got, err := New().exportStringsFromConfig(providerMetadataFieldsKey)

			// Assert.
			if err != nil {
				t.Fatalf("exportStringsFromConfig: %v", err)
			}
			if len(got) != len(tc.want) || (len(got) > 0 && !reflect.DeepEqual(got, tc.want)) {
				t.Errorf("entries = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestExportStringsFromConfig_RefusesWhatIsNotAList verifies a value that
// is neither a list of text entries nor a string is an error naming the
// option.
//
// Why it matters: "- csm.xname: node_id" and "csm.xname: node_id" both
// look like the option and decode to mappings; read leniently they name no
// provider key, so the export would send nothing and say nothing.
// Inputs: the nautobot config section, one mistaken form per case.
// Outputs: an error that names provider_metadata_fields.
// Data choice: a colon in place of "=" in a list entry and under the key
// are the two ways the syntax is most easily mistyped in YAML; a number
// stands for any scalar that is not text.
func TestExportStringsFromConfig_RefusesWhatIsNotAList(t *testing.T) {
	tests := []struct {
		name    string
		section string
	}{
		{"mapping in place of the list", "export:\n  provider_metadata_fields:\n    csm.xname: node_id\n"},
		{"entry written with a colon", "export:\n  provider_metadata_fields:\n    - csm.xname: node_id\n"},
		{"number", "export:\n  provider_metadata_fields: 7\n"},
		{"list holding a number", "export:\n  provider_metadata_fields: [csm.xname, 7]\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			setProviderConfigYAML(t, tc.section)

			// Act.
			got, err := New().exportStringsFromConfig(providerMetadataFieldsKey)

			// Assert.
			if err == nil {
				t.Fatalf("exportStringsFromConfig = %q, want an error", got)
			}
			if !strings.Contains(err.Error(), providerMetadataFieldsKey) {
				t.Errorf("error = %q, want it to name %s", err, providerMetadataFieldsKey)
			}
		})
	}
}

// TestProviderMetadataFields_FlagReplacesTheConfiguredList verifies the
// export uses the configured entries when --provider-metadata-field is not
// given and only the flag's entries when it is.
//
// Why it matters: a flag that added to the configured list could not be
// used to export less than the config asks for, and one that was not read
// would leave the command line with no say at all.
// Inputs: a config listing csm.xname; the provider's export command with
// the flag absent, then given once comma-separated and once more.
// Outputs: csm.xname alone, then the flag's three entries without it.
// Data choice: the flag is given both ways it can be written, and one of
// its entries renames the key, so the whole entry syntax passes through.
func TestProviderMetadataFields_FlagReplacesTheConfiguredList(t *testing.T) {
	// Arrange.
	setProviderConfigYAML(t, "export:\n  provider_metadata_fields: [csm.xname]\n")
	flagged := newExportCmd(t)
	mustSet(t, flagged, providerMetadataFieldsFlag, "hpcm.hpcm_uuid,csm.nid=node_id")
	mustSet(t, flagged, providerMetadataFieldsFlag, "redfish.bmc_fqdn")

	// Act.
	configured, configErr := New().providerMetadataFields(newExportCmd(t))
	p := New()
	replaced, flagErr := p.providerMetadataFields(flagged)

	// Assert.
	if configErr != nil || flagErr != nil {
		t.Fatalf("providerMetadataFields: config %v, flag %v", configErr, flagErr)
	}
	wantConfigured := []export.ProviderMetadataField{{Provider: "csm", Key: "xname", CustomField: "xname"}}
	if !reflect.DeepEqual(configured, wantConfigured) {
		t.Errorf("without the flag = %+v, want %+v", configured, wantConfigured)
	}
	wantReplaced := []export.ProviderMetadataField{
		{Provider: "hpcm", Key: "hpcm_uuid", CustomField: "hpcm_uuid"},
		{Provider: "csm", Key: "nid", CustomField: "node_id"},
		{Provider: "redfish", Key: "bmc_fqdn", CustomField: "bmc_fqdn"},
	}
	if !reflect.DeepEqual(replaced, wantReplaced) {
		t.Errorf("with the flag = %+v, want %+v", replaced, wantReplaced)
	}
	wantEntries := []string{"hpcm.hpcm_uuid", "csm.nid=node_id", "redfish.bmc_fqdn"}
	if !reflect.DeepEqual(p.Options.Export.ProviderMetadataFields, wantEntries) {
		t.Errorf("recorded option = %q, want %q", p.Options.Export.ProviderMetadataFields, wantEntries)
	}
}

// TestProviderMetadataFields_FlagIsNotHeldUpByTheConfig verifies a config
// value that cannot be read does not fail an export whose flag replaces it.
//
// Why it matters: the flag replaces the configured list, so a mistake in a
// value that is not used should not stop the run that works around it.
// Inputs: a config whose option is a mapping, and the flag naming
// csm.xname.
// Outputs: the flag's entry and no error.
// Data choice: the mapping is the mistake the config reader refuses on its
// own, which is what makes its being ignored here observable.
func TestProviderMetadataFields_FlagIsNotHeldUpByTheConfig(t *testing.T) {
	// Arrange.
	setProviderConfigYAML(t, "export:\n  provider_metadata_fields:\n    csm.xname: node_id\n")
	cmd := newExportCmd(t)
	mustSet(t, cmd, providerMetadataFieldsFlag, "csm.xname")

	// Act.
	fields, err := New().providerMetadataFields(cmd)

	// Assert.
	if err != nil {
		t.Fatalf("providerMetadataFields: %v", err)
	}
	want := []export.ProviderMetadataField{{Provider: "csm", Key: "xname", CustomField: "xname"}}
	if !reflect.DeepEqual(fields, want) {
		t.Errorf("fields = %+v, want %+v", fields, want)
	}
}

// statusesOnly answers the connection check an export setup makes and
// counts every request it receives.
func statusesOnly(t *testing.T, requests *atomic.Int32) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":0,"next":null,"previous":null,"results":[]}`))
	}))
	t.Cleanup(srv.Close)
	return srv
}

// exportConfigFor is a nautobot config section that reaches srv and needs
// no prompt, with the given provider_metadata_fields value.
func exportConfigFor(srv *httptest.Server, fields string) string {
	return "url: " + srv.URL + "\ntoken: test-token\n" +
		"default_location: DC1\ndefault_role: Compute\ndefault_status: Active\n" +
		"export:\n  provider_metadata_fields: " + fields + "\n"
}

// TestSetupExportFromConfig_HandsTheListToTheExporter verifies the entries
// the config lists are the projection the exporter is built with.
//
// Why it matters: every write takes the projection from the exporter's
// options; a list that was parsed and not passed on would export nothing
// while the config said otherwise.
// Inputs: a config listing csm.xname and a renamed hpcm.hpcm_uuid, and a
// server that answers the connection check.
// Outputs: an exporter whose options hold both fields.
// Data choice: one plain and one renamed entry, so both entry forms are
// seen arriving.
func TestSetupExportFromConfig_HandsTheListToTheExporter(t *testing.T) {
	// Arrange.
	var requests atomic.Int32
	srv := statusesOnly(t, &requests)
	setProviderConfigYAML(t, exportConfigFor(srv, "[csm.xname, hpcm.hpcm_uuid=hpcm_id]"))
	p := New()

	// Act.
	err := p.setupExportFromConfig(newExportCmd(t))

	// Assert.
	if err != nil {
		t.Fatalf("setupExportFromConfig: %v", err)
	}
	want := []export.ProviderMetadataField{
		{Provider: "csm", Key: "xname", CustomField: "xname"},
		{Provider: "hpcm", Key: "hpcm_uuid", CustomField: "hpcm_id"},
	}
	if p.exporter == nil {
		t.Fatal("setupExportFromConfig built no exporter")
	}
	if got := p.exporter.Options.ProviderMetadataFields; !reflect.DeepEqual(got, want) {
		t.Errorf("exporter fields = %+v, want %+v", got, want)
	}
}

// TestSetupExportFromConfig_RefusesABadEntryBeforeConnecting verifies an
// entry that cannot be read fails the export, quoting the entry, before
// Nautobot is contacted.
//
// Why it matters: an export that went ahead without the field would write
// every object and leave the operator to find the field missing afterwards.
// Inputs: a reachable server, and an entry with no provider given in the
// config in one case and on the flag in the other.
// Outputs: an error that quotes the entry, and no request to the server.
// Data choice: "xname" alone is the likeliest slip, a key with its
// provider left off.
func TestSetupExportFromConfig_RefusesABadEntryBeforeConnecting(t *testing.T) {
	tests := []struct {
		name   string
		config string
		flag   string
	}{
		{"in the config file", "[csm.nid, xname]", ""},
		{"on the flag", "[csm.nid]", "xname"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange.
			var requests atomic.Int32
			srv := statusesOnly(t, &requests)
			setProviderConfigYAML(t, exportConfigFor(srv, tc.config))
			cmd := newExportCmd(t)
			if tc.flag != "" {
				mustSet(t, cmd, providerMetadataFieldsFlag, tc.flag)
			}

			// Act.
			err := New().setupExportFromConfig(cmd)

			// Assert.
			if err == nil || !strings.Contains(err.Error(), `"xname"`) {
				t.Fatalf("setupExportFromConfig error = %v, want one quoting \"xname\"", err)
			}
			if n := requests.Load(); n != 0 {
				t.Errorf("Nautobot received %d requests before the entry was refused, want 0", n)
			}
		})
	}
}
