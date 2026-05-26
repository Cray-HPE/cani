## cani alpha import nautobot

Import assets using the nautobot provider

```
cani alpha import nautobot [flags]
```

### Options

```
      --default-location string   Default location for imported devices
      --default-role string       Default role for imported devices
      --default-status string     Default status for imported devices (default "Active")
  -h, --help                      help for nautobot
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --no-color              Disable colorized output
      --phase string          ETL phases to run: e (extract), et (+transform), etl (+load) (default "etl")
      --step                  Step through each item interactively (implies --debug)
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha import](cani_alpha_import.md)	 - Import assets into the inventory

