## cani alpha export

Export assets from the inventory.

### Synopsis

Export assets from the inventory.

```
cani alpha export [flags]
```

### Options

```
  -a, --all              List all components. This overrides the --type option
      --headers string   Comma separated list of fields to get (default "Type,Vlan,Role,SubRole,Nid,Alias,Name,ID,Location")
  -h, --help             help for export
  -t, --type string      Comma separated list of the types of components to output (default "Node,Cabinet")
```

### Options inherited from parent commands

```
      --config string   Path to the configuration file
  -D, --debug           additional debug output
  -v, --verbose         additional verbose output
```

### SEE ALSO

* [cani alpha](cani_alpha.md)	 - Run commands that are considered unstable.

###### Auto generated by spf13/cobra on 7-Aug-2023
