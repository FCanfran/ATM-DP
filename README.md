
# Creation process of the Bank (Graph) Database

# 1. Generation of the synthetic data

Note that since for the population of the database, typically is seen that is done through:
- CSV tables imports.
- With scripts to create with cypher commands all the nodes, relations... all the database.

For simplicity and to do it in a more stepwise manner, we are going to first create all the CSV data tables for the nodes and for the relations in the corresponding format and then we will populate the Neo4j GDB with those.

## Generating the synthetic data

### Bank

- name: string
- code: int
- loc_latitude: float
- loc_longitude: float

* coordinates

### Card

- number_id: string * 
- client_id: string * 
- expiration: date
- CVC: int
- extract_limit: float * 
- loc_latitude: float * 
- loc_longitude: float *

expiration and CVC -> not relevant: could be empty fields indeed or for all the Cards the same values.

### ATM

- ATM_id: string
- loc_latitude: float
- loc_longitude: float
- city: string
- country: string