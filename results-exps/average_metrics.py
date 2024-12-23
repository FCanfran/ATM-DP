import pandas as pd
import sys

if len(sys.argv) < 2:
    print("Error, run like: $>python average_traces.py <csvFile>")
    exit(1)


###################### TEST 23/12 #####################
# remove mrt_diff!!

df = pd.read_csv(sys.argv[1])

grouped = df.groupby(["test", "approach", "comp"], as_index=False)[
    ["tfft", "totaltime", "mrt", "mrt_diff"]
].mean()

grouped["tfft"] = grouped["tfft"].round(2)
grouped["totaltime"] = grouped["totaltime"].round(2)
grouped["mrt"] = grouped["mrt"].astype(int)

###################### TEST 23/12 #####################
grouped["mrt_diff"] = grouped["mrt_diff"].astype(int)
#######################################################

grouped = grouped[["test", "approach", "tfft", "totaltime", "mrt", "comp", "mrt_diff"]]

grouped.to_csv(sys.argv[1], index=False)
