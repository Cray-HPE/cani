## cani alpha update orphans

Interactively assign parents to orphaned items.

### Synopsis

Walk all orphaned racks and devices and interactively prompt
for a parent assignment. Candidates are ranked by name similarity,
hardware-type affinity, and provider metadata.

In --dry-run mode the assignments are written to a JSON plan file
that can be reviewed, edited, and applied later with --apply-plan.

```
cani alpha update orphans [flags]
```

### Options

```
      --apply-plan string   Apply a previously saved plan file instead of prompting
      --dry-run             Preview changes and save to a plan file without modifying inventory
  -h, --help                help for orphans
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

