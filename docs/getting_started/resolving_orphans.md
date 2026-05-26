# Resolving Orphans

When devices are imported without a parent rack or location, they are considered "orphaned." The `update orphans` command resolves these by interactively assigning parents.

## Interactive Resolution

```shell
# Walk through all orphaned racks and devices
cani alpha update orphans
```

Each orphan is presented with candidate parents ranked by:

- Name similarity
- Hardware-type affinity
- Provider metadata

## Dry Run

Preview the assignments without modifying the inventory:

```shell
# Save assignments to a plan file
cani alpha update orphans --dry-run
```

The plan file is saved to `~/.cani/resolve-plan.json` and can be reviewed or edited before applying.

## Apply A Plan

Apply a previously saved plan file:

```shell
# Apply the plan without prompting
cani alpha update orphans --apply-plan ~/.cani/resolve-plan.json
```

## Typical Workflow

Orphan resolution is usually needed after importing from sources that don't include rack placement:

```shell
# 1. Import devices (some may not have rack info)
cani alpha import redfish --root ./redfish-roots.json

# 2. Add racks if needed
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack

# 3. Resolve orphans
cani alpha update orphans
```
