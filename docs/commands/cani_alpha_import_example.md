## cani alpha import example

Import assets using the example provider

```
cani alpha import example [flags]
```

### Options

```
  -c, --csv string    CSV file to import (columns: PartNumber, Description, Quantity, [ConfigGroup])
  -f, --file string   YAML inventory file to import
  -h, --help          help for example
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

