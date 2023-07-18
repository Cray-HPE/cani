#! /usr/bin/env python3

# TODO I don't like the name of this file

import requests
import sys
import json

print("Clearing BSS")

result = requests.get("http://localhost:27778/boot/v1/bootparameters")
if result.status_code != 200:
    print(f"Unexpected status code: {result.status_code}, expected 200")
    sys.exit(1)

if result.json() is  None:
    print("No bootparams to remove")
else:
    bootparams = result.json()
    for i, bootparam in enumerate(bootparams, start=1):
        result = requests.delete("http://localhost:27778/boot/v1/bootparameters", json=bootparam)
        print(f"Deleted bootparameter {i} of {len(bootparams)}: {result.status_code}")

print("Loading BSS")
with open(sys.argv[1], 'r') as f:
    bootparams = json.load(f)
    for i, bootparam in enumerate(bootparams, start=1):
        result = requests.post("http://localhost:27778/boot/v1/bootparameters", json=bootparam)
        print(f"Created bootparameter {i} of {len(bootparams)}: {result.status_code}")
