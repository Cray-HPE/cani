package export

// openChamiEntry represents the BMC and node entry shape expected by
// OpenCHAMI's ex-bootstrap inventory FileFormat.
type openChamiEntry struct {
	Xname string `yaml:"xname"`
	MAC   string `yaml:"mac"`
	IP    string `yaml:"ip"`
}

// openChamiPayload is the full document layout for OpenCHAMI's nodes.yaml.
type openChamiPayload struct {
	BMCs  []openChamiEntry `yaml:"bmcs"`
	Nodes []openChamiEntry `yaml:"nodes"`
}
