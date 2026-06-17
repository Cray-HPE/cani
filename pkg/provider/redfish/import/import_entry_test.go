package import_

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Cray-HPE/cani/internal/config"
	"github.com/Cray-HPE/cani/pkg/provider/redfish/commands"
)

type fakeImportProvider struct {
	clearCalls int
	setCalls   int
	roots      []ServiceRoot
}

func (f *fakeImportProvider) ClearRoots() {
	f.clearCalls++
	f.roots = nil
}

func (f *fakeImportProvider) SetRoots(roots []ServiceRoot) {
	f.setCalls++
	f.roots = append([]ServiceRoot(nil), roots...)
}

func installImportProvider(t *testing.T) *fakeImportProvider {
	t.Helper()
	provider := &fakeImportProvider{}
	origGetter := providerGetter
	origRootFlag := commands.RootFlag
	origCfg := config.Cfg
	SetProviderGetter(func() interface {
		ClearRoots()
		SetRoots(roots []ServiceRoot)
	} {
		return provider
	})
	commands.RootFlag = ""
	config.Cfg = nil
	t.Cleanup(func() {
		providerGetter = origGetter
		commands.RootFlag = origRootFlag
		config.Cfg = origCfg
	})
	return provider
}

// TestImportReadsRootFileDeduplicatesAndStoresProvider verifies Import reads the
// root file, parses records, deduplicates them, and stores the result on the
// provider singleton.
//
// Why it matters: Import is the command entry point that feeds raw ServiceRoots
// into transform, so it must bridge file input to provider state without doing
// transformation itself.
// Inputs: a temporary file containing the same realistic ServiceRoot twice.
// Outputs: one stored root, one ClearRoots call, and one SetRoots call.
// Data choice: duplicate copies of redfish-root.json exercise the file path and
// duplicate-removal behavior in a compact fixture.
func TestImportReadsRootFileDeduplicatesAndStoresProvider(t *testing.T) {
	provider := installImportProvider(t)
	single := loadFixture(t, fixtureFile)
	path := filepath.Join(t.TempDir(), "roots.json")
	if err := os.WriteFile(path, []byte("["+string(single)+","+string(single)+"]"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	commands.RootFlag = path

	if err := Import(nil, nil, nil); err != nil {
		t.Fatalf("Import() error = %v, want nil", err)
	}

	if provider.clearCalls != 1 {
		t.Errorf("ClearRoots calls = %d, want 1", provider.clearCalls)
	}
	if provider.setCalls != 1 {
		t.Errorf("SetRoots calls = %d, want 1", provider.setCalls)
	}
	if len(provider.roots) != 1 {
		t.Fatalf("stored roots = %d, want 1", len(provider.roots))
	}
	root := provider.roots[0]
	assertField(t, "Product", root.Product, wantProduct)
	assertField(t, "UUID", root.UUID, "946a7d44-9967-4940-9490-f2d581950512")
	assertField(t, "ManagerFQDN", root.ManagerFQDN(), "foo.example.com")
}

// TestImportReadsStdinAndStoresDistinctRoots verifies Import reads stdin when no
// root file is configured and stores every distinct root.
//
// Why it matters: the Redfish command supports piping ServiceRoot JSON, and that
// path must feed the same provider state as file input.
// Inputs: os.Stdin redirected to the redfish-root-array.json fixture with
// commands.RootFlag empty. Outputs: two stored roots in fixture order.
// Data choice: the array fixture has two distinct BMC FQDNs, proving stdin input
// and non-duplicate preservation together.
func TestImportReadsStdinAndStoresDistinctRoots(t *testing.T) {
	provider := installImportProvider(t)
	stdinPath := filepath.Join(t.TempDir(), arrayFixtureFile)
	if err := os.WriteFile(stdinPath, loadFixture(t, arrayFixtureFile), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = stdin.Close()
	})

	if err := Import(nil, nil, nil); err != nil {
		t.Fatalf("Import() error = %v, want nil", err)
	}

	if provider.clearCalls != 1 || provider.setCalls != 1 {
		t.Fatalf("provider calls = clear %d / set %d, want 1 / 1", provider.clearCalls, provider.setCalls)
	}
	if len(provider.roots) != 2 {
		t.Fatalf("stored roots = %d, want 2", len(provider.roots))
	}
	assertField(t, "roots[0].ManagerFQDN", provider.roots[0].ManagerFQDN(), "bin.example.com")
	assertField(t, "roots[1].ManagerFQDN", provider.roots[1].ManagerFQDN(), "baz.example.com")
}

// TestImportReturnsParseErrorWithoutTouchingProvider verifies invalid Redfish
// JSON returns an error before provider state is cleared or replaced.
//
// Why it matters: a bad import should leave the previously loaded provider roots
// untouched so transform cannot consume partial or empty state by accident.
// Inputs: a temporary root file containing object-shaped JSON with no ServiceRoot
// identity fields. Outputs: an error and zero provider calls.
// Data choice: syntactically valid but semantically invalid JSON exercises the
// parser validation failure after readInput succeeds.
func TestImportReturnsParseErrorWithoutTouchingProvider(t *testing.T) {
	provider := installImportProvider(t)
	path := filepath.Join(t.TempDir(), "not-root.json")
	if err := os.WriteFile(path, []byte(`{"not":"a service root"}`), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	commands.RootFlag = path

	err := Import(nil, nil, nil)

	if err == nil {
		t.Fatal("Import() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), "does not appear to be a Redfish ServiceRoot") {
		t.Errorf("Import() error = %q, want ServiceRoot validation message", err.Error())
	}
	if provider.clearCalls != 0 || provider.setCalls != 0 {
		t.Errorf("provider calls = clear %d / set %d, want 0 / 0", provider.clearCalls, provider.setCalls)
	}
}

// TestImportReturnsReadErrorWithoutTouchingProvider verifies a missing root file
// fails before provider state is cleared or replaced.
//
// Why it matters: filesystem errors should not wipe previously imported raw
// roots, because transform depends on the provider retaining the last valid
// import state.
// Inputs: commands.RootFlag pointing at a nonexistent file. Outputs: a wrapped
// read error and zero provider calls.
// Data choice: a path under t.TempDir is guaranteed absent and keeps the test
// isolated from the developer's filesystem.
func TestImportReturnsReadErrorWithoutTouchingProvider(t *testing.T) {
	provider := installImportProvider(t)
	missingPath := filepath.Join(t.TempDir(), "missing-redfish-root.json")
	commands.RootFlag = missingPath

	err := Import(nil, nil, nil)

	if err == nil {
		t.Fatal("Import() error = nil, want file read error")
	}
	if !strings.Contains(err.Error(), "reading file "+missingPath) {
		t.Errorf("Import() error = %q, want missing file path", err.Error())
	}
	if provider.clearCalls != 0 || provider.setCalls != 0 {
		t.Errorf("provider calls = clear %d / set %d, want 0 / 0", provider.clearCalls, provider.setCalls)
	}
}

// TestImportReturnsStdinReadErrorWithoutTouchingProvider verifies stdin read
// errors are returned before provider state changes.
//
// Why it matters: stdin is a supported import source, and an I/O failure there
// must behave like a file read failure rather than clearing the provider.
// Inputs: os.Stdin temporarily set to a closed file and commands.RootFlag empty.
// Outputs: a wrapped stdin read error and zero provider calls.
// Data choice: reading a closed temp file is a deterministic local way to force
// io.ReadAll(os.Stdin) to return an error without external dependencies.
func TestImportReturnsStdinReadErrorWithoutTouchingProvider(t *testing.T) {
	provider := installImportProvider(t)
	stdinPath := filepath.Join(t.TempDir(), "stdin.txt")
	stdin, err := os.Create(stdinPath)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if err := stdin.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = stdin
	t.Cleanup(func() { os.Stdin = origStdin })

	err = Import(nil, nil, nil)

	if err == nil {
		t.Fatal("Import() error = nil, want stdin read error")
	}
	if !strings.Contains(err.Error(), "reading stdin") {
		t.Errorf("Import() error = %q, want stdin read context", err.Error())
	}
	if provider.clearCalls != 0 || provider.setCalls != 0 {
		t.Errorf("provider calls = clear %d / set %d, want 0 / 0", provider.clearCalls, provider.setCalls)
	}
}

// TestImportStepModeInterruptedDoesNotStoreRoots verifies EOF during StepMode
// aborts the import before provider state is updated.
//
// Why it matters: operator review mode must fail visibly on interrupted input and
// avoid committing records the operator did not approve.
// Inputs: one valid root file, StepMode enabled, stdin redirected to an empty
// file, and stdout redirected to os.DevNull. Outputs: a wrapped step-interrupted
// error and zero provider calls.
// Data choice: empty stdin deterministically triggers EOF at the first prompt
// without sleeps or external interaction.
func TestImportStepModeInterruptedDoesNotStoreRoots(t *testing.T) {
	provider := installImportProvider(t)
	path := filepath.Join(t.TempDir(), fixtureFile)
	if err := os.WriteFile(path, loadFixture(t, fixtureFile), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	commands.RootFlag = path
	config.Cfg = &config.Config{StepMode: true, NoColor: true}

	stdinPath := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(stdinPath, nil, 0o600); err != nil {
		t.Fatalf("WriteFile(stdin) error = %v", err)
	}
	stdin, err := os.Open(stdinPath)
	if err != nil {
		t.Fatalf("Open(stdin) error = %v", err)
	}
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("OpenFile(%s) error = %v", os.DevNull, err)
	}
	origStdin, origStdout := os.Stdin, os.Stdout
	os.Stdin, os.Stdout = stdin, devNull
	t.Cleanup(func() {
		os.Stdin, os.Stdout = origStdin, origStdout
		_ = stdin.Close()
		_ = devNull.Close()
	})

	err = Import(nil, nil, nil)

	if err == nil {
		t.Fatal("Import() error = nil, want step interrupted")
	}
	if !strings.Contains(err.Error(), "step interrupted") {
		t.Errorf("Import() error = %q, want step interrupted", err.Error())
	}
	if provider.clearCalls != 0 || provider.setCalls != 0 {
		t.Errorf("provider calls = clear %d / set %d, want 0 / 0", provider.clearCalls, provider.setCalls)
	}
}
