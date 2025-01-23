
List of files:

- `bankDataGenerator.py`: stable bank database data generator.
- `behavior.py`: program to generate the `behavior.csv` file, where the beahvior metrics of each of the
customers of the wisabi dataset are gathered.
- `txGenerator.py`: regular and anomalous transactions generator.
- `txGenerator-simplified.py`: simplified version of the regular and anomalous transactions generator.
- `csv`: directory with some generated bank data and transactions in csv format.
- `populatemodule`: golang module for the population of the stable bank database in Neo4j.
- `wisabi`: directory with the source csv files of the wisabi synthetic bank dataset.

# 2. Creation process of the Bank (Graph) Database

For simplicity and to do it in a more stepwise manner, first all the CSV data tables for the nodes and for the relations are created in the corresponding format and then with those CSV the Neo4j GDB is populated. 
Do this using the `bankDataGenerator.py` script.

# 2. Population of the Neo4j Graph Database

Prior to the population of the Neo4j graph database, a Neo4j graph database instance needs to be
created. This can be done either locally or in the cloud. See the following links to get more information on how to do this:

- Locally: [Neo4j Desktop](https://neo4j.com/docs/desktop-manual/current/)
- Cloud: [AuraDB](https://neo4j.com/cloud/platform/aura-graph-database/?ref=developer-guides)

More hands-on tutorial links:
- [Installation of Neo4j in Ubuntu 22.04](https://www.virtono.com/community/tutorial-how-to/how-to-install-neo4j-on-ubuntu-22-04/ )

Once we have a Neo4j graph database instance available, we can proceed to the population process. The explanation of this process can be seen in the `populatemodule` subdirectory.
