package csm

import (
	"testing"

	"github.com/Cray-HPE/cani/internal/cli"
	"github.com/Cray-HPE/cani/pkg/devicetypes"
)

func TestRegisterAndApplyDeviceUpdateFlags(t *testing.T) {
	p := New()
	cmd := &cli.Command{}
	p.RegisterDeviceUpdateFlags(cmd)

	if cmd.Flags().Lookup("nid") == nil {
		t.Fatal("expected --nid flag to be registered")
	}
	if cmd.Flags().Lookup("alias") == nil {
		t.Fatal("expected --alias flag to be registered")
	}

	if err := cmd.Flags().Set("nid", "42"); err != nil {
		t.Fatalf("set nid: %v", err)
	}
	if err := cmd.Flags().Set("alias", "nid000001"); err != nil {
		t.Fatalf("set alias: %v", err)
	}

	dev := &devicetypes.CaniDeviceType{}
	if err := p.ApplyDeviceUpdateFlags(cmd, dev); err != nil {
		t.Fatalf("apply: %v", err)
	}

	sub, ok := dev.GetProviderSubMap(p.Slug())
	if !ok {
		t.Fatal("expected csm provider sub-map")
	}
	if sub["nid"] != 42 {
		t.Errorf("nid = %v, want 42", sub["nid"])
	}
	aliases, ok := sub["aliases"].([]string)
	if !ok || len(aliases) != 1 || aliases[0] != "nid000001" {
		t.Errorf("aliases = %v, want [nid000001]", sub["aliases"])
	}
}

func TestApplyDeviceUpdateFlagsNoChange(t *testing.T) {
	p := New()
	cmd := &cli.Command{}
	p.RegisterDeviceUpdateFlags(cmd)

	dev := &devicetypes.CaniDeviceType{}
	if err := p.ApplyDeviceUpdateFlags(cmd, dev); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if _, ok := dev.GetProviderSubMap(p.Slug()); ok {
		t.Error("expected no provider metadata when flags unchanged")
	}
}
