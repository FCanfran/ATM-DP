


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

TODO: For the ongoing 
- add the alerts values






