## cani alpha update location

Update a location in the inventory.

### Synopsis

Update a location's fields by UUID or name.

```
cani alpha update location <uuid-or-name> [flags]
```

### Options

```
      --address string       Physical address
      --description string   Description
      --facility string      Facility name
  -h, --help                 help for location
      --name string          New name
      --status string        New status
      --type string          Location type (site, building, floor, room)
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
      --set stringArray       Set field value as key=value (repeatable)
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha update](cani_alpha_update.md)	 - Update items in the inventory.

