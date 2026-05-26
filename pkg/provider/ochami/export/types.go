package export

// openChamiBMC represents the minimal BMC entry expected by OpenCHAMI's nodes.yaml.
type openChamiBMC struct {
	Xname string `yaml:"xname"`
	IP    string `yaml:"ip"`
	MAC   string `yaml:"mac"`
}

// openChamiNode represents the minimal node entry expected by OpenCHAMI's nodes.yaml.
type openChamiNode struct {
	Xname       string   `yaml:"xname"`
	IP          string   `yaml:"ip"`
	BootMAC     string   `yaml:"boot_mac"`
	NID         *int     `yaml:"nid,omitempty"`
	Hostname    string   `yaml:"hostname,omitempty"`
	HostAliases []string `yaml:"host_aliases,omitempty"`
}

// openChamiPayload is the full document layout for OpenCHAMI's nodes.yaml.
type openChamiPayload struct {
	BMCs  []openChamiBMC  `yaml:"bmcs"`
	Nodes []openChamiNode `yaml:"nodes"`
}
