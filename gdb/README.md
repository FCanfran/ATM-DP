# Synthetic Bank Data Generation Tool

This tool contains a synthetic bank dataset generation tool based on the [Wisabi Bank Dataset](https://www.kaggle.com/datasets/obinnaiheanachor/wisabi-bank-dataset?resource=download), a Golang population module for the creation of the corresponding Neo4j database. It also contains a parametrizable synthetic transaction generator based on the created bank dataset.

## Contents

List of files:

- `bankDataGenerator.py`: stable bank database data generator.
- `behavior.py`: program to generate the `behavior.csv` file, where the beahvior metrics of each of the
customers of the wisabi dataset are gathered.
- `txGenerator.py`: regular and anomalous transactions generator.
- `txGenerator-simplified.py`: simplified version of the regular and anomalous transactions generator.
- `txGenerator-split.py`: auxiliary script for doing the splitting of the transactions in interaction-start and interaction-end. Generates the start and end interaction for each transaction.
- `txGenerator-join-split.py`: auxiliary script for joining multiple different transaction files (ordering them by timestamp) and performing the splitting of each transaction into its corresponding interaction-start and interaction-end interactions.
- `csv`: directory with some generated bank data and transactions in csv format.
- `populatemodule`: golang module for the population of the stable bank database in Neo4j.
- `wisabi`: directory with the source csv files of the wisabi synthetic bank dataset.

## Synthetic Bank Database Generator

### 1. Creation process of the Bank (Graph) Database

For simplicity and to do it in a more stepwise manner, first all the CSV data tables for the nodes and for the relations are created in the corresponding format and then with those CSV the Neo4j GDB is populated. 

- 1. Ensure to have a wisabi named directory with the csv files of the Wisabi Bank
Dataset (publicly available on Kaggle [Wisabi Bank Dataset](https://www.kaggle.com/datasets/obinnaiheanachor/wisabi-bank-dataset?resource=download)).

- 2. Ensure to have the `behavior.csv` file or run $> python3 behavior.py to create
it. This creates a csv file with the gathered customer behavior properties from this
dataset. Place it inside the `wisabi` directory.

- 3. Run `$> python3 bankDataGenerator.py` and introduce:
   - (a) Bank properties’ values.
   - (b) n = |ATM|, internal and external.
   - (c) m = |Cards|.

### 2. Population of the Neo4j Graph Database

Prior to the population of the Neo4j graph database, a Neo4j graph database instance needs to be
created. This can be done either locally or in the cloud. See the following links to get more information on how to do this:

- Locally: [Neo4j Desktop](https://neo4j.com/docs/desktop-manual/current/)
- Cloud: [AuraDB](https://neo4j.com/cloud/platform/aura-graph-database/?ref=developer-guides)

More hands-on tutorial links:
- [Installation of Neo4j in Ubuntu 22.04](https://www.virtono.com/community/tutorial-how-to/how-to-install-neo4j-on-ubuntu-22-04/ )

Once we have a Neo4j graph database instance available, we can proceed to the population process. The explanation of this process can be seen in the `populatemodule` subdirectory.

## Synthetic Transaction Stream Generator

It is implemented as a Python program `txGenerator.py`. On it we need to specify the value of the parameters needed to customize the generation of the stream of transactions. 
These parameters are:

### Customizable Parameters for Transaction Stream Generation

| **Parameter**                          | **Description** |
|----------------------------------------|----------------------------------------------------------------------------------------------------------------------------------|
| `START_DATE`                           | Start date (in date format: `"YYYY-MM-DD"`) |
| `NUM_DAYS`                             | Number of days duration of the transaction stream generated |
| `MAX_DISTANCE_SUBSET_THRESHOLD`        | Maximum allowed distance (in km) of the ATMs in the ATM subset to the client location residence |
| `MAX_SIZE_ATM_SUBSET_RATIO`            | Ratio ∈ [0,1] to limit the maximum size of the ATM subset from which regular transactions of a card are linked to. So that:  \|`ATM_subset`\| = `MAX_SIZE_ATM_SUBSET_RATIO` * \|`ATM`\| |
| `MAX_DURATION`                         | Maximum time (in seconds) duration of a transaction |
| `MEAN_DURATION`                        | Mean duration (in seconds) duration of a transaction |
| `STD_DURATION`                         | Standard deviation (in seconds) duration of a transaction |
| `REGULAR_SPEED`                        | Assumed as the normal speed (in km/h) at which a client normally can travel the distance between two geographical points |
| `ANOMALOUS_RATIO_1`                    | Ratio ∈ [0,1] of anomalous transactions of the fraud pattern I over the total number of regular transactions for each of the cards. |
| `ANOMALOUS_SPEED`                      | Assumption on the maximum speed (in km/h) at which the distance between two geographical points can be traveled, for the generation of anomalous transactions |
| `ANOMALOUS_TX_DURATION`                | Duration (in seconds) of an anomalous transaction |

*Table: Description of the customizable parameters for the transaction stream generation*

### Usage

To use it:

1. Ensure to have a `csv` named directory with the *csv* stable bank dataset files on which we want to simulate a transaction stream (use the bank data generator `bankDataGenerator.py` to produce it).
2. Run the following command:

   ```bash
   $ python3 txGenerator.py <outputFileName>
   ```
   introducing *outputFileName* as an argument to name the transaction stream dataset files to be generated.

The program generates a `tx` directory with the *csv* files representing the transaction stream dataset:

```
<outputFileName>-all.csv       # Joint regular and anomalous dataset
<outputFileName>-regular.csv    # Regular transaction dataset
<outputFileName>-anomalous.csv  # Anomalous transaction dataset
```

### Simplified version

Finally, a simplified version of this synthetic stream generator was developed in the Python program `txGenerator-simplified.py`.  
In this version, the `ATM_subset` is built from a random selection of ATMs of the bank network, and not based on the distance to the residence location of the cardholder.  
This results in a faster generator, since it reduces the time complexity of the generation.

To use it:
- Run `txGenerator-simplified.py`.

Optionally, afterwards:
- Run `txGenerator-split.py` to split the generated transactions in interaction-start and interaction-end interactions.
- Run `txGenerator-join-split.py` to join multiple different transaction files and to split them in interaction-start and interaction-end interactions.  


### Optimized Version: `txGeneratorOptimized.py`

Is a version of the original / base program so that the generation process is faster. The main idea is to avoid creating one different `ATM_subset` for each of the cards, and instead create only a certain limited number of `ATM_subset`'s that will be shared among subsets of cards. 

Adjust the parameter:

- `CARD_CHUNK_SIZE`: indicating the size/number of the set of cards that will share a same `ATM_subset`.

