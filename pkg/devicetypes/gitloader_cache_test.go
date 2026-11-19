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
package devicetypes

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoadFromGitRepoWithoutCloneLeavesNoCacheDir verifies that loading a
// types repo with cloning disabled and no existing checkout creates nothing.
//
// Why it matters: the loader used to write an empty types cache into the
// repo's cache directory in that case, and a later run with cloning enabled
// then failed with "git clone ... exit status 128" because git refuses a
// non-empty target directory.
func TestLoadFromGitRepoWithoutCloneLeavesNoCacheDir(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	repo := "https://example.invalid/acme/library.git"

	if err := LoadFromGitRepo(repo, false, false); err != nil {
		t.Fatalf("LoadFromGitRepo with clone disabled: %v", err)
	}
	cacheDir, err := gitCacheDir(repo)
	if err != nil {
		t.Fatal(err)
	}
	if pathExists(cacheDir) {
		t.Errorf("cache dir %s was created although nothing was cloned", cacheDir)
	}
}

// TestEnsureGitRepoClonesOverLeftoverCacheFile verifies that a cache
// directory holding only the types cache file written by earlier releases is
// replaced by a clone instead of making the clone fail.
func TestEnsureGitRepoClonesOverLeftoverCacheFile(t *testing.T) {
	bare, _ := libraryRepo(t)
	cacheDir := filepath.Join(t.TempDir(), "library")
	writeLibraryFile(t, cacheDir, cacheFileName, `{"version":1}`)

	if err := ensureGitRepo(bare, cacheDir, true, false); err != nil {
		t.Fatalf("ensureGitRepo over a leftover cache file: %v", err)
	}
	if !pathExists(filepath.Join(cacheDir, "device-types", "acme", "server.yaml")) {
		t.Error("clone missing device-types/acme/server.yaml")
	}
}

// TestEnsureGitRepoKeepsForeignCacheDir verifies that a cache directory with
// content other than the types cache file is neither deleted nor cloned into,
// and that the error names the directory.
func TestEnsureGitRepoKeepsForeignCacheDir(t *testing.T) {
	bare, _ := libraryRepo(t)
	cacheDir := filepath.Join(t.TempDir(), "library")
	writeLibraryFile(t, cacheDir, "notes.txt", "mine")

	err := ensureGitRepo(bare, cacheDir, true, false)
	if err == nil || !strings.Contains(err.Error(), cacheDir) {
		t.Fatalf("expected an error naming %s, got %v", cacheDir, err)
	}
	if data, readErr := os.ReadFile(filepath.Join(cacheDir, "notes.txt")); readErr != nil || string(data) != "mine" {
		t.Errorf("foreign file was modified or removed: %q, %v", data, readErr)
	}
}
