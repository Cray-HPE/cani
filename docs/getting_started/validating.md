# Validating

In many situations, it is helpful for the inventory provider to validate any changes and for the user to see them.

## Validate An Inventory

Running the `validate` subcommand will validate `cani`'s proposed changes with the external inventory provider.  If something is missing, the required information is shown for the user to take action on, often through the `update` subcommand.

```shell
# example, validate the existing inventory
cani alpha validate
```

If any issues are detected, it will show what is problematic for the provider:

```
1:35PM WRN This may fail in the HMS Simulator without Network information.
1:35PM ERR Inventory data validation errors encountered
1:35PM ERR   fbc29321-9519-42f1-a3d7-c7fbd15f1b2d: System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:1
1:35PM ERR     - Missing required information: Alias is not set
1:35PM ERR     - Missing required information: NID is not set
1:35PM ERR     - Missing required information: Role is not set
1:35PM ERR   8947602e-37a0-49a1-b019-a441e198da25: System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:1
1:35PM ERR     - Missing required information: Alias is not set
1:35PM ERR     - Missing required information: NID is not set
1:35PM ERR     - Missing required information: Role is not set
1:35PM ERR   ec50efb0-7953-4802-9274-e17e43caa4e8: System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:1->Node:0
1:35PM ERR     - Missing required information: Alias is not set
1:35PM ERR     - Missing required information: NID is not set
1:35PM ERR     - Missing required information: Role is not set
1:35PM ERR   7d4cd47e-558c-4075-91ae-043ca141717b: System:0->Cabinet:9000->Chassis:1->NodeBlade:0->NodeCard:0->Node:0
1:35PM ERR     - Missing required information: Alias is not set
1:35PM ERR     - Missing required information: NID is not set
1:35PM ERR     - Missing required information: Role is not set
```
