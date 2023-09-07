# Schema Docs

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |

| Property                                         | Pattern | Type             | Deprecated | Definition                               | Title/Description |
| ------------------------------------------------ | ------- | ---------------- | ---------- | ---------------------------------------- | ----------------- |
| + [manufacturer](#manufacturer )                 | No      | string           | No         | -                                        | -                 |
| + [model](#model )                               | No      | string           | No         | -                                        | -                 |
| + [hardware-type](#hardware-type )               | No      | enum (of string) | No         | In types.json#/definitions/hardware-type | -                 |
| + [slug](#slug )                                 | No      | string           | No         | In components.json#/definitions/slug     | -                 |
| - [part_number](#part_number )                   | No      | string           | No         | -                                        | -                 |
| - [u_height](#u_height )                         | No      | number           | No         | -                                        | -                 |
| - [is_full_depth](#is_full_depth )               | No      | boolean          | No         | -                                        | -                 |
| - [airflow](#airflow )                           | No      | enum (of string) | No         | -                                        | -                 |
| - [weight](#weight )                             | No      | number           | No         | -                                        | -                 |
| - [weight_unit](#weight_unit )                   | No      | enum (of string) | No         | -                                        | -                 |
| - [front_image](#front_image )                   | No      | boolean          | No         | -                                        | -                 |
| - [rear_image](#rear_image )                     | No      | boolean          | No         | -                                        | -                 |
| - [subdevice_role](#subdevice_role )             | No      | enum (of string) | No         | -                                        | -                 |
| - [console-ports](#console-ports )               | No      | array            | No         | -                                        | -                 |
| - [console-server-ports](#console-server-ports ) | No      | array            | No         | -                                        | -                 |
| - [power-ports](#power-ports )                   | No      | array            | No         | -                                        | -                 |
| - [power-outlets](#power-outlets )               | No      | array            | No         | -                                        | -                 |
| - [interfaces](#interfaces )                     | No      | array            | No         | -                                        | -                 |
| - [front-ports](#front-ports )                   | No      | array            | No         | -                                        | -                 |
| - [rear-ports](#rear-ports )                     | No      | array            | No         | -                                        | -                 |
| - [module-bays](#module-bays )                   | No      | array            | No         | -                                        | -                 |
| - [device-bays](#device-bays )                   | No      | array            | No         | -                                        | -                 |
| - [identifications](#identifications )           | No      | array            | No         | -                                        | -                 |
| - [inventory-items](#inventory-items )           | No      | array            | No         | -                                        | -                 |
| - [comments](#comments )                         | No      | string           | No         | -                                        | -                 |
| - [provider_defaults](#provider_defaults )       | No      | object           | No         | -                                        | -                 |

## <a name="manufacturer"></a>1. Property `root > manufacturer`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

## <a name="model"></a>2. Property `root > model`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

## <a name="hardware-type"></a>3. Property `root > hardware-type`

|                |                                       |
| -------------- | ------------------------------------- |
| **Type**       | `enum (of string)`                    |
| **Required**   | Yes                                   |
| **Defined in** | types.json#/definitions/hardware-type |

Must be one of:

* "Cabinet"
* "Chassis"
* "ChassisManagementModule"
* "CabinetEnvironmentalController"
* "NodeBlade"
* "NodeCard"
* "NodeController"
* "Node"
* "ManagementSwitchEnclosure"
* "ManagementSwitch"
* "ManagementSwitchController"
* "HighSpeedSwitchEnclosure"
* "HighSpeedSwitch"
* "HighSpeedSwitchController"
* "CabinetPDUController"
* "CabinetPDU"
* "CoolingDistributionUnit"

## <a name="slug"></a>4. Property `root > slug`

|                |                                   |
| -------------- | --------------------------------- |
| **Type**       | `string`                          |
| **Required**   | Yes                               |
| **Defined in** | components.json#/definitions/slug |

| Restrictions                      |                                                                                 |
| --------------------------------- | ------------------------------------------------------------------------------- |
| **Must match regular expression** | ```^[-a-z0-9_]+$``` [Test](https://regex101.com/?regex=%5E%5B-a-z0-9_%5D%2B%24) |

## <a name="part_number"></a>5. Property `root > part_number`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="u_height"></a>6. Property `root > u_height`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

| Restrictions    |        |
| --------------- | ------ |
| **Multiple of** | 0.5    |
| **Minimum**     | &ge; 0 |

## <a name="is_full_depth"></a>7. Property `root > is_full_depth`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

## <a name="airflow"></a>8. Property `root > airflow`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "front-to-rear"
* "rear-to-front"
* "left-to-right"
* "right-to-left"
* "side-to-rear"
* "passive"

## <a name="weight"></a>9. Property `root > weight`

|              |          |
| ------------ | -------- |
| **Type**     | `number` |
| **Required** | No       |

| Restrictions    |        |
| --------------- | ------ |
| **Multiple of** | 0.01   |
| **Minimum**     | &ge; 0 |

## <a name="weight_unit"></a>10. Property `root > weight_unit`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "kg"
* "g"
* "lb"
* "oz"

## <a name="front_image"></a>11. Property `root > front_image`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

## <a name="rear_image"></a>12. Property `root > rear_image`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

## <a name="subdevice_role"></a>13. Property `root > subdevice_role`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "parent"
* "child"

## <a name="console-ports"></a>14. Property `root > console-ports`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be      | Description |
| ------------------------------------ | ----------- |
| [console-port](#console-ports_items) | -           |

### <a name="autogenerated_heading_2"></a>14.1. root > console-ports > console-port

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/console-port               |

| Property                               | Pattern | Type             | Deprecated | Definition | Title/Description |
| -------------------------------------- | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#console-ports_items_name )   | No      | string           | No         | -          | -                 |
| - [label](#console-ports_items_label ) | No      | string           | No         | -          | -                 |
| + [type](#console-ports_items_type )   | No      | enum (of string) | No         | -          | -                 |

#### <a name="console-ports_items_name"></a>14.1.1. Property `root > console-ports > console-ports items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="console-ports_items_label"></a>14.1.2. Property `root > console-ports > console-ports items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="console-ports_items_type"></a>14.1.3. Property `root > console-ports > console-ports items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "de-9"
* "db-25"
* "rj-11"
* "rj-12"
* "rj-45"
* "mini-din-8"
* "usb-a"
* "usb-b"
* "usb-c"
* "usb-mini-a"
* "usb-mini-b"
* "usb-micro-a"
* "usb-micro-b"
* "usb-micro-ab"
* "other"

## <a name="console-server-ports"></a>15. Property `root > console-server-ports`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be                    | Description |
| -------------------------------------------------- | ----------- |
| [console-server-port](#console-server-ports_items) | -           |

### <a name="autogenerated_heading_3"></a>15.1. root > console-server-ports > console-server-port

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/console-server-port        |

| Property                                      | Pattern | Type             | Deprecated | Definition | Title/Description |
| --------------------------------------------- | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#console-server-ports_items_name )   | No      | string           | No         | -          | -                 |
| - [label](#console-server-ports_items_label ) | No      | string           | No         | -          | -                 |
| + [type](#console-server-ports_items_type )   | No      | enum (of string) | No         | -          | -                 |

#### <a name="console-server-ports_items_name"></a>15.1.1. Property `root > console-server-ports > console-server-ports items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="console-server-ports_items_label"></a>15.1.2. Property `root > console-server-ports > console-server-ports items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="console-server-ports_items_type"></a>15.1.3. Property `root > console-server-ports > console-server-ports items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "de-9"
* "db-25"
* "rj-12"
* "rj-45"
* "mini-din-8"
* "usb-a"
* "usb-b"
* "usb-c"
* "usb-mini-a"
* "usb-mini-b"
* "usb-micro-a"
* "usb-micro-b"
* "usb-micro-ab"
* "other"

## <a name="power-ports"></a>16. Property `root > power-ports`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be  | Description |
| -------------------------------- | ----------- |
| [power-port](#power-ports_items) | -           |

### <a name="autogenerated_heading_4"></a>16.1. root > power-ports > power-port

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/power-port                 |

| Property                                               | Pattern | Type             | Deprecated | Definition | Title/Description |
| ------------------------------------------------------ | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#power-ports_items_name )                     | No      | string           | No         | -          | -                 |
| - [label](#power-ports_items_label )                   | No      | string           | No         | -          | -                 |
| + [type](#power-ports_items_type )                     | No      | enum (of string) | No         | -          | -                 |
| - [maximum_draw](#power-ports_items_maximum_draw )     | No      | integer          | No         | -          | -                 |
| - [allocated_draw](#power-ports_items_allocated_draw ) | No      | integer          | No         | -          | -                 |

#### <a name="power-ports_items_name"></a>16.1.1. Property `root > power-ports > power-ports items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="power-ports_items_label"></a>16.1.2. Property `root > power-ports > power-ports items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="power-ports_items_type"></a>16.1.3. Property `root > power-ports > power-ports items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "iec-60320-c6"
* "iec-60320-c8"
* "iec-60320-c14"
* "iec-60320-c16"
* "iec-60320-c20"
* "iec-60320-c22"
* "iec-60309-p-n-e-4h"
* "iec-60309-p-n-e-6h"
* "iec-60309-p-n-e-9h"
* "iec-60309-2p-e-4h"
* "iec-60309-2p-e-6h"
* "iec-60309-2p-e-9h"
* "iec-60309-3p-e-4h"
* "iec-60309-3p-e-6h"
* "iec-60309-3p-e-9h"
* "iec-60309-3p-n-e-4h"
* "iec-60309-3p-n-e-6h"
* "iec-60309-3p-n-e-9h"
* "nema-1-15p"
* "nema-5-15p"
* "nema-5-20p"
* "nema-5-30p"
* "nema-5-50p"
* "nema-6-15p"
* "nema-6-20p"
* "nema-6-30p"
* "nema-6-50p"
* "nema-10-30p"
* "nema-10-50p"
* "nema-14-20p"
* "nema-14-30p"
* "nema-14-50p"
* "nema-14-60p"
* "nema-15-15p"
* "nema-15-20p"
* "nema-15-30p"
* "nema-15-50p"
* "nema-15-60p"
* "nema-l1-15p"
* "nema-l5-15p"
* "nema-l5-20p"
* "nema-l5-30p"
* "nema-l5-50p"
* "nema-l6-15p"
* "nema-l6-20p"
* "nema-l6-30p"
* "nema-l6-50p"
* "nema-l10-30p"
* "nema-l14-20p"
* "nema-l14-30p"
* "nema-l14-50p"
* "nema-l14-60p"
* "nema-l15-20p"
* "nema-l15-30p"
* "nema-l15-50p"
* "nema-l15-60p"
* "nema-l21-20p"
* "nema-l21-30p"
* "nema-l22-30p"
* "cs6361c"
* "cs6365c"
* "cs8165c"
* "cs8265c"
* "cs8365c"
* "cs8465c"
* "ita-c"
* "ita-e"
* "ita-f"
* "ita-ef"
* "ita-g"
* "ita-h"
* "ita-i"
* "ita-j"
* "ita-k"
* "ita-l"
* "ita-m"
* "ita-n"
* "ita-o"
* "usb-a"
* "usb-b"
* "usb-c"
* "usb-mini-a"
* "usb-mini-b"
* "usb-micro-a"
* "usb-micro-b"
* "usb-micro-ab"
* "usb-3-b"
* "usb-3-micro-b"
* "dc-terminal"
* "saf-d-grid"
* "ubiquiti-smartpower"
* "hardwired"
* "other"

#### <a name="power-ports_items_maximum_draw"></a>16.1.4. Property `root > power-ports > power-ports items > maximum_draw`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

#### <a name="power-ports_items_allocated_draw"></a>16.1.5. Property `root > power-ports > power-ports items > allocated_draw`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

## <a name="power-outlets"></a>17. Property `root > power-outlets`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be      | Description |
| ------------------------------------ | ----------- |
| [power-outlet](#power-outlets_items) | -           |

### <a name="autogenerated_heading_5"></a>17.1. root > power-outlets > power-outlet

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/power-outlet               |

| Property                                         | Pattern | Type             | Deprecated | Definition | Title/Description |
| ------------------------------------------------ | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#power-outlets_items_name )             | No      | string           | No         | -          | -                 |
| - [label](#power-outlets_items_label )           | No      | string           | No         | -          | -                 |
| + [type](#power-outlets_items_type )             | No      | enum (of string) | No         | -          | -                 |
| - [power_port](#power-outlets_items_power_port ) | No      | string           | No         | -          | -                 |
| - [feed_leg](#power-outlets_items_feed_leg )     | No      | enum (of string) | No         | -          | -                 |

#### <a name="power-outlets_items_name"></a>17.1.1. Property `root > power-outlets > power-outlets items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="power-outlets_items_label"></a>17.1.2. Property `root > power-outlets > power-outlets items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="power-outlets_items_type"></a>17.1.3. Property `root > power-outlets > power-outlets items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "iec-60320-c5"
* "iec-60320-c7"
* "iec-60320-c13"
* "iec-60320-c15"
* "iec-60320-c19"
* "iec-60320-c21"
* "iec-60309-p-n-e-4h"
* "iec-60309-p-n-e-6h"
* "iec-60309-p-n-e-9h"
* "iec-60309-2p-e-4h"
* "iec-60309-2p-e-6h"
* "iec-60309-2p-e-9h"
* "iec-60309-3p-e-4h"
* "iec-60309-3p-e-6h"
* "iec-60309-3p-e-9h"
* "iec-60309-3p-n-e-4h"
* "iec-60309-3p-n-e-6h"
* "iec-60309-3p-n-e-9h"
* "nema-1-15r"
* "nema-5-15r"
* "nema-5-20r"
* "nema-5-30r"
* "nema-5-50r"
* "nema-6-15r"
* "nema-6-20r"
* "nema-6-30r"
* "nema-6-50r"
* "nema-10-30r"
* "nema-10-50r"
* "nema-14-20r"
* "nema-14-30r"
* "nema-14-50r"
* "nema-14-60r"
* "nema-15-15r"
* "nema-15-20r"
* "nema-15-30r"
* "nema-15-50r"
* "nema-15-60r"
* "nema-l1-15r"
* "nema-l5-15r"
* "nema-l5-20r"
* "nema-l5-30r"
* "nema-l5-50r"
* "nema-l6-15r"
* "nema-l6-20r"
* "nema-l6-30r"
* "nema-l6-50r"
* "nema-l10-30r"
* "nema-l14-20r"
* "nema-l14-30r"
* "nema-l14-50r"
* "nema-l14-60r"
* "nema-l15-20r"
* "nema-l15-30r"
* "nema-l15-50r"
* "nema-l15-60r"
* "nema-l21-20r"
* "nema-l21-30r"
* "nema-l22-30r"
* "CS6360C"
* "CS6364C"
* "CS8164C"
* "CS8264C"
* "CS8364C"
* "CS8464C"
* "ita-e"
* "ita-f"
* "ita-g"
* "ita-h"
* "ita-i"
* "ita-j"
* "ita-k"
* "ita-l"
* "ita-m"
* "ita-n"
* "ita-o"
* "ita-multistandard"
* "usb-a"
* "usb-micro-b"
* "usb-c"
* "dc-terminal"
* "hdot-cx"
* "saf-d-grid"
* "ubiquiti-smartpower"
* "hardwired"
* "other"

#### <a name="power-outlets_items_power_port"></a>17.1.4. Property `root > power-outlets > power-outlets items > power_port`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="power-outlets_items_feed_leg"></a>17.1.5. Property `root > power-outlets > power-outlets items > feed_leg`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "A"
* "B"
* "C"

## <a name="interfaces"></a>18. Property `root > interfaces`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be | Description |
| ------------------------------- | ----------- |
| [interface](#interfaces_items)  | -           |

### <a name="autogenerated_heading_6"></a>18.1. root > interfaces > interface

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/interface                  |

| Property                                    | Pattern | Type             | Deprecated | Definition | Title/Description |
| ------------------------------------------- | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#interfaces_items_name )           | No      | string           | No         | -          | -                 |
| - [label](#interfaces_items_label )         | No      | string           | No         | -          | -                 |
| + [type](#interfaces_items_type )           | No      | enum (of string) | No         | -          | -                 |
| - [poe_mode](#interfaces_items_poe_mode )   | No      | enum (of string) | No         | -          | -                 |
| - [poe_type](#interfaces_items_poe_type )   | No      | enum (of string) | No         | -          | -                 |
| - [mgmt_only](#interfaces_items_mgmt_only ) | No      | boolean          | No         | -          | -                 |

#### <a name="interfaces_items_name"></a>18.1.1. Property `root > interfaces > interfaces items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="interfaces_items_label"></a>18.1.2. Property `root > interfaces > interfaces items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="interfaces_items_type"></a>18.1.3. Property `root > interfaces > interfaces items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "virtual"
* "bridge"
* "lag"
* "100base-fx"
* "100base-lfx"
* "100base-tx"
* "100base-t1"
* "1000base-t"
* "1000base-x-gbic"
* "1000base-x-sfp"
* "2.5gbase-t"
* "5gbase-t"
* "10gbase-t"
* "10gbase-cx4"
* "10gbase-x-sfpp"
* "10gbase-x-xfp"
* "10gbase-x-xenpak"
* "10gbase-x-x2"
* "25gbase-x-sfp28"
* "40gbase-x-qsfpp"
* "50gbase-x-sfp28"
* "100gbase-x-cfp"
* "100gbase-x-cfp2"
* "100gbase-x-cfp4"
* "100gbase-x-cpak"
* "100gbase-x-qsfp28"
* "200gbase-x-cfp2"
* "200gbase-x-qsfp56"
* "400gbase-x-qsfpdd"
* "400gbase-x-osfp"
* "ieee802.11a"
* "ieee802.11g"
* "ieee802.11n"
* "ieee802.11ac"
* "ieee802.11ad"
* "ieee802.11ax"
* "ieee802.15.1"
* "gsm"
* "cdma"
* "lte"
* "sonet-oc3"
* "sonet-oc12"
* "sonet-oc48"
* "sonet-oc192"
* "sonet-oc768"
* "sonet-oc1920"
* "sonet-oc3840"
* "1gfc-sfp"
* "2gfc-sfp"
* "4gfc-sfp"
* "8gfc-sfpp"
* "16gfc-sfpp"
* "32gfc-sfp28"
* "64gfc-qsfpp"
* "128gfc-qsfp28"
* "infiniband-sdr"
* "infiniband-ddr"
* "infiniband-qdr"
* "infiniband-fdr10"
* "infiniband-fdr"
* "infiniband-edr"
* "infiniband-hdr"
* "infiniband-ndr"
* "infiniband-xdr"
* "t1"
* "e1"
* "t3"
* "e3"
* "xdsl"
* "docsis"
* "cisco-stackwise"
* "cisco-stackwise-plus"
* "cisco-flexstack"
* "cisco-flexstack-plus"
* "cisco-stackwise-80"
* "cisco-stackwise-160"
* "cisco-stackwise-320"
* "cisco-stackwise-480"
* "juniper-vcp"
* "extreme-summitstack"
* "extreme-summitstack-128"
* "extreme-summitstack-256"
* "extreme-summitstack-512"
* "gpon"
* "xg-pon"
* "xgs-pon"
* "ng-pon2"
* "epon"
* "10g-epon"
* "other"

#### <a name="interfaces_items_poe_mode"></a>18.1.4. Property `root > interfaces > interfaces items > poe_mode`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "pd"
* "pse"

#### <a name="interfaces_items_poe_type"></a>18.1.5. Property `root > interfaces > interfaces items > poe_type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | No                 |

Must be one of:

* "type1-ieee802.3af"
* "type2-ieee802.3at"
* "type3-ieee802.3bt"
* "type4-ieee802.3bt"
* "passive-24v-2pair"
* "passive-24v-4pair"
* "passive-48v-2pair"
* "passive-48v-4pair"

#### <a name="interfaces_items_mgmt_only"></a>18.1.6. Property `root > interfaces > interfaces items > mgmt_only`

|              |           |
| ------------ | --------- |
| **Type**     | `boolean` |
| **Required** | No        |

## <a name="front-ports"></a>19. Property `root > front-ports`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be  | Description |
| -------------------------------- | ----------- |
| [front-port](#front-ports_items) | -           |

### <a name="autogenerated_heading_7"></a>19.1. root > front-ports > front-port

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |
| **Defined in**            | components.json#/definitions/front-port                                   |

| Property                                                       | Pattern | Type             | Deprecated | Definition | Title/Description |
| -------------------------------------------------------------- | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#front-ports_items_name )                             | No      | string           | No         | -          | -                 |
| - [label](#front-ports_items_label )                           | No      | string           | No         | -          | -                 |
| + [type](#front-ports_items_type )                             | No      | enum (of string) | No         | -          | -                 |
| - [color](#front-ports_items_color )                           | No      | string           | No         | -          | -                 |
| + [rear_port](#front-ports_items_rear_port )                   | No      | string           | No         | -          | -                 |
| - [rear_port_position](#front-ports_items_rear_port_position ) | No      | integer          | No         | -          | -                 |

#### <a name="front-ports_items_name"></a>19.1.1. Property `root > front-ports > front-ports items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="front-ports_items_label"></a>19.1.2. Property `root > front-ports > front-ports items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="front-ports_items_type"></a>19.1.3. Property `root > front-ports > front-ports items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "8p8c"
* "8p6c"
* "8p4c"
* "8p2c"
* "6p6c"
* "6p4c"
* "6p2c"
* "4p4c"
* "4p2c"
* "gg45"
* "tera-4p"
* "tera-2p"
* "tera-1p"
* "110-punch"
* "bnc"
* "f"
* "n"
* "mrj21"
* "st"
* "sc"
* "sc-apc"
* "fc"
* "lc"
* "lc-apc"
* "mtrj"
* "mpo"
* "lsh"
* "lsh-apc"
* "splice"
* "cs"
* "sn"
* "sma-905"
* "sma-906"
* "urm-p2"
* "urm-p4"
* "urm-p8"
* "other"

#### <a name="front-ports_items_color"></a>19.1.4. Property `root > front-ports > front-ports items > color`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

| Restrictions                      |                                                                                   |
| --------------------------------- | --------------------------------------------------------------------------------- |
| **Must match regular expression** | ```^[a-f0-9]{6}$``` [Test](https://regex101.com/?regex=%5E%5Ba-f0-9%5D%7B6%7D%24) |

#### <a name="front-ports_items_rear_port"></a>19.1.5. Property `root > front-ports > front-ports items > rear_port`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="front-ports_items_rear_port_position"></a>19.1.6. Property `root > front-ports > front-ports items > rear_port_position`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

## <a name="rear-ports"></a>20. Property `root > rear-ports`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be | Description |
| ------------------------------- | ----------- |
| [rear-port](#rear-ports_items)  | -           |

### <a name="autogenerated_heading_8"></a>20.1. root > rear-ports > rear-port

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/rear-port                  |

| Property                                    | Pattern | Type             | Deprecated | Definition | Title/Description |
| ------------------------------------------- | ------- | ---------------- | ---------- | ---------- | ----------------- |
| + [name](#rear-ports_items_name )           | No      | string           | No         | -          | -                 |
| - [label](#rear-ports_items_label )         | No      | string           | No         | -          | -                 |
| + [type](#rear-ports_items_type )           | No      | enum (of string) | No         | -          | -                 |
| - [color](#rear-ports_items_color )         | No      | string           | No         | -          | -                 |
| - [positions](#rear-ports_items_positions ) | No      | integer          | No         | -          | -                 |

#### <a name="rear-ports_items_name"></a>20.1.1. Property `root > rear-ports > rear-ports items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="rear-ports_items_label"></a>20.1.2. Property `root > rear-ports > rear-ports items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="rear-ports_items_type"></a>20.1.3. Property `root > rear-ports > rear-ports items > type`

|              |                    |
| ------------ | ------------------ |
| **Type**     | `enum (of string)` |
| **Required** | Yes                |

Must be one of:

* "8p8c"
* "8p6c"
* "8p4c"
* "8p2c"
* "6p6c"
* "6p4c"
* "6p2c"
* "4p4c"
* "4p2c"
* "gg45"
* "tera-4p"
* "tera-2p"
* "tera-1p"
* "110-punch"
* "bnc"
* "f"
* "n"
* "mrj21"
* "st"
* "sc"
* "sc-apc"
* "fc"
* "lc"
* "lc-apc"
* "mtrj"
* "mpo"
* "lsh"
* "lsh-apc"
* "splice"
* "cs"
* "sn"
* "sma-905"
* "sma-906"
* "urm-p2"
* "urm-p4"
* "urm-p8"

#### <a name="rear-ports_items_color"></a>20.1.4. Property `root > rear-ports > rear-ports items > color`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

| Restrictions                      |                                                                                   |
| --------------------------------- | --------------------------------------------------------------------------------- |
| **Must match regular expression** | ```^[a-f0-9]{6}$``` [Test](https://regex101.com/?regex=%5E%5Ba-f0-9%5D%7B6%7D%24) |

#### <a name="rear-ports_items_positions"></a>20.1.5. Property `root > rear-ports > rear-ports items > positions`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

## <a name="module-bays"></a>21. Property `root > module-bays`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be  | Description |
| -------------------------------- | ----------- |
| [module-bay](#module-bays_items) | -           |

### <a name="autogenerated_heading_9"></a>21.1. root > module-bays > module-bay

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |
| **Defined in**            | components.json#/definitions/module-bay                                   |

| Property                                   | Pattern | Type   | Deprecated | Definition | Title/Description |
| ------------------------------------------ | ------- | ------ | ---------- | ---------- | ----------------- |
| + [name](#module-bays_items_name )         | No      | string | No         | -          | -                 |
| - [label](#module-bays_items_label )       | No      | string | No         | -          | -                 |
| - [position](#module-bays_items_position ) | No      | string | No         | -          | -                 |

#### <a name="module-bays_items_name"></a>21.1.1. Property `root > module-bays > module-bays items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="module-bays_items_label"></a>21.1.2. Property `root > module-bays > module-bays items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="module-bays_items_position"></a>21.1.3. Property `root > module-bays > module-bays items > position`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="device-bays"></a>22. Property `root > device-bays`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be  | Description |
| -------------------------------- | ----------- |
| [device-bay](#device-bays_items) | -           |

### <a name="autogenerated_heading_10"></a>22.1. root > device-bays > device-bay

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/device-bay                 |

| Property                                 | Pattern | Type    | Deprecated | Definition | Title/Description |
| ---------------------------------------- | ------- | ------- | ---------- | ---------- | ----------------- |
| + [name](#device-bays_items_name )       | No      | string  | No         | -          | -                 |
| - [label](#device-bays_items_label )     | No      | string  | No         | -          | -                 |
| + [ordinal](#device-bays_items_ordinal ) | No      | integer | No         | -          | -                 |
| - [allowed](#device-bays_items_allowed ) | No      | object  | No         | -          | -                 |
| - [default](#device-bays_items_default ) | No      | object  | No         | -          | -                 |

#### <a name="device-bays_items_name"></a>22.1.1. Property `root > device-bays > device-bays items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="device-bays_items_label"></a>22.1.2. Property `root > device-bays > device-bays items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="device-bays_items_ordinal"></a>22.1.3. Property `root > device-bays > device-bays items > ordinal`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | Yes       |

#### <a name="device-bays_items_allowed"></a>22.1.4. Property `root > device-bays > device-bays items > allowed`

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |

| Property                                                     | Pattern | Type  | Deprecated | Definition | Title/Description |
| ------------------------------------------------------------ | ------- | ----- | ---------- | ---------- | ----------------- |
| - [slug](#device-bays_items_allowed_slug )                   | No      | array | No         | -          | -                 |
| - [hardware-type](#device-bays_items_allowed_hardware-type ) | No      | array | No         | -          | -                 |

##### <a name="device-bays_items_allowed_slug"></a>22.1.4.1. Property `root > device-bays > device-bays items > allowed > slug`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be               | Description |
| --------------------------------------------- | ----------- |
| [slug](#device-bays_items_allowed_slug_items) | -           |

##### <a name="autogenerated_heading_11"></a>22.1.4.1.1. root > device-bays > device-bays items > allowed > slug > slug

|                        |               |
| ---------------------- | ------------- |
| **Type**               | `string`      |
| **Required**           | No            |
| **Same definition as** | [slug](#slug) |

##### <a name="device-bays_items_allowed_hardware-type"></a>22.1.4.2. Property `root > device-bays > device-bays items > allowed > hardware-type`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be                                 | Description |
| --------------------------------------------------------------- | ----------- |
| [hardware-type](#device-bays_items_allowed_hardware-type_items) | -           |

##### <a name="autogenerated_heading_12"></a>22.1.4.2.1. root > device-bays > device-bays items > allowed > hardware-type > hardware-type

|                        |                                 |
| ---------------------- | ------------------------------- |
| **Type**               | `enum (of string)`              |
| **Required**           | No                              |
| **Same definition as** | [hardware-type](#hardware-type) |

#### <a name="device-bays_items_default"></a>22.1.5. Property `root > device-bays > device-bays items > default`

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |

| Property                                   | Pattern | Type   | Deprecated | Definition             | Title/Description |
| ------------------------------------------ | ------- | ------ | ---------- | ---------------------- | ----------------- |
| - [slug](#device-bays_items_default_slug ) | No      | string | No         | Same as [slug](#slug ) | -                 |

##### <a name="device-bays_items_default_slug"></a>22.1.5.1. Property `root > device-bays > device-bays items > default > slug`

|                        |               |
| ---------------------- | ------------- |
| **Type**               | `string`      |
| **Required**           | No            |
| **Same definition as** | [slug](#slug) |

## <a name="identifications"></a>23. Property `root > identifications`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be           | Description |
| ----------------------------------------- | ----------- |
| [identifications](#identifications_items) | -           |

### <a name="autogenerated_heading_13"></a>23.1. root > identifications > identifications

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/identifications            |

| Property                                               | Pattern | Type   | Deprecated | Definition | Title/Description |
| ------------------------------------------------------ | ------- | ------ | ---------- | ---------- | ----------------- |
| + [manufacturer](#identifications_items_manufacturer ) | No      | string | No         | -          | -                 |
| + [model](#identifications_items_model )               | No      | string | No         | -          | -                 |
| - [part-number](#identifications_items_part-number )   | No      | string | No         | -          | -                 |

#### <a name="identifications_items_manufacturer"></a>23.1.1. Property `root > identifications > identifications > manufacturer`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="identifications_items_model"></a>23.1.2. Property `root > identifications > identifications > model`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="identifications_items_part-number"></a>23.1.3. Property `root > identifications > identifications > part-number`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="inventory-items"></a>24. Property `root > inventory-items`

|              |         |
| ------------ | ------- |
| **Type**     | `array` |
| **Required** | No      |

|                      | Array restrictions |
| -------------------- | ------------------ |
| **Min items**        | N/A                |
| **Max items**        | N/A                |
| **Items unicity**    | False              |
| **Additional items** | False              |
| **Tuple validation** | See below          |

| Each item of this array must be          | Description |
| ---------------------------------------- | ----------- |
| [inventory-item](#inventory-items_items) | -           |

### <a name="autogenerated_heading_14"></a>24.1. root > inventory-items > inventory-item

|                           |                                                         |
| ------------------------- | ------------------------------------------------------- |
| **Type**                  | `object`                                                |
| **Required**              | No                                                      |
| **Additional properties** | [[Not allowed]](# "Additional Properties not allowed.") |
| **Defined in**            | components.json#/definitions/inventory-item             |

| Property                                               | Pattern | Type   | Deprecated | Definition | Title/Description |
| ------------------------------------------------------ | ------- | ------ | ---------- | ---------- | ----------------- |
| + [name](#inventory-items_items_name )                 | No      | string | No         | -          | -                 |
| - [label](#inventory-items_items_label )               | No      | string | No         | -          | -                 |
| - [manufacturer](#inventory-items_items_manufacturer ) | No      | string | No         | -          | -                 |
| - [part_id](#inventory-items_items_part_id )           | No      | string | No         | -          | -                 |

#### <a name="inventory-items_items_name"></a>24.1.1. Property `root > inventory-items > inventory-items items > name`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | Yes      |

#### <a name="inventory-items_items_label"></a>24.1.2. Property `root > inventory-items > inventory-items items > label`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="inventory-items_items_manufacturer"></a>24.1.3. Property `root > inventory-items > inventory-items items > manufacturer`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="inventory-items_items_part_id"></a>24.1.4. Property `root > inventory-items > inventory-items items > part_id`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="comments"></a>25. Property `root > comments`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

## <a name="provider_defaults"></a>26. Property `root > provider_defaults`

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |

| Property                         | Pattern | Type   | Deprecated | Definition | Title/Description |
| -------------------------------- | ------- | ------ | ---------- | ---------- | ----------------- |
| - [csm](#provider_defaults_csm ) | No      | object | No         | -          | -                 |

### <a name="provider_defaults_csm"></a>26.1. Property `root > provider_defaults > csm`

|                           |                                                                           |
| ------------------------- | ------------------------------------------------------------------------- |
| **Type**                  | `object`                                                                  |
| **Required**              | No                                                                        |
| **Additional properties** | [[Any type: allowed]](# "Additional Properties of any type are allowed.") |

| Property                                                     | Pattern | Type    | Deprecated | Definition | Title/Description |
| ------------------------------------------------------------ | ------- | ------- | ---------- | ---------- | ----------------- |
| - [Class](#provider_defaults_csm_Class )                     | No      | string  | No         | -          | -                 |
| - [Ordinal](#provider_defaults_csm_Ordinal )                 | No      | integer | No         | -          | -                 |
| - [StartingHmnVlan](#provider_defaults_csm_StartingHmnVlan ) | No      | integer | No         | -          | -                 |
| - [EndingHmnVlan](#provider_defaults_csm_EndingHmnVlan )     | No      | integer | No         | -          | -                 |

#### <a name="provider_defaults_csm_Class"></a>26.1.1. Property `root > provider_defaults > csm > Class`

|              |          |
| ------------ | -------- |
| **Type**     | `string` |
| **Required** | No       |

#### <a name="provider_defaults_csm_Ordinal"></a>26.1.2. Property `root > provider_defaults > csm > Ordinal`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

#### <a name="provider_defaults_csm_StartingHmnVlan"></a>26.1.3. Property `root > provider_defaults > csm > StartingHmnVlan`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

#### <a name="provider_defaults_csm_EndingHmnVlan"></a>26.1.4. Property `root > provider_defaults > csm > EndingHmnVlan`

|              |           |
| ------------ | --------- |
| **Type**     | `integer` |
| **Required** | No        |

----------------------------------------------------------------------------------------------------------------------------
Generated using [json-schema-for-humans](https://github.com/coveooss/json-schema-for-humans) on 2023-09-06 at 13:03:52 -0500
