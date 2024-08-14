# Continuous Query Engine for the Detection of Anomalous ATM transactions

In the present we can find:

- `gdb`: graph database generator. It contains a synthetic bank dataset generation tool based on the [Wisabi Bank Dataset](https://www.kaggle.com/datasets/obinnaiheanachor/wisabi-bank-dataset?resource=download) and 
a Golang population module for the creation of the corresponding Neo4j database.

- `pipeline`: It contains a Golang first prototype implementation of the Dynamic Pipeline Query Engine for the detection of anomalous ATM transactions (DP-ATM).