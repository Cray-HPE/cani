#! /usr/bin/env bash
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