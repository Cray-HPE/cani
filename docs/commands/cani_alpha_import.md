## cani alpha import

Import assets into the inventory

### Synopsis

Import assets into the inventory from an external source using a provider.

```
cani alpha import PROVIDER [flags]
```

### Options

```
  -h, --help           help for import
      --no-color       Disable colorized output
      --phase string   ETL phases to run: e (extract), et (+transform), etl (+load) (default "etl")
      --step           Step through each item interactively (implies --debug)
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
* [cani alpha import csm](cani_alpha_import_csm.md)	 - Import assets using the csm provider
* [cani alpha import example](cani_alpha_import_example.md)	 - Import assets using the example provider
* [cani alpha import nautobot](cani_alpha_import_nautobot.md)	 - Import assets using the nautobot provider
* [cani alpha import ochami](cani_alpha_import_ochami.md)	 - Import assets using the ochami provider
* [cani alpha import redfish](cani_alpha_import_redfish.md)	 - Import assets using the redfish provider

