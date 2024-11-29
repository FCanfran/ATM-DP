import diefpy
import pandas as pd  # for displaying the data in a nice way
import matplotlib.pyplot as plt
import sys
import os

# Testing script of the diefpy library

# 4.1. Visualizing Execution Time of SPARQL Query Engines

COLORS = ["#ECC30B", "#D56062", "#84BCDA"]

if len(sys.argv) < 2:
    print("Error, run like: $>python test.py resultsDirectoryPath")
    exit(1)

# Read name of the directory
input_dir = sys.argv[1]

metrics_all = []
header_metrics = False

# Subdirectories
for subdir in sorted(os.listdir(input_dir)):
    if os.path.isdir(os.path.join(input_dir, subdir)):
        print(subdir)
        metrics_file = os.path.join(input_dir, subdir, "metrics.csv")
        print(metrics_file)

        if os.path.isfile(metrics_file):  # Check if 'metrics.csv' exists
            try:
                # Read the CSV file
                if not header_metrics:
                    # Read with the header for the first file
                    df = pd.read_csv(metrics_file)
                    header_metrics = True
                else:
                    # Skip the header for subsequent files
                    df = pd.read_csv(metrics_file, header=0)

                print(df)
                metrics_all.append(df)
                # print(metrics_all)
            except Exception as e:
                print(f"Error reading {metrics_file}: {e}")

if metrics_all:
    metrics_all_df = pd.concat(metrics_all, ignore_index=True)
    # output metrics all csv
    output_metrics = input_dir + "metrics.csv"
    metrics_all_df.to_csv(output_metrics, index=False, header=header_metrics)
    print(f"Combined CSV saved to {output_metrics}")
else:
    print("No metrics.csv files found to combine.")


# same for traces
metrics_all = []
header_metrics = False

# Subdirectories
for subdir in sorted(os.listdir(input_dir)):
    if os.path.isdir(os.path.join(input_dir, subdir)):
        print(subdir)
        metrics_file = os.path.join(input_dir, subdir, "trace.csv")
        print(metrics_file)

        if os.path.isfile(metrics_file):  # Check if 'metrics.csv' exists
            try:
                # Read the CSV file
                if not header_metrics:
                    # Read with the header for the first file
                    df = pd.read_csv(metrics_file)
                    header_metrics = True
                else:
                    # Skip the header for subsequent files
                    df = pd.read_csv(metrics_file, header=0)

                print(df)
                metrics_all.append(df)
                # print(metrics_all)
            except Exception as e:
                print(f"Error reading {metrics_file}: {e}")

if metrics_all:
    metrics_all_df = pd.concat(metrics_all, ignore_index=True)
    # output metrics all csv
    output_metrics = input_dir + "trace.csv"
    metrics_all_df.to_csv(output_metrics, index=False, header=header_metrics)
    print(f"Combined CSV saved to {output_metrics}")
else:
    print("No traces.csv files found to combine.")

############################ RESULT PLOTS AND METRICS ############################

outputPlotDir = input_dir + "plots/"
if not os.path.exists(outputPlotDir):
    os.makedirs(outputPlotDir)

traces = diefpy.load_trace(input_dir + "trace.csv")

# Plot the answer trace recorded in `traces` for query `Q9.sparql`
diefpy.plot_answer_trace(traces, "dp", COLORS)
plt.savefig(outputPlotDir + "traces.png")

# computing dief@t until the time unit 10s
# dt = diefpy.dieft(traces, "dp", 10)
# print(pd.DataFrame(dt).head())

# computing dief@t until the time unit when the slowest approach finalizes its execution
# (in toy-trace.csv is t=~22.69s)
dt = diefpy.dieft(traces, "dp")
print(pd.DataFrame(dt).head())

print(input_dir + "metrics.csv")
metrics = diefpy.load_metrics(input_dir + "metrics.csv")

# Execution time plot
diefpy.plot_execution_time(metrics, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "execTime.png")

# Create all metrics from the `traces` and `metrics`
# computes the results reported in the previously mentioned experiment, i.e.,
# - dief@t
# - time to first tuple (tfft)
# - execution time (totaltime)
# - number of answers produced (comp?)
# - throughput (= comp/totaltime)
# - inverse time to first tuple (1/tfft)
# - inverse execution time (1/totaltime)
exp1 = diefpy.performance_of_approaches_with_dieft(traces, metrics)

print(pd.DataFrame(exp1[exp1["test"] == "dp"]).head())

# Create radar plot to compare the performance of the approaches with dief@t and other metrics.
# - Plot interpretation: Higher is better.
diefpy.plot_performance_of_approaches_with_dieft(exp1, "dp", COLORS)
plt.savefig(outputPlotDir + "radar-dieft.png")


# dief@k:
# The metric dief@k measures the diefficiency of a query engine while producing
# the first k answers when executing a query. Intuitively, approaches that require
# a shorter period of time to produce a certain number of answers are more efficient
# dief@k interpretation: Lower is better

# dief@k producing the first 5 answers
dk = diefpy.diefk(traces, "dp", 5)
print(pd.DataFrame(dk).head())

# dief@k producing the first 10 answers
dk = diefpy.diefk(traces, "dp", 10)
print(pd.DataFrame(dk).head())

# producing 50% of the answers
dk = diefpy.diefk2(traces, "dp", 0.50)
print(pd.DataFrame(dk).head())

# 4.6. Measuring dief@t at Different Answer Completeness Percentages
# compares the performance of the three variants when producing different
# answer completeness percentages (25%, 50%, 75%, 100%) using dief@k.

# method diefpy.continuous_efficiency_with_diefk computes the dief@k metric
# for the previously mentioned answer completeness percentages.
# Plot interpretation: Lower is better.
exp2 = diefpy.continuous_efficiency_with_diefk(traces)
diefpy.plot_continuous_efficiency_with_diefk(exp2, "dp", COLORS)
plt.savefig(outputPlotDir + "radar-diefk.png")
