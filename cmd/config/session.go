package config

import "github.com/Cray-HPE/cani/internal/domain"

type Session struct {
	DomainOptions *domain.NewOpts `yaml:"domain_options"`
	Active        bool            `yaml:"active"`
}
