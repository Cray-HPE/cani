# Export and Import Data using CSV

The state of a session can be exported and imported in a CSV format (Comma-Separated Values).
The CSV data can be viewed and edited in a text editor or any spreadsheet software that supports CSV.

Both export and import require that a session has been started.

## Export

This example, assumes that a session has been initialized.

```shell
# example: export data and save it to file.csv
cani alpha export > file.csv
```

The types of hardware that is exported can be specified

```shell
# example: export data for all types of hardware
cani alpha export --all
```

```shell
# example: export data for just Nodes and NodeBlades
cani alpha export --type node,nodeblade
```

The columns in the CSV can also be controlled by selecting specific headers
```shell
# example: export vlans for cabinets
cani alpha export --type cabinet --headers id,type,vlan
```

### Import

The CSV file can be modified and then imported back into cani's data.
This is a way to make bulk changes, such as, setting the NIDs.
Only some of the fields can be changed. The modifiable fields include: Vlan, Nid, Alias, Role, and SubRole.

```shell
# example: import data from file
cani alpha import file.csv
```

```shell
# example: import data standard input
cat file.csv | cani alpha import
cat file.csv | cani alpha import -
```

