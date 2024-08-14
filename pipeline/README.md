
# Dynamic Pipeline Engine (DP-ATM)

This module contains a first prototype implementation of the DP-ATM for detecting anomalous ATM transaction patterns.


The connection and queries of the the dynamic pipeline engine filters to the stable graph database will be done using similar methods as the ones used in the `populatemodule` of the graph database creation directory `gdb`. `SafeConnect()` to connect and `writeQuery()` to do the queries in cypher, but in this case specifying the corresponding needed queries to obtain additional information of the stable database so that the fraud pattern matching algorithms can be performed.

The connection with the corresponding Neo4j graph database instance is done using the `SafeConnect()` method that uses the specified URI, USERNAME and PASSWORD in the required `.env` file. The `.env` file needs to have the `NEO4J_URI`, `NEO4J_USERNAME` and `NEO4J_PASSWORD` and needs to be placed in the `cmd` directory, for example:
```
NEO4J_URI="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="xxxxxxx"
```

Once the connection is established, the cypher queries through our golang module will be done using the `writeQuery()` function.