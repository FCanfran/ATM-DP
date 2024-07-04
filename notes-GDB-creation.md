
Use of CSV import clause of cypher. See documentation [here](https://neo4j.com/docs/cypher-manual/5/clauses/load-csv/).

- **Requirement**: Place all the CSVs to import under the `/var/lib/neo4j/import` directory.
- **Note that:** It is better to import the corresponding data types as they are and not as 
strings. This is supported by NEo4j and it is better since it allows to query more effectively and to process it with type-specific Cypher query functions.
- **[Neo4j spatial functions](https://neo4j.com/docs/cypher-manual/current/functions/spatial/)** 
- **[Point data type](https://neo4j.com/docs/api/python-driver/current/types/spatial.html)**: it can be really useful to make use of pre-built spatial cypher functions such as distance function between 2 points.

Some notes on the nodes and relationships naming conventions: [here](https://neo4j.com/docs/cypher-manual/current/syntax/naming/)

- Node labels: Singular form and CamelCase -> "CustomerOrder"
- Relationship labels: Uppercase with underscores -> "HAS_ORDER"
- Properties: lowercase with underscores -> "first_name"

### ATMs

```
ATM_id,loc_latitude,loc_longitude,city,country
LAGO-0,11.693922,8.472546,Kano,Nigeria
LAGO-1,4.891065,7.400862,Oyigbo,Nigeria
LAGO-2,9.050378,7.470354,Abuja,Nigeria
```

```
LOAD CSV WITH HEADERS FROM 'file:///csv/atm.csv' AS row
MERGE (a:ATM {ATM_id: row.ATM_id, loc_latitude: toFloat(row.loc_latitude), loc_longitude: toFloat(row.loc_longitude), city: row.city, country: row.country});
``` 

Note that:
- ATM is the node label, the rest are the properties of this kind of node.
- Latitude and longitude stored as float values; note that they could also be stored
as Point cypher data type. However for the moment it is left like this. In the future
it could be converted when querying or directly be set as cypher point data type as property.


### Bank

```
name,code,loc_latitude,loc_longitude
Lagos Bank,LAGO,6.478685,3.368442
Kano Bank,KANO,11.994949,8.520313
Abuya Bank,ABYA,9.042977,7.478564
```

```
LOAD CSV WITH HEADERS FROM 'file:///csv/bank.csv' AS row
MERGE (b:Bank {name: row.name, code: row.code, loc_latitude: toFloat(row.loc_latitude), loc_longitude: toFloat(row.loc_longitude)});
``` 

Notes:

- The `code` is stored as a string and not as an integer, since to make it more clear it 
was already generated as a string code name.

### ATM-Bank relationships

atm-bank.csv:

```
code,ATM_id
LAGO,LAGO-0
LAGO,LAGO-1
LAGO,LAGO-2
```

```
LOAD CSV WITH HEADERS FROM 'file:///csv/atm-bank.csv' AS row
             MATCH (a:ATM {ATM_id: row.ATM_id})
             MATCH (b:Bank {code: row.code})
             MERGE (a)-[r:BELONGS_TO]->(b);
```

Notes:
- The edges are labeled as `BELONGS_TO`.