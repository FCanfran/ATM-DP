# Continuous Query Engine for the Detection of Anomalous ATM transactions

This repository contains an implementation of the Continuous Query Engine for the detection of Anomalous ATM transactions. This is a project done as part of my master's thesis at UPC (Universitat Politècnica de Catalunya), Barcelona, Spain.

The repository also contains a synthetic bank graph database generator, which includes a customizable synthetic transaction stream generator. Full detail description is given in the master's thesis document `report.pdf`.

## Contents

The contents of this repository are:

- `gdb`: graph database generator. It contains a synthetic bank dataset generation tool based on the [Wisabi Bank Dataset](https://www.kaggle.com/datasets/obinnaiheanachor/wisabi-bank-dataset?resource=download), a Golang population module for the creation of the corresponding Neo4j database. It also contains a parametrizable synthetic transaction generator based on the created bank dataset. More details of this tool are given inside this directory.

- `auxiliary-tests`: Some auxiliary tests of the project.

- `results-processor`: It contains scripts for the generation of the results plots of the experiments carried on the project. These scripts make use of the custom-modified version of the [diefpy tool](https://sdm-tib.github.io/diefpy/).

- `results`: It contains all the result plots and files for all the experiments carried along the project.

- `report.pdf`: The Master Thesis in pdf format.

Then we can find different versions of the Golang implementation of the Continuous Query Engine for the detection of anomalous ATM transactions (ATM-DP):

- `engine-E0`: ATM-DP with real-time simulation of the transaction input stream.

- `engine-E1-alerts`: ATM-DP with high-load stress simulation of the transaction input stream. Only positive fraud checks (alerts) are processed as results of the engine.

- `engine-E1-checks`: ATM-DP with high-load stress simulation of the transaction input stream. All fraud checks are processed as results of the engine.

## Usage

To run any of these variants execute the `main.go` inside the corresponding `cmd` subdirectory of the desired version:

```
>$ go run main.go <executionDescriptionFile>
```

where you need to specify the details on the paramteres on the execution on a `csv` file `<executionDescriptionFile>` and set up the details of the connection to the Neo4j graph database through a `.env` file. Regarding the configuration of the `<executionDescriptionFile>` it has to follow this same format:

```
txFile,test,approach,maxFilterSize
../input/medium/7-0.03.csv,7-0.03-8c,8c-10f,50000
```

On it we need to indicate:

- `txFile`: Path to the input transactions stream `csv` file to be used.
- `test`: Name of the performed test - you can put the name you want.
- `approach`: Name of the performed approach - you can put the name you want (e.g. we indicate the number of cores with which we execute and the number of filters)
- `maxFilterSize`: Value of the parameter that sets up the maximum capacity (in number of cards) of a filter.


With respect to the `.env` file; we need to specify the connection credentials to the Neo4j graph database that is going to be used:

```
NEO4J_URI="bolt://localhost:7687"
NEO4J_USERNAME="neo4j"
NEO4J_PASSWORD="xxxxx"
```


The execution will generate an `output` directory with some files related with the output of the experiment:
- `alerts.txt`: it contains a register of the emitted alerts.
- `metrics.csv`: metrics summary of the execution.
- `out-log.txt`: auxiliary output events emitted by the system.
- `txLog.txt`: log file of the input transactions.
- `trace.csv`: trace of all the emitted answers/results; it is a `csv` file with this format:

```
test,approach,answer,time,responseTime,isPositive  
7-0.03-8c,8c-10f,1,0.30,38915722,0
7-0.03-8c,8c-10f,2,0.32,52218740,0
```

such that for every answer it shows the time at which it was emitted `time`, its response time `responseTime` and a label `isPositive` indicating whether the answer is an alert (1) or not (0). This last parameter is only shown in the trace file of the engine versions that consider all checks as results and not only the alerts.

## Contact 

For any doubt on the repository, feel free to contact at fernando.martin.canfran@estudiantat.upc.edu
