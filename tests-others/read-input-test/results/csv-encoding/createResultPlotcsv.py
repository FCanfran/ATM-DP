import pandas as pd
import matplotlib.pyplot as plt
import sys


labels = ["worker (chunkSize=$10^{2}$)", "no-worker (no chunks)"]

if len(sys.argv) < 3:
    print("Usage: $> python createResultPlot f1 f2")
    exit(1)

file_names = sys.argv[1], sys.argv[2]

i = 0

for file_name in file_names:
    df = pd.read_csv(file_name, header=None, names=["numRows", "ms"])
    averages = df.groupby("numRows")["ms"].mean().reset_index()
    print(averages)
    plt.plot(averages["numRows"], averages["ms"], marker="o", label=labels[i])
    i += 1


plt.xlabel("Number of rows")
plt.ylabel("Time (ms)")
title = f"Total time to read (ms) vs csv-encoding type (worker with chunks / no-worker without chunks)"
plt.title(title)
plt.grid(True)
plt.legend(title="Variants")
plt.xscale("log")  # Logarithmic scale

plt.savefig("csv-encoding.png", format="png", dpi=300, bbox_inches="tight")
###
plt.show()
