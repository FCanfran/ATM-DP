import pandas as pd
import matplotlib.pyplot as plt
import sys

labels = ["apache-arrow-1", "apache-arrow-2", "csv-encoding"]

if len(sys.argv) < 5:
    print("Usage: $> python createResultPlot f1 f2 f3 numRowsFile(power of 10)")
    exit(1)

file_names = sys.argv[1], sys.argv[2], sys.argv[3]
file_rows = int(sys.argv[4])

i = 0

for file_name in file_names:
    df = pd.read_csv(file_name, header=None, names=["chunkSize", "ms"])
    averages = df.groupby("chunkSize")["ms"].mean().reset_index()
    print(averages)
    plt.plot(averages["chunkSize"], averages["ms"], marker="o", label=labels[i])
    i += 1


plt.xlabel("Chunk size (number of rows)")
plt.ylabel("Time (ms)")
title = f"Total time to read (ms) vs chunk size - file of $10^{file_rows}$ rows"
plt.title(title)
plt.grid(True)
plt.legend(title="Variants")
plt.xscale("log")  # Logarithmic scale if chunk sizes are exponentially distributed

outfilename = f"10-{file_rows}.png"
plt.savefig(outfilename, format="png", dpi=300, bbox_inches="tight")

plt.show()
