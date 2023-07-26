#! /usr/bin/env bash
#
# MIT License
#
# (C) Copyright 2023 Hewlett Packard Enterprise Development LP
#
# Permission is hereby granted, free of charge, to any person obtaining a
# copy of this software and associated documentation files (the "Software"),
# to deal in the Software without restriction, including without limitation
# the rights to use, copy, modify, merge, publish, distribute, sublicense,
# and/or sell copies of the Software, and to permit persons to whom the
# Software is furnished to do so, subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included
# in all copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
# FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
# THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
# OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
# ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
# OTHER DEALINGS IN THE SOFTWARE.
#
set -eu

curl -X PUT -k -i https://localhost:8443/apis/sls/v1/networks/HMN_MTN -d '{
  "Name": "HMN_MTN",
  "FullName": "Mountain Compute Hardware Management Network",
  "IPRanges": [
    "10.104.0.0/17"
  ],
  "Type": "ethernet",
  "ExtraProperties": {
    "CIDR": "10.104.0.0/17",
    "MTU": 9000,
    "Subnets": [],
    "VlanRange": [
      2000,
      2999
    ]
  }
}'

curl -X PUT -k -i https://localhost:8443/apis/sls/v1/networks/NMN_MTN -d '{
  "Name": "NMN_MTN",
  "FullName": "Mountain Compute Node Management Network",
  "IPRanges": [
    "10.100.0.0/17"
  ],
  "Type": "ethernet",
  "ExtraProperties": {
    "CIDR": "10.100.0.0/17",
    "MTU": 9000,
    "Subnets": [],
    "VlanRange": [
      3000,
      3999
    ]
  }
}'