# ETL Pipeline

Each provider follows an Extract-Transform-Load (ETL) pattern when importing data. This pipeline converts provider-specific data into the portable inventory format.

## Pipeline Stages

### Extract

The provider connects to the external source and pulls raw data. This might be a REST API call (Nautobot, CSM), a JSON file (Redfish, Ochami), or a flat file (Example).

### Transform

Raw data is converted to the inventory data model. Device types are matched against the device-types library, and parent-child relationships are established.

### Load

The transformed inventory is saved to the local datastore. Subsequent commands (`add`, `remove`, `update`, `classify`, `show`) work against this local copy.

## Example

```shell
# 1. Import devices from Redfish
cani alpha import redfish --root ./redfish-roots.json

# 2. Classify unrecognized devices
cani alpha classify --auto

# 3. Add racks if they were not included in the import
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack

# 4. Resolve orphaned devices (those without a parent rack)
cani alpha update orphans

# 5. Export to Nautobot
cani alpha export nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN
```

## Injected Metadata

Providers may inject metadata into the inventory during import. For example, the CSM provider stores schema versioning:

```json
"ExtraProperties": {
    "@cani.id": "ebb15e9a-f418-408c-8315-601483f3b279",
    "@cani.lastModified": "2023-07-03 18:24:29.12421 +0000 UTC",
    "@cani.slsSchemaVersion": "v1alpha1"
}
```

This metadata enables `cani` to track provenance and detect schema changes across imports.
