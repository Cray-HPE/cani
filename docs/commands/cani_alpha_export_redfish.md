## cani alpha export redfish

Export inventory using the redfish provider

```
cani alpha export redfish [flags]
```

### Options

```
  -h, --help   help for redfish
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --dry-run               Preview changes without making API calls
      --merge                 Update existing devices instead of skipping conflicts
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha export](cani_alpha_export.md)	 - Export inventory to an external provider

