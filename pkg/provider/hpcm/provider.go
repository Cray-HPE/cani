package hpcm

import import_ "github.com/Cray-HPE/cani/pkg/provider/hpcm/import"

// Hpcm implements the provider.Provider interface
type Hpcm struct {
	Options  *ImportOptions
	RawNodes []import_.Node
}

// New creates a new Hpcm provider instance
func New() *Hpcm {
	return &Hpcm{}
}

// ClearNodes resets the raw node storage for a fresh import.
func (p *Hpcm) ClearNodes() {
	p.RawNodes = nil
}

// SetNodes stores raw nodes from the import phase.
func (p *Hpcm) SetNodes(nodes []import_.Node) {
	p.RawNodes = nodes
}

// GetNodes returns the raw nodes for the transform phase.
func (p *Hpcm) GetNodes() []import_.Node {
	return p.RawNodes
}

func (p *Hpcm) Slug() string {
	return "hpcm"
}
