# CANI

`cani` is a hardware inventory tool. It provides its own portable inventory format while retaining compatibility with external inventory providers. This makes it possible to use `cani` as either a main inventory source or to migrate from one inventory format to another.

## Quickstart

This shows a quick overview of using `cani` to import inventory data from an external provider, manipulate it locally, and export the changes back.

```shell
# Import inventory from a Nautobot instance
cani alpha import nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN

# Add a rack
cani alpha add rack hpe-48u-800mmx1200mm-g2-enterprise-shock-rack

# Add a device into the rack
cani alpha add device hpe-crayex-ex420-compute-blade --auto --accept

# Classify any unclassified devices
cani alpha classify --auto

# Show the current inventory
cani alpha show device

# Update a device
cani alpha update device --role Compute --alias nid00001

# Export the changes back to the provider
cani alpha export nautobot --url http://localhost:8081 --token $NAUTOBOT_TOKEN
```

