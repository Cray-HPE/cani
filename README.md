<p align="center">
  <img src="https://user-images.githubusercontent.com/3843505/235496554-806630e3-a818-4e04-8d46-6a024994d08f.png"" width="150" height="150" alt="cani">
  <br>
  <strong>cani: Cani's Automated Nomicon Inventory</strong>
</p>

> Can I manage an inventory of an entire datacenter? From subfloor to top-of-rack, **yes** you can.

# `cani` Inventory Hardware

You can inventory hardware with this utility.  Phase 1 of this tool communicates with SLS.  It also defines a portable inventory data structure that will allow other data sources to plug in such as HPCM, MAAS, Harvester, etc.  Utilizing Extract, Transform, and Load (ETL) functions, any potential data source can be integrated with this utility.

Phase 1 implements some basic functional commands that are consistent across activites someone needs to do to maintain their systems:

- `add` 
- `remove`
- `list` 

Some simple examples of this might be:

```shell
cani list
cani add switch [FLAGS]...
cani remove cabinet [FLAGS]...
```
