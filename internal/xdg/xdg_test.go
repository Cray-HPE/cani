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
package xdg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFreshInstall_UsesXDG(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	got := ConfigHome()
	want := filepath.Join(tmp, ".config", appName)
	if got != want {
		t.Errorf("ConfigHome() = %q, want %q", got, want)
	}

	got = DataHome()
	want = filepath.Join(tmp, ".local", "share", appName)
	if got != want {
		t.Errorf("DataHome() = %q, want %q", got, want)
	}

	got = CacheHome()
	want = filepath.Join(tmp, ".cache", appName)
	if got != want {
		t.Errorf("CacheHome() = %q, want %q", got, want)
	}
}

func TestLegacyExists_FallsBack(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	legacy := filepath.Join(tmp, "."+appName)
	os.MkdirAll(legacy, 0o755)

	got := ConfigHome()
	if got != legacy {
		t.Errorf("ConfigHome() = %q, want legacy %q", got, legacy)
	}

	got = DataHome()
	if got != legacy {
		t.Errorf("DataHome() = %q, want legacy %q", got, legacy)
	}
}

func TestExplicitEnvVar_OverridesLegacy(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)

	legacy := filepath.Join(tmp, "."+appName)
	os.MkdirAll(legacy, 0o755)

	customConfig := filepath.Join(tmp, "myconfig")
	t.Setenv("XDG_CONFIG_HOME", customConfig)
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	got := ConfigHome()
	want := filepath.Join(customConfig, appName)
	if got != want {
		t.Errorf("ConfigHome() = %q, want %q", got, want)
	}

	got = DataHome()
	if got != legacy {
		t.Errorf("DataHome() = %q, want legacy %q", got, legacy)
	}
}

func TestXDGDirExists_PreferredOverLegacy(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	legacy := filepath.Join(tmp, "."+appName)
	os.MkdirAll(legacy, 0o755)

	xdgConfig := filepath.Join(tmp, ".config", appName)
	os.MkdirAll(xdgConfig, 0o755)

	got := ConfigHome()
	if got != xdgConfig {
		t.Errorf("ConfigHome() = %q, want XDG %q", got, xdgConfig)
	}
}

func TestConfigFilePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	got := ConfigFilePath()
	want := filepath.Join(tmp, ".config", appName, appName+".yml")
	if got != want {
		t.Errorf("ConfigFilePath() = %q, want %q", got, want)
	}
}

func TestDatastorePath(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	got := DatastorePath()
	want := filepath.Join(tmp, ".local", "share", appName, appName+"db.json")
	if got != want {
		t.Errorf("DatastorePath() = %q, want %q", got, want)
	}
}

func TestTypesCacheDir(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("XDG_DATA_HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")

	got := TypesCacheDir()
	want := filepath.Join(tmp, ".cache", appName, "types-cache")
	if got != want {
		t.Errorf("TypesCacheDir() = %q, want %q", got, want)
	}
}
