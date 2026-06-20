---
applyTo: "pkg/devicetypes/**/*.yaml"
---
# Devicetype YAML Lint Rules

When editing or reviewing this file, enforce these rules:

## Required Fields
- `slug` must be non-empty, lowercase kebab-case, and unique across the category
- `manufacturer` must match the parent directory name
- `type` must be a known value: blade, node, nodecard, chassis, switch, mgmt-switch, hsn-switch, cabinet-pdu, cdu, cec, cmm, nodecontroller, gpu, nic, adapter, transceiver, power-supply, cable, rack, cabinet

## Field Key Casing (NetBox devicetype-library parity)
- Component-collection keys are **kebab-case** and must stay that way: `console-ports`, `power-ports`, `module-bays`, `device-bays`, `allowed-children`, `hardware-type`
- Do NOT rename these to snake_case — the loader and the 50+ bundled library files depend on the kebab-case keys, and snake_casing them breaks loading the hardware library
- Scalar fields and entry sub-fields keep NetBox's snake_case: `part_number`, `u_height`, `is_full_depth`, `subdevice_role`, `weight_unit`, `mgmt_only`, `maximum_draw`

## License Header
- File must begin with the MIT license comment block (lines starting with `#`)
- A `---` separator must appear between the license block and the YAML content

## Interfaces
- `type` must use a valid constant: 1000base-t, 10gbase-t, 10gbase-x-sfpp, 25gbase-x-sfp28, 40gbase-x-qsfpp, 100gbase-x-qsfp28, 200gbase-x-qsfp56, 400gbase-x-qsfpdd, 400gbase-x-osfp, virtual, lag
- Management-only ports (iLO, BMC, mgmt0) must set `mgmt_only: true`

## Formatting
- 2-space indentation, no tabs
- Slug convention: `<manufacturer>-<model>-<variant>` in lowercase kebab-case
- Description should be present and accurate
