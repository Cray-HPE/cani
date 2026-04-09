---
applyTo: "pkg/devicetypes/**/*.yaml"
---
# Devicetype YAML Lint Rules

When editing or reviewing this file, enforce these rules:

## Required Fields
- `slug` must be non-empty, lowercase kebab-case, and unique across the category
- `manufacturer` must match the parent directory name
- `type` must be a known value: blade, node, nodecard, chassis, switch, mgmt-switch, hsn-switch, cabinet-pdu, cdu, cec, cmm, nodecontroller, gpu, nic, adapter, transceiver, power-supply, cable, rack, cabinet

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
