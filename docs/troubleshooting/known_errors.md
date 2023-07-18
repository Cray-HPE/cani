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
