# Importing Data

The `import` command pulls inventory data from an external provider into the local datastore. Each provider has its own flags for authentication and connection details.

## Import From A Provider

```shell
# Import from a Nautobot instance
cani alpha import nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN

# Import from a Redfish ServiceRoot JSON file
cani alpha import redfish --root ./redfish-roots.json

# Import from an Ochami export
cani alpha import ochami --source ./ochami-export.json

# Import from a CSV or YAML file using the example provider
cani alpha import example --source ./inventory.csv
```

## CSM Provider

The CSM provider connects to SLS. Credentials and the API host are required.

```shell
# Import from CSM over the CMN
cani alpha import csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host cmn.example.com

# Import from CSM on an NCN
cani alpha import csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local
```

## What Happens During Import

Import runs the provider's ETL pipeline:

1. **Extract** — Pull raw data from the external source
2. **Transform** — Convert to the portable inventory format
3. **Load** — Save the result to the local datastore

Once imported, the inventory can be manipulated with `add`, `remove`, `update`, `classify`, and `show` commands before exporting back.
