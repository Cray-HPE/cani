# Validating

In many situations, it is helpful for the inventory provider to validate any changes and for the user to see them.

## Validate An Inventory

Running the `validate` subcommand will validate `cani`'s proposed changes with the external inventory provider.  If something is missing, the required information is shown for the user to take action on, often through the `update` subcommand.

```shell
# example, validate the existing inventory
cani alpha validate
```
