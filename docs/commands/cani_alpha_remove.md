## cani alpha remove

Remove items from the inventory

### Synopsis

Remove items from the inventory.

```
cani alpha remove [flags]
```

### Options

```
  -f, --force   Remove items without confirmation.
  -h, --help    help for remove
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
* [cani alpha remove cable](cani_alpha_remove_cable.md)	 - Remove a cable from the inventory.
* [cani alpha remove device](cani_alpha_remove_device.md)	 - Remove a device from the inventory.
* [cani alpha remove location](cani_alpha_remove_location.md)	 - Remove a location from the inventory.
* [cani alpha remove module](cani_alpha_remove_module.md)	 - Remove a module from the inventory.
* [cani alpha remove rack](cani_alpha_remove_rack.md)	 - Remove a rack from the inventory.

