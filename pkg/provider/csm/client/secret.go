package client

import (
	"log"
	"os"
)

// ResolveSecret returns a credential value, preferring the CLI-flag value but
// falling back to the named environment variable when the flag is empty.
//
// Passing secrets as CLI flags exposes them to other local users via process
// listings (ps) and shell history, so a warning is emitted whenever a flag
// value is used. The environment variable is the recommended source.
func ResolveSecret(flagValue, flagName, envName string) string {
	if flagValue != "" {
		log.Printf("WARNING: --%s was passed on the command line; secrets in CLI flags are visible to other local users (ps, shell history). Prefer the %s environment variable.", flagName, envName)
		return flagValue
	}
	return os.Getenv(envName)
}
