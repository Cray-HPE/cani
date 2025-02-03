# Known `cani` Errors And Their Solutions

## `Required network HMN_MTN is missing` or `Required network NMN_MTN is missing`

For the CSM provider, this means the existing SLS data is unstable as the `Networks` key is missing a required field.

To fix the issue, craft an appropriate JSON file and use it to create the network in SLS.

### Create An HMN_MTN or NMN_MTN Network JSON File

Here is an example `hmn_mtn.json` file that creates the network when patched into SLS. 

> Note: update the site-specific information as needed (IP, VLANs, etc.)

```json
{
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
}
```

Here is the same, but for `nmn_mtn.json`.

> Note: update the site-specific information as needed (IP, VLANs, etc.)

```json
{
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
}
```



### Create The Networks in SLS
  
Use the `cray` command to add the network(s):

```shell
# Backup SLS
cray sls dumpstate list --format json > sls_dump_"$(date +"%s")".json 

# Patch in the appropriate network using the JSON files created earlier
cray sls networks update ./hmn_mtn.json HMN_MTN 
cray sls networks update ./nmn_mtn.json NMN_MTN
```

## `Hardware/<XNAME>/ExtraProperties/NID: expected integer, but got string`

This appears when a NID number in the `ExtraProperties.NID` key is populated with a string instead of an interger.

Inspect each NID for any strings.

```shell
cray sls dumpstate list --format json | jq '.Hardware[] | select(.ExtraProperties.NID | type == "string") | .ExtraProperties'
```

The `"2"` and `"3:` in the output below cause the issue in this example.

```json
{
  "Aliases": [
    "nid000002"
  ],
  "NID": "2",
  "Role": "Compute"
}
{
  "Aliases": [
    "nid000003"
  ],
  "NID": "3",
  "Role": "Compute"
}
```

### Convert The Strings To Integers In SLS Using `jq`

Get the affected xnames into an array for looping through.

```shell
# via cray
xnames=($(cray sls dumpstate list --format json  | jq -r '[.Hardware[] | select(.ExtraProperties.NID | type == "string") | .Xname] | join(" ")'))

# or via curl
xnames=($(curl https://api-gw-service-nmn.local/apis/sls/v1/dumpstate | jq -r '[.Hardware[] | select(.ExtraProperties.NID | type == "string") | .Xname] | join(" ")'))
```

Dump each hardware component into a file and edit their values back to integers.

```shell
cray sls dumpstate list --format json | jq '.Hardware[] | select(.ExtraProperties.NID | type == "string") | .ExtraProperties.NID |= tonumber' > fixed.json
```

Loop through each fixed entry and update it in SLS.

```shell
for xname in "${xnames[@]}"
do
  # via cray
  jq --arg xname "${xname}" '. | select(.Xname == $xname)' < fixed.json > "${xname}.json"
  cray sls hardware update "${xname}" --payload-file "${xname}.json"
  
  # or via curl
  curl -X 'PUT' https://api-gw-service-nmn.local/apis/sls/v1/hardware/"${xname}" -H "'Authorization: Bearer ${TOKEN}'" -H 'Accept: application/json' -H 'Content-Type: application/json' -d "$(jq -c -n --arg xname "${xname}" 'inputs | select(.Xname == $xname) | .' < fixed.json)"
done
```

