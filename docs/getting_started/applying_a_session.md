# Applying A Session

After adding, removing, and/or updating any amount of hardware to the inventory, applying a session runs some validations and prompts for confirmation before commiting the changes.

## Apply A Session

Apply the session, which commits any changes to the external inventory provider.

```shell
# example after running some 'cani add', 'cani remove', 'cani update', subcommands
cani alpha session apply # runs validation and if successful, prompts to commit the changes
# example commit changes without prompting
cani alpha session apply --commit
```
