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

package config

import (
	"os"
	"strconv"
	"strings"
)

// This file provides the standard-library replacement for the small slice of
// viper the codebase used: environment-variable overrides layered on top of the
// loaded config file, with the precedence env var > config file > zero value.
// Environment variables use the historical viper mapping: the "CANI_" prefix
// followed by the upper-cased, underscore-joined key path
// (e.g. provider "nautobot", key "url" -> CANI_NAUTOBOT_URL).

const envPrefix = "CANI"

// EnvKey builds the CANI_-prefixed environment variable name for a provider
// config key path.
func EnvKey(providerName string, keys ...string) string {
	parts := append([]string{envPrefix, providerName}, keys...)
	joined := strings.Join(parts, "_")
	return strings.ToUpper(strings.NewReplacer("-", "_", ".", "_").Replace(joined))
}

// LookupEnv returns the environment override for a provider config key path,
// reporting whether it was set.
func LookupEnv(providerName string, keys ...string) (string, bool) {
	return os.LookupEnv(EnvKey(providerName, keys...))
}

// LookupString resolves a provider config string with precedence
// env var > config file > "".
func LookupString(providerName string, keys ...string) string {
	if v, ok := LookupEnv(providerName, keys...); ok {
		return v
	}
	return GetNestedString(providerName, "", keys...)
}

// LookupBool resolves a provider config boolean with precedence
// env var > config file > false.
func LookupBool(providerName string, keys ...string) bool {
	if v, ok := LookupEnv(providerName, keys...); ok {
		b, err := strconv.ParseBool(v)
		return err == nil && b
	}
	if v, ok := GetNestedValue(providerName, keys...); ok {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}
