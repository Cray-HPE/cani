# Adding Hardware

Systems change over time and adding hardware is often needed to expand the available resources.  Equally as common is adding hardware to replace a faulty unit.  Once a session has started, adding different types of hardware is possible.

## Add Hardware

One common use-case is adding a new cabinet to the system.

```shell
# example: adding a supported cabinet
# list available types
cani alpha add cabinet --list-supported-types
# add a ex2000 cabinet, accepting recommended values
cani alpha add cabinet hpe-ex2000 --auto --accept 
```

### List Hardware

The recent addition can be viewed with a `list` subcommand.

```shell
# example: show all cabinets
cani alpha list cabinet
```
