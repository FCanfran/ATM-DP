import pandas as pd
import sys

if len(sys.argv) < 2:
    print("Error, run like: $>python average_traces.py <csvFile>")
    exit(1)

df = pd.read_csv(sys.argv[1])

# Group by 'answer' and calculate the mean of 'time'
grouped = df.groupby(["test", "approach", "answer"], as_index=False)["time"].mean()

grouped["time"] = grouped["time"].round(2)

grouped.to_csv(sys.argv[1], index=False)
