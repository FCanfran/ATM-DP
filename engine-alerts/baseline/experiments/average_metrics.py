import pandas as pd
import sys

if len(sys.argv) < 2:
    print("Error, run like: $>python average_traces.py <csvFile>")
    exit(1)

df = pd.read_csv(sys.argv[1])

grouped = df.groupby(["test", "approach", "comp"], as_index=False)[
    ["tfft", "totaltime"]
].mean()

grouped["tfft"] = grouped["tfft"].round(2)
grouped["totaltime"] = grouped["totaltime"].round(2)

grouped = grouped[["test", "approach", "tfft", "totaltime", "comp"]]

grouped.to_csv(sys.argv[1], index=False)
