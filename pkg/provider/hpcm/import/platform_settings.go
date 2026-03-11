package import_

// PlatformSettings holds platform configuration from HPCM.
type PlatformSettings struct {
	Name            string `json:"name,omitempty"`
	Architecture    string `json:"architecture,omitempty"`
	SerialPort      string `json:"serialPort,omitempty"`
	SerialPortSpeed string `json:"serialPortSpeed,omitempty"`
	VendorsArgs     string `json:"vendorsArgs,omitempty"`
}
