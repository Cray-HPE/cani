## cani alpha classify

Classify unclassified devices in the inventory

### Synopsis

Scan the local inventory for devices that have no device type slug or
model and interactively assign a type from the hardware library.

Examples:
  # Interactively classify all unclassified devices
  cani alpha classify

  # Auto-accept suggestions with score >= 90
  cani alpha classify --auto

  # Only classify devices matching a name pattern
  cani alpha classify --filter "ncn-.*"

```
cani alpha classify [flags]
```

### Options

```
      --auto             Auto-accept top suggestion if score >= threshold
      --auto-score int   Minimum score for --auto acceptance (0-100) (default 90)
      --filter string    Regex filter on device name
  -h, --help             help for classify
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

