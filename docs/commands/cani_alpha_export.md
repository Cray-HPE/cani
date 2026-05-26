## cani alpha export

Export inventory to an external provider

### Synopsis

Export the CANI inventory to an external provider using a provider.

```
cani alpha export PROVIDER [flags]
```

### Options

```
      --dry-run   Preview changes without making API calls
  -h, --help      help for export
      --merge     Update existing devices instead of skipping conflicts
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

* [cani alpha](cani_alpha.md)	 - Run commands that are considered unstable.
* [cani alpha export csm](cani_alpha_export_csm.md)	 - Export inventory using the csm provider
* [cani alpha export example](cani_alpha_export_example.md)	 - Export inventory using the example provider
* [cani alpha export nautobot](cani_alpha_export_nautobot.md)	 - Export inventory using the nautobot provider
* [cani alpha export ochami](cani_alpha_export_ochami.md)	 - Export inventory using the ochami provider
* [cani alpha export redfish](cani_alpha_export_redfish.md)	 - Export inventory using the redfish provider

