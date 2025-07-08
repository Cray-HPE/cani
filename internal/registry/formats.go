package registry

import (
	"fmt"
	"sync"

	"github.com/spf13/cobra"
)

// FormatHandler defines how a provider handles a specific format
type FormatHandler func(data interface{}, cmd *cobra.Command, args []string) error

// FormatInfo contains metadata about a format
type FormatInfo struct {
	Name        string
	Description string
	Provider    string
	Handler     FormatHandler
}

// FormatRegistry manages format registration across providers
type FormatRegistry struct {
	mu      sync.RWMutex
	formats map[string]map[string]FormatInfo // command -> format -> info
}

var globalRegistry = &FormatRegistry{
	formats: make(map[string]map[string]FormatInfo),
}

// RegisterFormat registers a new format for a specific command
func RegisterFormat(command, format, provider, description string, handler FormatHandler) error {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	if globalRegistry.formats[command] == nil {
		globalRegistry.formats[command] = make(map[string]FormatInfo)
	}

	if _, exists := globalRegistry.formats[command][format]; exists {
		return fmt.Errorf("format '%s' already registered for command '%s'", format, command)
	}

	globalRegistry.formats[command][format] = FormatInfo{
		Name:        format,
		Description: description,
		Provider:    provider,
		Handler:     handler,
	}

	return nil
}

// GetValidFormats returns all valid formats for a command
func GetValidFormats(command string) []string {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	var formats []string
	if cmdFormats, exists := globalRegistry.formats[command]; exists {
		for format := range cmdFormats {
			formats = append(formats, format)
		}
	}
	return formats
}

// GetFormatInfo returns information about a specific format
func GetFormatInfo(command, format string) (FormatInfo, bool) {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	if cmdFormats, exists := globalRegistry.formats[command]; exists {
		if info, exists := cmdFormats[format]; exists {
			return info, true
		}
	}
	return FormatInfo{}, false
}

// HandleFormat executes the handler for a specific format
func HandleFormat(command, format string, data interface{}, cmd *cobra.Command, args []string) error {
	info, exists := GetFormatInfo(command, format)
	if !exists {
		return fmt.Errorf("format '%s' not supported for command '%s'", format, command)
	}
	return info.Handler(data, cmd, args)
}

// ListFormats returns all formats with their metadata for a command
func ListFormats(command string) map[string]FormatInfo {
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	result := make(map[string]FormatInfo)
	if cmdFormats, exists := globalRegistry.formats[command]; exists {
		for format, info := range cmdFormats {
			result[format] = info
		}
	}
	return result
}
