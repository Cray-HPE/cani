# Stopping A Session

After adding, removing, and/or updating any amount of hardware to the inventory, stopping a sessions runs some validations and prompts for confirmation before commiting the changes.

## Stop A Session

Stop the session and commit any changes to the external inventory provider.

```shell
# example after running some 'cani add', 'cani remove', 'cani update', subcommands
cani alpha session stop # runs validation and if successful, prompts to commit the changes
# example commit changes without prompting
cani alpha session stop --commit
```
