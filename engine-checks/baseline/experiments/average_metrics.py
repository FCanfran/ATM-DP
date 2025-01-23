import pandas as pd
import sys

if len(sys.argv) < 2:
    print("Error, run like: $>python average_metrics.py <csvFile>")
    exit(1)

df = pd.read_csv(sys.argv[1])

grouped = df.groupby(["test", "approach"], as_index=False)[
    ["tfft", "totaltime", "mrt", "checks", "alerts"]
].mean()

grouped["tfft"] = grouped["tfft"].astype(int)
grouped["totaltime"] = grouped["totaltime"].round(2)
grouped["mrt"] = grouped["mrt"].astype(int)
grouped["checks"] = grouped["checks"].astype(int)
grouped["alerts"] = grouped["alerts"].astype(int)


grouped = grouped[["test", "approach", "tfft", "totaltime", "mrt", "checks", "alerts"]]

grouped.to_csv(sys.argv[1], index=False)
