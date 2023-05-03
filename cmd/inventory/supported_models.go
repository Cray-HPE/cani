/*
MIT License

(C) Copyright 2022 Hewlett Packard Enterprise Development LP

Permission is hereby granted, free of charge, to any person obtaining a
copy of this software and associated documentation files (the "Software"),
to deal in the Software without restriction, including without limitation
the rights to use, copy, modify, merge, publish, distribute, sublicense,
and/or sell copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included
in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.
*/
package inventory

func SupportedHardware() []Hardware {
	return []Hardware{
		{
			Model:        "Windom",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "Compute",
		},
		{
			Model:        "Windom",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "Compute",
		},
		{
			Model:        "Grizzly Peak 40 GB",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Grizzly Peak 80 GB with Nvidia A100",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Bard Peak with AMD MI200 and SS11",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Castle",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "DL325",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "NCN",
		},
		{
			Model:        "DL325",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "UAN",
		},
		{
			Model:        "DL-325 Gen10+",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "NCN",
		},
		{
			Model:        "DL-325 Gen10+",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "UAN",
		},
		// v1
		{
			Model:        "DL-385 v1",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "Compute",
		},
		{
			Model:        "DL-385 v1 with Nvidia A100",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "NCN",
		},
		{
			Model:        "DL-385 v1 with Nvidia V100",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "UAN",
		},
		{
			Model:        "DL-385 v1 with Nvidia 6000",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "UAN",
		},
		{
			Model:        "DL-385 v1 with Nvidia MI100",
			Manufacturer: "HPE",
			Alias:        "Rome",
			Type:         "UAN",
		},
		// v2
		{
			Model:        "DL-385 v2 with Nvidia A100",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "NCN",
		},
		{
			Model:        "DL-385 v2 with Nvidia A100",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "UAN",
		},
		{
			Model:        "DL-385 v2 with AMD MI100",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "UAN",
		},
		{
			Model:        "DL-385 v2 with Nvidia A40",
			Manufacturer: "HPE",
			Alias:        "Milan",
		},
		{
			Model:        "DL-385 Gen10+ with AMD MI100",
			Manufacturer: "HPE",
			Alias:        "Milan",
		},
		{
			Model:        "Apollo 2000",
			Manufacturer: "HPE",
			Alias:        "Milan",
			Type:         "Compute",
		},
		{
			Model:        "Apollo 6500 XL675d with Nvidia A40",
			Manufacturer: "HPE",
			Type:         "UAN",
		},
		{
			Model:        "Apollo 6500 XL675d with Nvidia A40",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Apollo 6500 XL675d with Nvidia A100",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Apollo 6500 XL645d w/Nvidia A100",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Apollo 6500 XL645d with AMD MI100",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "Apollo 6500 Gen10+ with AMD MI200",
			Manufacturer: "HPE",
			Type:         "Compute",
		},
		{
			Model:        "DL-360 Gen11+",
			Manufacturer: "HPE",
			Alias:        "Sapphire Rapids",
			Type:         "NCN",
		},
		{
			Model:        "DL-360 Gen11+",
			Manufacturer: "HPE",
			Alias:        "Sapphire Rapids",
			Type:         "UAN",
		},
		{
			Model:        "DL-380 Gen11+",
			Manufacturer: "HPE",
			Alias:        "Sapphire Rapids",
			Type:         "UAN",
		},
		{
			Model:        "R272-Z30-00",
			Manufacturer: "Gigabyte",
			Type:         "NCN",
		},
		{
			Model:        "R272-Z30-00",
			Manufacturer: "Gigabyte",
			Type:         "UAN",
		},
		{
			Model:        "R272-Z30-YF",
			Manufacturer: "Gigabyte",
			Type:         "NCN",
		},
		{
			Model:        "R272-Z30-YF",
			Manufacturer: "Gigabyte",
			Type:         "UAN",
		},
		{
			Model:        "R272-Z30-YF",
			Manufacturer: "Gigabyte",
			Type:         "Compute",
		},
		{
			Model:        " H262-Z61-00",
			Manufacturer: "Gigabyte",
			Type:         "UAN",
		},
		{
			Model:        " H262-Z61-00",
			Manufacturer: "Gigabyte",
			Type:         "Compute",
		},
		{
			Model:        "H262-Z63-YF",
			Manufacturer: "Gigabyte",
			Type:         "Compute",
		},
		{
			Model: "",
			Alias: "ServerTech",
			Type:  "PDU",
		},
		{
			Model:        "CIS P9S23A",
			Manufacturer: "HPE",
			Type:         "PDU",
		},
	}
}
