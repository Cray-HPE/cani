# Inventory Format Overview

`cani`'s inventory format is a map of UUIDs.  In Go, this is represented as `map[uuid.UUID]Hardware`. 

```json
{
  "f7448392-1e1c-45d0-9c59-be7dfc44c15c": {
    "ID": "f7448392-1e1c-45d0-9c59-be7dfc44c15c",
    "Type": "Cabinet",
    "Vendor": "HPE",
    "Model": "EX2000",
    "Status": "staged",
    "Parent": "7e3de0fa-e3d6-421b-9d25-c0192d2a5966",
    "Children": [
      "00050177-a309-4fde-bf85-70452b228e24",
      "004ecb7f-50bb-4975-9973-b6c617d6cc82",
      "16c65e22-9401-4b26-aaed-44d5c55d650f"
    ],
    "LocationPath": [
      {
        "HardwareType": "System",
        "Ordinal": 0
      },
      {
        "HardwareType": "Cabinet",
        "Ordinal": 1001
      }
    ],
    "LocationOrdinal": 1001
  }
}
```

In the sample above, `f7448392-1e1c-45d0-9c59-be7dfc44c15c` is the key.  This piece of Hardware is of the `Type: Cabinet`.  It has `Children`, which are other UUID's.  The `LocationPath` is a machine-friendly slice of where it exits in the hardware tree and is not often used by humans, but can be useful.

```
{
  "f7448392-1e1c-45d0-9c59-be7dfc44c15c": {           <---uuid of a unique hardware item in the inventory
    "ID": "f7448392-1e1c-45d0-9c59-be7dfc44c15c",     <---the same uuid, but accessible as a readable field for other consumers
    "Type": "Cabinet",
    "Vendor": "HPE",
    "Model": "EX2000",
    "Status": "staged",
    "Parent": "7e3de0fa-e3d6-421b-9d25-c0192d2a5966",
    "Children": [
      "00050177-a309-4fde-bf85-70452b228e24",
      "004ecb7f-50bb-4975-9973-b6c617d6cc82",
      "16c65e22-9401-4b26-aaed-44d5c55d650f"
    ],
    "LocationPath": [
      {
        "HardwareType": "System",
        "Ordinal": 0
      },
      {
        "HardwareType": "Cabinet",
        "Ordinal": 1001
      }
    ],
    "LocationOrdinal": 1001
  }
}
```
