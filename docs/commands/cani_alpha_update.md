## cani alpha update

Update items in the inventory.

### Synopsis

Update items in the inventory.

```
cani alpha update [flags]
```

### Options

```
  -h, --help              help for update
      --set stringArray   Set field value as key=value (repeatable)
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
* [cani alpha update cable](cani_alpha_update_cable.md)	 - Update a cable in the inventory.
* [cani alpha update device](cani_alpha_update_device.md)	 - Update a device in the inventory.
* [cani alpha update location](cani_alpha_update_location.md)	 - Update a location in the inventory.
* [cani alpha update module](cani_alpha_update_module.md)	 - Update a module in the inventory.
* [cani alpha update orphans](cani_alpha_update_orphans.md)	 - Interactively assign parents to orphaned items.
* [cani alpha update rack](cani_alpha_update_rack.md)	 - Update a rack in the inventory.

