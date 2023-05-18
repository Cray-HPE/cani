package config

import "github.com/Cray-HPE/cani/internal/domain"

// Session defines the session configuration and domain options
type Session struct {
	DomainOptions *domain.NewOpts `yaml:"domain_options"`
	Domain        *domain.Domain  `yaml:"domain"`
	Active        bool            `yaml:"active"`
}
