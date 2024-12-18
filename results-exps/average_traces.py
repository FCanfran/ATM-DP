import pandas as pd
import sys

if len(sys.argv) < 2:
    print("Error, run like: $>python average_traces.py <csvFile>")
    exit(1)

df = pd.read_csv(sys.argv[1])

# grouped = df.groupby(["test", "approach", "answer"], as_index=False)["time"].mean()

grouped = df.groupby(["test", "approach", "answer"], as_index=False).agg(
    {"time": "mean", "responseTime": "mean"}
)

grouped["time"] = grouped["time"].round(2)
grouped["responseTime"] = grouped["responseTime"].round(2)

grouped.to_csv(sys.argv[1], index=False)
