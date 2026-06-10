package devicetypes

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// LoadFromGitRepo clones or pulls a git repository into a local cache
// directory, then loads hardware types from it. The cache lives under
// ~/.cani/types-cache/<sanitized-repo-name>/.
// When pull is false the repo is cloned on first use but never updated.
func LoadFromGitRepo(repoURL string, clone, pull bool) error {
	if err := validateRepoURL(repoURL); err != nil {
		return err
	}

	cacheDir, err := gitCacheDir(repoURL)
	if err != nil {
		return fmt.Errorf("resolving cache dir: %w", err)
	}

	if err := ensureGitRepo(repoURL, cacheDir, clone, pull); err != nil {
		return err
	}

	source := "git:" + repoURL
	return LoadFromDir(cacheDir, source)
}

// gitCacheDir returns the local cache path for a given repo URL.
func gitCacheDir(repoURL string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	name := sanitizeRepoName(repoURL)
	return filepath.Join(home, ".cani", "types-cache", name), nil
}

// ensureGitRepo clones the repository if it doesn't exist locally,
// or pulls the latest changes if it does and pull is true.
func ensureGitRepo(repoURL, cacheDir string, clone, pull bool) error {
	if err := requireGit(); err != nil {
		return err
	}

	gitDir := filepath.Join(cacheDir, ".git")
	if dirExists(gitDir) {
		if !pull {
			if Debug {
				log.Printf("Skipping pull for %s (types_repo_pull is false)", repoURL)
			}
			return nil
		}
		return gitPull(cacheDir)
	}

	if !clone {
		if Debug {
			log.Printf("Skipping clone for %s (types_repo_clone is false)", repoURL)
		}
		return nil
	}
	return gitClone(repoURL, cacheDir)
}

// requireGit checks that the git binary is on PATH.
func requireGit() error {
	_, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("git is required for --types-repo but not found on PATH")
	}
	return nil
}

// validateRepoURL rejects types-repo URLs that could be abused for argument
// injection or arbitrary command execution. Git transports such as "ext::"
// and "file::" can run shell commands, and a leading dash would be parsed by
// git as a command-line option, so only an explicit allowlist is accepted.
func validateRepoURL(repoURL string) error {
	if repoURL == "" {
		return fmt.Errorf("types repo URL is empty")
	}
	if strings.HasPrefix(repoURL, "-") {
		return fmt.Errorf("invalid types repo URL %q: must not start with '-'", repoURL)
	}
	// scp-style shorthand, e.g. git@github.com:org/repo.git
	if isSCPSyntax(repoURL) {
		return nil
	}
	u, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid types repo URL %q: %w", repoURL, err)
	}
	switch u.Scheme {
	case "https", "http", "ssh", "git":
		return nil
	}
	return fmt.Errorf("unsupported types repo URL scheme %q (allowed: https, http, ssh, git, or scp-style user@host:path)", u.Scheme)
}

// isSCPSyntax reports whether s looks like git's scp-style SSH shorthand,
// e.g. "git@github.com:org/repo.git", where the ':' appears before any '/'.
func isSCPSyntax(s string) bool {
	at := strings.Index(s, "@")
	colon := strings.Index(s, ":")
	if at <= 0 || colon < at {
		return false
	}
	slash := strings.Index(s, "/")
	return slash == -1 || colon < slash
}

// gitClone clones a repository into the target directory.
func gitClone(repoURL, target string) error {
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return fmt.Errorf("creating cache parent dir: %w", err)
	}

	if Debug {
		log.Printf("Cloning types repo %s into %s", repoURL, target)
	}

	cmd := exec.Command("git", "clone", "--depth", "1", "--", repoURL, target)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git clone %s: %w", repoURL, err)
	}
	return nil
}

// gitPull fetches and merges the latest changes in an existing clone.
func gitPull(repoDir string) error {
	// Invalidate the types cache so it is rebuilt after the pull.
	removeDirCache(repoDir)

	if Debug {
		log.Printf("Pulling latest types from %s", repoDir)
	}

	cmd := exec.Command("git", "-C", repoDir, "pull", "--ff-only")
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull in %s: %w", repoDir, err)
	}
	return nil
}
