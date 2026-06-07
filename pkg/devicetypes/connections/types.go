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
package connections

// ConnectionMap is the top-level structure of a connection map YAML file.
// It declaratively describes desired cable connections between devices
// using human-readable names and brace-expansion patterns.
type ConnectionMap struct {
	Version       string            `yaml:"version" json:"version"`
	CableDefaults *CableDefaults    `yaml:"cable_defaults,omitempty" json:"cable_defaults,omitempty"`
	Connections   []ConnectionEntry `yaml:"connections" json:"connections"`
}

// CableDefaults provides default cable properties applied to every
// connection that does not override them.
type CableDefaults struct {
	Type       string `yaml:"type,omitempty" json:"type,omitempty"`
	Status     string `yaml:"status,omitempty" json:"status,omitempty"`
	Color      string `yaml:"color,omitempty" json:"color,omitempty"`
	LengthUnit string `yaml:"length_unit,omitempty" json:"length_unit,omitempty"`
}

// ConnectionEntry defines one or more cables between device/port pairs.
// Device and port fields support brace-expansion patterns (e.g.
// "node-{01..48}") for bulk connections with zip semantics.
type ConnectionEntry struct {
	A     Endpoint    `yaml:"a" json:"a"`
	B     Endpoint    `yaml:"b" json:"b"`
	Cable *CableProps `yaml:"cable,omitempty" json:"cable,omitempty"`
}

// Endpoint identifies one side of a cable connection.
type Endpoint struct {
	Device string `yaml:"device" json:"device"`
	Port   string `yaml:"port" json:"port"`
	Mac    string `yaml:"mac,omitempty" json:"mac,omitempty"`
}

// CableProps holds per-connection cable properties that override defaults.
type CableProps struct {
	Type       string   `yaml:"type,omitempty" json:"type,omitempty"`
	Label      string   `yaml:"label,omitempty" json:"label,omitempty"`
	Color      string   `yaml:"color,omitempty" json:"color,omitempty"`
	Length     *float64 `yaml:"length,omitempty" json:"length,omitempty"`
	LengthUnit string   `yaml:"length_unit,omitempty" json:"length_unit,omitempty"`
	Status     string   `yaml:"status,omitempty" json:"status,omitempty"`
}
