## cani alpha remove module

Remove a module from the inventory.

### Synopsis

Remove a module by UUID or name.

```
cani alpha remove module <uuid-or-name> [flags]
```

### Options

```
  -h, --help   help for module
```

### Options inherited from parent commands

```
      --config string         config file (default "/Users/jsalmela/.cani/cani.yml")
      --datastore string      datastore type (json, postgres) (default "json")
      --debug                 enable debug mode
  -f, --force                 Remove items without confirmation.
      --strict                require a resolved device type (slug) for all devices (default true)
      --types-dirs strings    local directories with additional hardware types
      --types-repo-pull       pull latest changes from types repos on startup
      --types-repos strings   git repo URLs with additional hardware types
```

### SEE ALSO

* [cani alpha remove](cani_alpha_remove.md)	 - Remove items from the inventory

