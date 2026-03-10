# Known Errors

This page lists known errors organized by provider. Select a provider below to
see its specific errors and solutions.

## By Provider

| Provider | Description |
| -------- | ----------- |
| [CSM](csm/known_errors.md) | SLS network and xname issues |
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

