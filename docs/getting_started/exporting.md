# Exporting Data

After adding, removing, updating, or classifying hardware in the local inventory, the `export` command pushes the changes to an external provider.

## Export To A Provider

```shell
# Export to a Nautobot instance
cani alpha export nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN

# Export to CSM SLS
cani alpha export csm \
  --csm-keycloak-username username \
  --csm-keycloak-password password \
  --csm-api-host api-gw-service-nmn.local
```

## Dry Run

Preview the changes without making any API calls:

```shell
# See what would be pushed without actually pushing
cani alpha export nautobot --dry-run
```

## Merge Mode

By default, conflicts are skipped. Use `--merge` to combine with existing data:

```shell
# Merge with existing devices instead of skipping conflicts
cani alpha export nautobot --merge
```
