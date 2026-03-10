## cani alpha export nautobot

Export inventory using the nautobot provider

```
cani alpha export nautobot [flags]
```

### Options

```
      --create-device-types   Create missing device types in Nautobot (default true)
      --create-locations      Create missing locations in Nautobot (default true)
      --create-roles          Create missing roles in Nautobot (default true)
      --create-statuses       Create missing statuses in Nautobot (default true)
      --dry-run               Log planned actions without making API calls
  -h, --help                  help for nautobot
      --merge                 Merge with existing devices instead of skipping conflicts
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha export](cani_alpha_export.md)	 - Export inventory to an external provider

