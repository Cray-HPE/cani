package config

import "github.com/Cray-HPE/cani/internal/plugin"

// Session defines the session configuration and domain options
type Session struct {
	DomainOptions *plugin.NewOpts `yaml:"domain_options"`
	Domain        *plugin.Plugin  `yaml:"domain"`
	Active        bool            `yaml:"active"`
}
