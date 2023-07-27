#! /usr/bin/env python3
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
