package import_

import "time"

// ImageSettings holds image configuration from HPCM.
type ImageSettings struct {
	Name               string    `json:"name,omitempty"`
	Kernel             string    `json:"kernel,omitempty"`
	CloningBlockDevice string    `json:"cloningBlockDevice,omitempty"`
	CloningDate        time.Time `json:"cloningDate,omitempty"`
}
