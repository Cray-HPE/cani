package transform

import (
	"os"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
	import_ "github.com/Cray-HPE/cani/pkg/provider/redfish/import"
)

// fakeRedfishProvider satisfies the anonymous provider interface consumed by
// Transform, returning a canned list of ServiceRoots.
type fakeRedfishProvider struct {
	roots []import_.ServiceRoot
}

func (f fakeRedfishProvider) GetRoots() []import_.ServiceRoot { return f.roots }

// withProviderGetter installs a provider that returns roots and restores the
// previous getter when the test ends.
func withProviderGetter(t *testing.T, roots []import_.ServiceRoot) {
	t.Helper()
	orig := providerGetter
	t.Cleanup(func() { providerGetter = orig })
	SetProviderGetter(func() interface {
		GetRoots() []import_.ServiceRoot
	} {
		return fakeRedfishProvider{roots: roots}
	})
}

// TestTransform_TransformsProviderRoots verifies Transform pulls ServiceRoots from
// the registered provider and converts each into a CANI device.
//
// Why it matters: Transform is the provider entry point the import command calls,
// so roots collected from a BMC must surface as inventory devices; this also
// proves SetProviderGetter wires the singleton in.
// Inputs: an empty existing inventory; the provider returns one testRoot().
// Outputs: a TransformResult holding one device whose Name equals the root
// Product.
// Data choice: a single well-formed root proves the getter is consulted and the
// one-root-to-one-device contract holds without ordering ambiguity.
func TestTransform_TransformsProviderRoots(t *testing.T) {
	withProviderGetter(t, []import_.ServiceRoot{testRoot()})

	result, err := Transform(devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("Transform() error = %v, want nil", err)
	}
	if len(result.Devices) != 1 {
		t.Fatalf("Devices = %d, want 1", len(result.Devices))
	}
	for _, dev := range result.Devices {
		if dev.Name != wantProduct {
			t.Errorf("device Name = %q, want %q", dev.Name, wantProduct)
		}
	}
}

// TestTransform_NoRoots_ReturnsEmpty verifies Transform returns an empty result
// and no error when the provider yields zero ServiceRoots.
//
// Why it matters: importing from a source with nothing to offer must be a
// graceful no-op rather than a failure that aborts the import pipeline.
// Inputs: an empty existing inventory; the provider returns nil roots. Outputs: a
// non-nil TransformResult with zero devices and zero modules.
// Data choice: a nil root slice is the canonical "nothing to import" signal the
// early-return guard checks for.
func TestTransform_NoRoots_ReturnsEmpty(t *testing.T) {
	withProviderGetter(t, nil)

	result, err := Transform(devicetypes.Inventory{})
	if err != nil {
		t.Fatalf("Transform() error = %v, want nil", err)
	}
	if len(result.Devices) != 0 {
		t.Errorf("Devices = %d, want 0", len(result.Devices))
	}
	if len(result.Modules) != 0 {
		t.Errorf("Modules = %d, want 0", len(result.Modules))
	}
}

// TestBuildProviderMetadata_MinimalRoot verifies optional BMC and HPE keys are
// omitted when the ServiceRoot carries no OEM data.
//
// Why it matters: provider metadata should record only what the BMC actually
// reported, so absent fields must not appear as empty-string noise.
// Inputs: a ServiceRoot with the four base fields set and no OEM block. Outputs:
// the nested redfish sub-map.
// Data choice: omitting OEM data exercises the false branch of every optional
// non-empty-value guard, complementing the full-root metadata test.
func TestBuildProviderMetadata_MinimalRoot(t *testing.T) {
	root := import_.ServiceRoot{
		OdataType:      "#ServiceRoot.v1_13_0.ServiceRoot",
		Product:        "Generic Server",
		Vendor:         "ACME",
		UUID:           "11111111-1111-1111-1111-111111111111",
		RedfishVersion: "1.0.0",
	}

	meta := buildProviderMetadata(root)
	redfishMeta, ok := meta[providerKeyRedfish].(map[string]any)
	if !ok {
		t.Fatal("metadata must nest under the redfish key")
	}
	for _, key := range []string{"redfish_version", metaKeyRedfishUUID, "vendor", "odata_type"} {
		if _, present := redfishMeta[key]; !present {
			t.Errorf("missing base key %q", key)
		}
	}
	for _, key := range []string{"bmc_type", "bmc_firmware", "bmc_fqdn", "bmc_hostname", "product_tag", "system_family"} {
		if _, present := redfishMeta[key]; present {
			t.Errorf("optional key %q should be absent for a minimal root", key)
		}
	}
}

// withStepMode enables interactive StepMode, feeds stdinData to os.Stdin, and
// silences os.Stdout for the duration of the test. All swapped globals are
// restored on cleanup.
func withStepMode(t *testing.T, stdinData string) {
	t.Helper()

	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe() error = %v", err)
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open %s error = %v", os.DevNull, err)
	}

	origCfg, origStdin, origStdout := config.Cfg, os.Stdin, os.Stdout
	config.Cfg = &config.Config{StepMode: true, NoColor: true}
	os.Stdin = stdinR
	os.Stdout = devNull

	go func() {
		_, _ = stdinW.WriteString(stdinData)
		_ = stdinW.Close()
	}()

	t.Cleanup(func() {
		os.Stdin, os.Stdout, config.Cfg = origStdin, origStdout, origCfg
		_ = stdinR.Close()
		_ = devNull.Close()
	})
}

// TestTransformRoots_StepModeAdvancesOnEnter verifies step-through mode renders
// each root and proceeds when the operator presses Enter.
//
// Why it matters: StepMode lets operators inspect every mapping before it is
// committed, so the transform must pause and then continue cleanly per root.
// Inputs: one testRoot() with StepMode enabled and a newline queued on stdin.
// Outputs: a TransformResult with one device and no error.
// Data choice: a single newline is the minimal input that satisfies the per-root
// "Press Enter" prompt for exactly one root.
func TestTransformRoots_StepModeAdvancesOnEnter(t *testing.T) {
	withStepMode(t, "\n")

	result, err := transformRoots([]import_.ServiceRoot{testRoot()}, nil)
	if err != nil {
		t.Fatalf("transformRoots() error = %v, want nil", err)
	}
	if len(result.Devices) != 1 {
		t.Errorf("Devices = %d, want 1", len(result.Devices))
	}
}

// TestTransformRoots_StepModeInterrupted verifies a closed stdin (EOF) at the
// prompt aborts the transform with a wrapped "step interrupted" error.
//
// Why it matters: if the operator's input stream ends mid-review, the import must
// fail loudly instead of silently committing partial work.
// Inputs: one testRoot() with StepMode enabled and an immediately closed stdin.
// Outputs: a nil result and an error containing "step interrupted".
// Data choice: empty stdin data makes ReadString hit EOF on the first prompt,
// driving the error-wrap branch.
func TestTransformRoots_StepModeInterrupted(t *testing.T) {
	withStepMode(t, "")

	result, err := transformRoots([]import_.ServiceRoot{testRoot()}, nil)
	if err == nil {
		t.Fatal("transformRoots() error = nil, want a step-interrupted error")
	}
	if !strings.Contains(err.Error(), "step interrupted") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "step interrupted")
	}
	if result != nil {
		t.Errorf("result = %+v, want nil on interruption", result)
	}
}
