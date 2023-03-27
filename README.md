<p align="center">
  <h1>csminv</h1>
  <strong>From subfloor to top-of-rack, manage your HPC cluster's inventory!</strong>
</p>

# Quick Start


## Extract The Current Inventory

> Users must be able to extract the inventory on a running CSM machine to a portable format

```shell
csminv extract sls_input_file.json system_config.yaml paddle.json [OTHER FILES]...
```

This will return JSON with a key that contains the extracted data (the "old") as well as a "new" key for the new inventory format.

```json
{
 "Extract": {
  "CanuConfig": {},
  "CsiConfig": {},
  "SlsConfig": {},
 },
 "Inventory": { <-- keys to be determined as the inventory evolves
  "Cabinets": [],
  "Switches": [],
 },
}
```

### Transform The Existing Data

The `NewCsmInventory` will use the extracted information and transform it into the new data structure.  This allows for a transition off of the old and into the new without having to change this application much.

## Define the Schema

From here, we can define the new schema and then enforce it.  This can happen automatically as the structure is defined and commented in the code:

```shell
csminv schema show
```

```json
{
 "$schema": "https://json-schema.org/draft/2020-12/schema",
 "$id": "https://github.com/Cray-HPE/csminv/csminv/inventory",
 "$ref": "#/$defs/Inventory",
 "$defs": {
  "CanuConfig": {
   "properties": {
    "canu_version": {
     "type": "string",
     "description": "Version of canu used to generate the paddle file"
    },
    ...
    ...
```

The schema can be defined in code.  See the `CanuConfig.properties.description` maps to this:

```go
type CanuConfig struct {
	// Version of canu used to generate the paddle file
	CanuVersion string `json:"canu_version" env:"CANU_CANU_VERSION" default:"" flag:"canu-version" usage:"Version of canu" jsonschema:"required"`
```

## Adding/removing hardware

Once the data is transformed to the new format, we have both the existing data and the new, so we can add or remove hardware using the existing procedures, while maintaining the new inventory at the same time.

```shell
csminv list
csminv add switch [FLAGS]...
csminv remove cabinet [FLAGS]...
```

# Tests

Install [shellspec](https://shellspec.info) and run `make test`.

This builds the binary and then runs it under several scenarios to determine that the correct output is seen.

## Writing Tests

If you add a new command, create a new `spec/something_spec.sh` that follows the format of the other files.  Each flag and different output should be accounted for in the tests.  If you need to test for something, but it is not ready yet, add a Todo-style test:

```shell
It 'is a Todo'
End
```

This will let it show up as a tests in `# TODO` format, while still allowing the suite to pass.

```
not ok 54 - is a Todo # TODO Not yet implemented
```
