package devicetypes

import "testing"

// TestGitLoaderPlaceholder ensures the test file is not empty so `go test` recognises it.
func TestGitLoaderPlaceholder(t *testing.T) {
	t.Log("gitloader tests are commented out; enable them when CI has git available")
}

// | Function         | Happy-path test                          | Failure test                              |
// |------------------|------------------------------------------|-------------------------------------------|
// | LoadFromGitRepo  | TestLoadFromGitRepoHappyPath             | TestLoadFromGitRepoInvalidURL             |
// | gitCacheDir      | TestGitCacheDirHappyPath                 | TestGitCacheDirFailure                    |
// | ensureGitRepo    | TestEnsureGitRepoClone                   | TestEnsureGitRepoNoGit                    |
// | requireGit       | TestRequireGitHappyPath                  | TestRequireGitNotFound                    |
// | gitClone         | TestGitCloneHappyPath                    | TestGitCloneInvalidRepo                   |
// | gitPull          | TestGitPullHappyPath                     | TestGitPullNotARepo                       |

// ---------------------------------------------------------------------------
// gitCacheDir
// ---------------------------------------------------------------------------

// func TestGitCacheDirHappyPath(t *testing.T) {
// 	dir, err := gitCacheDir("https://github.com/example/repo.git")
// 	if err != nil {
// 		t.Fatalf("unexpected error: %v", err)
// 	}

// 	home, _ := os.UserHomeDir()
// 	expected := filepath.Join(home, ".cani", "types-cache", sanitizeRepoName("https://github.com/example/repo.git"))
// 	if dir != expected {
// 		t.Errorf("got %q, want %q", dir, expected)
// 	}
// }

// func TestGitCacheDirFailure(t *testing.T) {
// 	// Unset HOME so UserHomeDir fails.
// 	orig := os.Getenv("HOME")
// 	os.Unsetenv("HOME")
// 	defer os.Setenv("HOME", orig)

// 	_, err := gitCacheDir("https://github.com/example/repo.git")
// 	if err == nil {
// 		t.Fatal("expected error when HOME is unset, got nil")
// 	}
// 	if !strings.Contains(err.Error(), "home") {
// 		t.Errorf("error should mention home directory, got: %v", err)
// 	}
// }

// ---------------------------------------------------------------------------
// requireGit
// ---------------------------------------------------------------------------

// Does not run well in CI

// func TestRequireGitHappyPath(t *testing.T) {
// 	// git should be available in the test environment.
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH, skipping happy-path test")
// 	}

// 	if err := requireGit(); err != nil {
// 		t.Fatalf("expected nil error, got: %v", err)
// 	}
// }

// func TestRequireGitNotFound(t *testing.T) {
// 	// Remove git from PATH by setting PATH to an empty directory.
// 	orig := os.Getenv("PATH")
// 	t.Cleanup(func() { os.Setenv("PATH", orig) })

// 	emptyDir := t.TempDir()
// 	os.Setenv("PATH", emptyDir)

// 	err := requireGit()
// 	if err == nil {
// 		t.Fatal("expected error when git is not on PATH, got nil")
// 	}
// 	if !strings.Contains(err.Error(), "git is required") {
// 		t.Errorf("error should mention git requirement, got: %v", err)
// 	}
// }

// ---------------------------------------------------------------------------
// ensureGitRepo
// ---------------------------------------------------------------------------

// func TestEnsureGitRepoClone(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Create a bare local repo to clone from.
// 	bareDir := t.TempDir()
// 	if out, err := exec.Command("git", "init", "--bare", bareDir).CombinedOutput(); err != nil {
// 		t.Fatalf("git init --bare failed: %v\n%s", err, out)
// 	}

// 	cloneDir := filepath.Join(t.TempDir(), "clone")
// 	if err := ensureGitRepo(bareDir, cloneDir, true); err != nil {
// 		t.Fatalf("ensureGitRepo clone failed: %v", err)
// 	}

// 	// .git directory should exist after cloning.
// 	if !dirExists(filepath.Join(cloneDir, ".git")) {
// 		t.Error(".git directory not found after clone")
// 	}
// }

// func TestEnsureGitRepoNoGit(t *testing.T) {
// 	// Remove git from PATH so ensureGitRepo fails at requireGit.
// 	orig := os.Getenv("PATH")
// 	t.Cleanup(func() { os.Setenv("PATH", orig) })

// 	emptyDir := t.TempDir()
// 	os.Setenv("PATH", emptyDir)

// 	err := ensureGitRepo("https://example.com/repo.git", t.TempDir(), false)
// 	if err == nil {
// 		t.Fatal("expected error when git is unavailable, got nil")
// 	}
// 	if !strings.Contains(err.Error(), "git is required") {
// 		t.Errorf("unexpected error message: %v", err)
// 	}
// }

// func TestEnsureGitRepoPullFalseSkipsPull(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Create a bare repo and clone it.
// 	bareDir := t.TempDir()
// 	if out, err := exec.Command("git", "init", "--bare", bareDir).CombinedOutput(); err != nil {
// 		t.Fatalf("git init --bare failed: %v\n%s", err, out)
// 	}

// 	cloneDir := filepath.Join(t.TempDir(), "clone")
// 	if err := ensureGitRepo(bareDir, cloneDir, true); err != nil {
// 		t.Fatalf("initial clone failed: %v", err)
// 	}

// 	// With pull=false the existing clone should be left as-is (no error).
// 	if err := ensureGitRepo(bareDir, cloneDir, false); err != nil {
// 		t.Fatalf("ensureGitRepo with pull=false should succeed, got: %v", err)
// 	}
// }

// ---------------------------------------------------------------------------
// gitClone
// ---------------------------------------------------------------------------

// func TestGitCloneHappyPath(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Set up a local bare repo as the clone source.
// 	bareDir := t.TempDir()
// 	if out, err := exec.Command("git", "init", "--bare", bareDir).CombinedOutput(); err != nil {
// 		t.Fatalf("git init --bare failed: %v\n%s", err, out)
// 	}

// 	target := filepath.Join(t.TempDir(), "nested", "clone")
// 	if err := gitClone(bareDir, target); err != nil {
// 		t.Fatalf("gitClone failed: %v", err)
// 	}

// 	if !dirExists(filepath.Join(target, ".git")) {
// 		t.Error(".git directory not found in cloned repo")
// 	}
// }

// func TestGitCloneInvalidRepo(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	target := filepath.Join(t.TempDir(), "clone")
// 	err := gitClone("file:///nonexistent/repo", target)
// 	if err == nil {
// 		t.Fatal("expected error for invalid repo URL, got nil")
// 	}
// 	if !strings.Contains(err.Error(), "git clone") {
// 		t.Errorf("error should mention git clone, got: %v", err)
// 	}
// }

// ---------------------------------------------------------------------------
// gitPull
// ---------------------------------------------------------------------------

// func TestGitPullHappyPath(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Create a non-bare repo with one commit to serve as the remote.
// 	srcDir := t.TempDir()
// 	for _, args := range [][]string{
// 		{"init", srcDir},
// 		{"-C", srcDir, "commit", "--allow-empty", "-m", "seed"},
// 	} {
// 		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
// 			t.Fatalf("git %v failed: %v\n%s", args, err, out)
// 		}
// 	}

// 	cloneDir := filepath.Join(t.TempDir(), "clone")
// 	if out, err := exec.Command("git", "clone", srcDir, cloneDir).CombinedOutput(); err != nil {
// 		t.Fatalf("git clone failed: %v\n%s", err, out)
// 	}

// 	if err := gitPull(cloneDir); err != nil {
// 		t.Fatalf("gitPull failed: %v", err)
// 	}
// }

// func TestGitPullNotARepo(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Pull inside a directory that is not a git repo.
// 	notRepo := t.TempDir()
// 	err := gitPull(notRepo)
// 	if err == nil {
// 		t.Fatal("expected error for non-repo directory, got nil")
// 	}
// 	if !strings.Contains(err.Error(), "git pull") {
// 		t.Errorf("error should mention git pull, got: %v", err)
// 	}
// }

// ---------------------------------------------------------------------------
// LoadFromGitRepo
// ---------------------------------------------------------------------------

// func TestLoadFromGitRepoHappyPath(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Create a bare repo so the clone step succeeds.
// 	bareDir := t.TempDir()
// 	if out, err := exec.Command("git", "init", "--bare", bareDir).CombinedOutput(); err != nil {
// 		t.Fatalf("git init --bare failed: %v\n%s", err, out)
// 	}

// 	// Override HOME so the cache lands in a temp directory.
// 	origHome := os.Getenv("HOME")
// 	tmpHome := t.TempDir()
// 	os.Setenv("HOME", tmpHome)
// 	t.Cleanup(func() { os.Setenv("HOME", origHome) })

// 	// LoadFromGitRepo will clone then call LoadFromDir on an empty repo.
// 	// LoadFromDir on an empty tree should not error.
// 	err := LoadFromGitRepo(bareDir, true)
// 	if err != nil {
// 		t.Fatalf("LoadFromGitRepo failed: %v", err)
// 	}

// 	// Verify the cache directory was created.
// 	cacheDir := filepath.Join(tmpHome, ".cani", "types-cache", sanitizeRepoName(bareDir))
// 	if !dirExists(cacheDir) {
// 		t.Error("expected cache directory to exist")
// 	}
// }

// func TestLoadFromGitRepoInvalidURL(t *testing.T) {
// 	if _, err := exec.LookPath("git"); err != nil {
// 		t.Skip("git not on PATH")
// 	}

// 	// Override HOME so we don't pollute the real filesystem.
// 	origHome := os.Getenv("HOME")
// 	tmpHome := t.TempDir()
// 	os.Setenv("HOME", tmpHome)
// 	t.Cleanup(func() { os.Setenv("HOME", origHome) })

// 	err := LoadFromGitRepo("file:///nonexistent/repo", false)
// 	if err == nil {
// 		t.Fatal("expected error for invalid repo URL, got nil")
// 	}
// }
