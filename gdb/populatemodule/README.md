# Population of the GDB

2 different ways:  

- Use CSV import clause of cypher (`csvimport`).
- Parse CSV data and run creation Cypher queries against a Neo4j database (`cypherimport`).

**If not previously installed**, it is needed to install the Neo4j Go driver to be able 
to interact with a Neo4j instance through our golang population application. See the 
[reference Neo4j Go Driver page](https://neo4j.com/docs/go-manual/current/) to do 
the installation.

### Connecting with the (stable) Neo4j graph database

In the `populatemodule` the connection with the corresponding Neo4j graph database instance is done using the `SafeConnect()` method that uses the specified URI, USERNAME and PASSWORD in the required `.env` file. The `.env` file needs to have the `NEO4J_URI`, `NEO4J_USERNAME` and `NEO4J_PASSWORD` and needs to be placed in the indicated directory (*depending on the chosen populating way*), for example:
```
NEO4J_URI="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="xxxxxxx"
```

Once the connection is established, all the cypher queries to do the population process in any of the two different ways are done using the `writeQuery()` function.

## 1. CSV import (csvimport)

Use of CSV import clause of cypher. See documentation [here](https://neo4j.com/docs/cypher-manual/5/clauses/load-csv/).

**Requirements**: 
- Place all the CSVs to import under the `/var/lib/neo4j/import` directory.  
- Place a `.env` file in the `csvimport` directory indicating the `NEO4J_URI`, `NEO4J_USERNAME` and `NEO4J_PASSWORD`
for example:
```
NEO4J_URI="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="xxxxxxx"
```
- Run the golang program in the `csvimport` subdirectory.

```
$> go run main.go
```

## 2. Creation of cypher queries (cypherimport)

Using a language library (of golang) to parse CSV data and run creation Cypher queries against a Neo4j database.

*Needed to create the graph database from CSV files directly from them, accesing these files from the same machine as
where we run the golang program. In our case to create the gdb on the cluster VM without having to place the CSVs on 
that machine*.

**Requirements**: 
- Indicate the CSV folder path when as argument to the main program.
- Place a `.env` file in the `cypherimport` directory indicating the `NEO4J_URI`, `NEO4J_USERNAME` and `NEO4J_PASSWORD`
for example:
```
NEO4J_URI="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="xxxxxxx"
```

- In our case, since the VM is hosted in the cluster of the UPC, we need to connect to the VPN beforehand. Explain more / omit? 
Separate in another subsection... also explaining where is hosted the gdb and the details of the VM of the UPC cluster...

- Run the golang program in the `cypherimport` subdirectory. It automatically runs the cypher commands needed for the populating process. **Note:** in this case we need to provide the path of the csv files *csvPath* argument on the local machine on which we are running this script.

```
$> go run main.go <csvPath>
```

# Verification

To simply check that the Neo4j graph database instance was successfuly populated, we can simply open it and execute the following cypher command:

```
MATCH (n) RETURN n
```

This should return all the nodes and relations that were created with the `populatemodule` program.


