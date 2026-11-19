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

// Package xdg resolves XDG Base Directory paths for CANI with a fallback to
// the legacy ~/.cani/ layout. If the legacy directory exists and the XDG
// locations do not, the legacy paths are used to avoid breaking existing
// installations.
package xdg

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "cani"

// ConfigHome returns the directory for configuration files.
// XDG: $XDG_CONFIG_HOME/cani (default: ~/.config/cani)
// Fallback: ~/.cani
func ConfigHome() string {
	if dir := resolveXDG("XDG_CONFIG_HOME", ".config"); dir != "" {
		return dir
	}
	return legacyDir()
}

// DataHome returns the directory for persistent data (datastore, types-cache).
// XDG: $XDG_DATA_HOME/cani (default: ~/.local/share/cani)
// Fallback: ~/.cani
func DataHome() string {
	if dir := resolveXDG("XDG_DATA_HOME", ".local/share"); dir != "" {
		return dir
	}
	return legacyDir()
}

// CacheHome returns the directory for cache files.
// XDG: $XDG_CACHE_HOME/cani (default: ~/.cache/cani)
// Fallback: ~/.cani
func CacheHome() string {
	if dir := resolveXDG("XDG_CACHE_HOME", ".cache"); dir != "" {
		return dir
	}
	return legacyDir()
}

// ConfigFilePath returns the full path to the config file.
func ConfigFilePath() string {
	return filepath.Join(ConfigHome(), appName+".yml")
}

// DatastorePath returns the full path to the JSON datastore.
func DatastorePath() string {
	return filepath.Join(DataHome(), appName+"db.json")
}

// TypesCacheDir returns the directory for cloned types repositories.
func TypesCacheDir() string {
	return filepath.Join(CacheHome(), "types-cache")
}

// resolveXDG checks whether the XDG path should be used. It returns the
// resolved app directory if:
//   - the env var is set explicitly, OR
//   - the XDG default app dir already exists, OR
//   - the legacy dir does NOT exist (fresh install)
//
// Otherwise it returns "" to signal fallback to legacy.
func resolveXDG(envVar, defaultRelative string) string {
	home := homeDir()
	if home == "" {
		return ""
	}

	// Determine the base XDG directory.
	base := os.Getenv(envVar)
	envExplicit := base != ""
	if base == "" {
		base = filepath.Join(home, defaultRelative)
	}

	appDir := filepath.Join(base, appName)

	// If the env var was explicitly set, always use XDG.
	if envExplicit {
		return appDir
	}

	// If the XDG app dir already exists, use it.
	if dirExists(appDir) {
		return appDir
	}

	// If the legacy dir does NOT exist (fresh install), use XDG.
	if !dirExists(legacyDir()) {
		return appDir
	}

	// Legacy dir exists but XDG dir does not — fall back to legacy.
	return ""
}

func legacyDir() string {
	home := homeDir()
	if home == "" {
		return "." + appName
	}
	return filepath.Join(home, "."+appName)
}

func homeDir() string {
	if runtime.GOOS == "windows" {
		// On Windows, use USERPROFILE as fallback.
		if h := os.Getenv("USERPROFILE"); h != "" {
			return h
		}
	}
	h, _ := os.UserHomeDir()
	return h
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
