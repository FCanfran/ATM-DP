
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

### Card

card.csv:

```
number_id,client_id,expiration,CVC,loc_latitude,loc_longitude,extract_limit,amount_avg,amount_std,withdrawal_day
LAGO-0,0,2024-07-03,999,9.002378,7.581718,126556.6,25311.32,28105.03,0.2904
LAGO-1,1,2024-07-03,999,9.09826,7.602691,108483.4,21696.68,23203.51,0.5781
KANO-0,0,2024-07-03,999,9.171933,7.389227,113673.45,22734.69,22709.98,0.2685
KANO-1,1,2024-07-03,999,10.733811,7.876416,119193.55,23838.71,23348.16,0.2548
ABYA-0,0,2024-07-03,999,9.002271,7.63722,109608.7,21921.74,23508.11,0.6301
ABYA-1,1,2024-07-03,999,9.105015,7.383571,102427.2,20485.44,21348.12,0.2822
```

```
LOAD CSV WITH HEADERS FROM 'file:///csv/card.csv' AS row
MERGE (c:Card {number_id: row.number_id, client_id: row.client_id, expiration: date(row.expiration), CVC: toInteger(row.CVC), extract_limit: toFloat(row.extract_limit), loc_latitude: toFloat(row.loc_latitude), loc_longitude: toFloat(row.loc_longitude)});
``` 

Notes:

- We do not include the fields that were generated to define the behavior of the card. They are only used for the generation of the transactions: `amount_avg`, `amount_std`,`withdrawal_day`.
- `expiration`: set as date type.

**Possible improvements:**

- `CVC`: set as integer data type, although it could be set as string, and probably will occupy less space this way.
- `extract_limit`: set as float data type, although it could be rounded to be set as integer and therefore occupy less space.

**A change:**
- The card identifier: followed the same structure as the ATM ids. It is changed to make it more clear the difference between these two nodes apart from the label.

### Card-Bank relationships

card-bank.csv:

```
code,number_id
LAGO,LAGO-0
LAGO,LAGO-1
KANO,KANO-0
KANO,KANO-1
ABYA,ABYA-0
ABYA,ABYA-1
```

```
LOAD CSV WITH HEADERS FROM 'file:///csv/card-bank.csv' AS row
             MATCH (c:Card {number_id: row.number_id})
             MATCH (b:Bank {code: row.code})
             MERGE (c)-[r:ISSUED_BY]->(b);
```

Notes:
- The edges are labeled as `ISSUED_BY`.

## Adding uniqueness constraints

It happens that, so far, when adding for example into the database at has the same ID as other already existing in the database, that node is also merged withouth problem. 
We would like to avoid this, so that we have a uniquess constraint on the IDs of each of the kind of nodes. Therefore, when adding a new ATM node that has the same ID as another ATM already existing in the database, we are aware of this and we do not let this insertion to happen.

Also, it appears as recommendation to do this: *"Always create uniqueness constraints prior to importing data, to avoid duplicates or colliding entities. If the source file contains duplicated data and the right constraints are in place, Cypher raises an error."*

### ATM

Create node property uniqueness constraints on `ATM_id`:

```
CREATE CONSTRAINT ATM_id IF NOT EXISTS
FOR (a:ATM) REQUIRE a.ATM_id IS UNIQUE
```

### Card

Create node property uniqueness constraints on Card `number_id`:

```
CREATE CONSTRAINT number_id IF NOT EXISTS
FOR (c:Card) REQUIRE c.number_id IS UNIQUE
```

### Bank

Create node property uniqueness constraints on Bank `code`:

```
CREATE CONSTRAINT code IF NOT EXISTS
FOR (b:Bank) REQUIRE b.code IS UNIQUE
```