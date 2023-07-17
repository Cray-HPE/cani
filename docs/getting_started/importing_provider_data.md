# Import Inventory Data From An External Provider

A common starting point when working with inventory data is to import it into `cani`, manipulating it, and pushing the updated data back.

## Import From A Provider

In this example, a session is started with the CSM provider.  Data is imported from SLS.  

> Note: This example uses -k (insecure), which is not recommended, but is available for this alpha release

```shell
# example: starting a session 
cani alpha session start csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local

# importing the data and inject cani metadata, including schema version to the provider
cani alpha session import
```

### Injected Metadata

For the CSM provider, `cani` metadata is injected when the session is stopped.

```json
"ExtraProperties": {
        "@cani.id": "ebb15e9a-f418-408c-8315-601483f3b279",
        "@cani.lastModified": "2023-07-03 18:24:29.12421 +0000 UTC",
        "@cani.slsSchemaVersion": "v1alpha1",
}
```

Since SLS can accept any arbitrary data, this allows `cani` to check if the data is compliant with a specific version of the schema.
