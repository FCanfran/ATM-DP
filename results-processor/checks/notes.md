


# By fixed number of cores

- `results-fix-cores-main.sh`: runs for each of the specified core values labels, the program `results-fix-cores-a.sh`.

- `results-fix-cores-a.sh`: run of the `dieffpy-cores.py` for the specified core value experiment.


## `Dieffpy-cores.py` 

### Plots

#### `execTime.png`

#### `mrt.png`

#### `radar-diefk.png`

#### `radar-dieft.png`

Modified so to:

- Replace Checks by inverse mrt (mrt^-1) -> it is more informative, since
the number of checks it is expected to be the same in all the cases.

#### `traces-response-time-reduced.png`

- Transform the unit of the responseTime field values: from ns to ms (after custom load_trace - traces_response_time)

#### `traces-response-time.png`

- Transform the unit of the responseTime field values: from ns to ms (after custom load_trace - traces_response_time)


#### `traces.png`

### Results text file: `dieffpy-out.txt`

Modified so to:

- include checks (instead of labeling with comp)
- include mrt
- detail of the units of tfft (ms), mrt (ms) and totaltime (s).

- switch the diefk outputs: 
    - 5 -> 500
    - 10 -> 1000 first answers

TODO: For the ongoing 
- add the alerts values

# Usage

## By cores

Obtain processed results and plots by cores (for each of the indicated number of cores variations):
```
$> ./results-fix-cores-main.sh <results-directory> <testName>
```

## By filters 

```
$> ./results-fix-filters-main.sh <results-directory> <testName>
```

## Combined

```
$> ./results-combined.sh <resultsDirectoryPath> <TEST(name)> <DO_JOIN(0:no,1:yes)> <num_interactions>"
```

- `num_interactions`: refers to the number of interactions (openings and closings of transactions), therefore `num_tx` x 2.

