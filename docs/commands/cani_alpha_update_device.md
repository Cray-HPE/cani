## cani alpha update device

Update a device in the inventory.

### Synopsis

Update a device's fields by UUID or name.

```
cani alpha update device <uuid-or-name> [flags]
```

### Options

```
      --alias string         Node alias
      --description string   Description
      --face string          Rack face (front, rear)
  -h, --help                 help for device
      --name string          New name
      --nid int              Node ID
      --parent string        Parent UUID or name (rack or device)
      --position int         Rack U position
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

