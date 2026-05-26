## cani alpha update module

Update a module in the inventory.

### Synopsis

Update a module's fields by UUID or name.

```
cani alpha update module <uuid-or-name> [flags]
```

### Options

```
      --bay string           Module bay name
      --description string   Description
  -h, --help                 help for module
      --name string          New name
      --role string          New role
      --status string        New status
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

