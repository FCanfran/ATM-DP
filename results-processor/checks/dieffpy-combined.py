import diefpy
import pandas as pd
import matplotlib.pyplot as plt
import sys
import os
import re


#########################################################################################
# edit: to show approaches listed respecting numerical order in the plots
# OTHERWISE: comment these lines and use the diefpy corresponding function call
import matplotlib.lines as mlines
import matplotlib.ticker as mticker
import numpy as np
from matplotlib.figure import Figure

from diefpy.radaraxes import radar_factory


DEFAULT_COLORS = ("#ECC30B", "#D56062", "#84BCDA")
"""Default colors for printing plots: yellow, red, blue"""

COLORS = (
    "#ECC30B",  # Yellow
    "#D56062",  # Coral
    "#84BCDA",  # Light Blue
    "#F5A623",  # Orange
    "#7ED321",  # Green
    "#4A90E2",  # Blue
    "#9013FE",  # Purple
    "#50E3C2",  # Teal
    "#F8E71C",  # Light Yellow
    "#B8E986",  # Light Green
    "#D0021B",  # Red
    "#8B572A",  # Brown
    "#9B9B9B",  # Gray
    "#F2C94C",  # Mustard
    "#E94E77",  # Pink
    "#56CCF2",  # Sky Blue
    "#FF477E",  # Hot Pink
    "#35A7FF",  # Sky Blue
    "#F9AFAE",  # Light Coral
    "#21C28E",  # Mint Green
    "#C8E8C1",  # Pale Green
    "#F1A7C7",  # Soft Pink
)


####################################################################################################################
def load_metrics(filename: str) -> np.ndarray:
    """
    Reads the other metrics from a CSV file.

    Conventional query performance measurements.
    The attribues of the file specified in the header are expected to be:

    * *test*: the name of the executed test
    * *approach*: the name of the approach executed
    * *tfft*: time elapsed until the first answer was generated
    * *totaltime*: time elapsed until the last answer was generated
    * *mrt*: mean response time
    * *checks*: number of checks produced
    * *alerts*: number of alerts produced

    :param filename: Path to the CSV file that contains the other metrics.
                     Attributes of the file specified in the header: test, approach, tfft, totaltime, comp.
    :return: Dataframe with the other metrics. Attributes of the dataframe: test, approach, tfft, totaltime, comp.

    **Examples**

    >>> load_trace("data/metrics.csv")
    """
    # Loading data.
    # names=True is not an error, it is valid for reading the column names from the data
    if np.__version__ >= "1.23.0":
        df = np.genfromtxt(
            filename, delimiter=",", names=True, dtype=None, encoding="utf8", ndmin=1
        )
    else:
        df = np.genfromtxt(
            filename, delimiter=",", names=True, dtype=None, encoding="utf8"
        )

    # Return dataframe in order.
    return df[["test", "approach", "tfft", "totaltime", "mrt", "checks", "alerts"]]


def load_trace_reduced(filename: str) -> np.ndarray:
    """
    Reads answer traces from a CSV file.

    Answer traces record the points in time when an approach produces an answer.
    The attribues of the file specified in the header are expected to be:

    * *test*: the name of the executed test
    * *approach*: the name of the approach executed
    * *answer*: the number of the answer produced
    * *time*: time elapsed from the start of the execution until the generation of the answer

    :param filename: Path to the CSV file that contains the answer traces.
                     Attributes of the file specified in the header: test, approach, answer, time.
    :return: Dataframe with the answer trace. Attributes of the dataframe: test, approach, answer, time.

    **Examples**

    >>> load_trace("data/traces.csv")
    """
    # Loading data.
    # names=True is not an error, it is valid for reading the column names from the data
    if np.__version__ >= "1.23.0":
        df = np.genfromtxt(
            filename, delimiter=",", names=True, dtype=None, encoding="utf8", ndmin=1
        )
    else:
        df = np.genfromtxt(
            filename, delimiter=",", names=True, dtype=None, encoding="utf8"
        )

    # F: Fix -> allow more characters for the test field
    new_dtype = []
    for field in df.dtype.descr:
        if field[0] == "test":  # Adjust the size of the 'test' field
            new_dtype.append((field[0], "U50"))
        else:  # Retain the original dtype for other fields
            new_dtype.append(field)

    # Convert the array to the new dtype
    df = np.array(df, dtype=new_dtype)

    # Return dataframe in order.
    return df[["test", "approach", "answer", "time"]]


def plot_execution_time_edit_1(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        # x = subset["filters"]
        x = range(len(subset))
        y = subset["totaltime"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel("Execution Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    filter_labels = metrics["filters"].unique()  # Get the unique filter values
    ax.set_xticks(range(len(filter_labels)))  # Set ticks at equal intervals
    ax.set_xticklabels(
        filter_labels, rotation=45
    )  # Use the actual filter values as tick labels

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_execution_time_edit_2(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        x = subset["filters"]
        y = subset["totaltime"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel("Execution Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_execution_time_edit_cores_1(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by filter
    unique_filters = metrics["filters"].unique()
    for i, filter_count in enumerate(unique_filters):
        subset = metrics[metrics["filters"] == filter_count]
        # x = subset["filters"]
        x = range(len(subset))
        y = subset["totaltime"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{filter_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# cores")
    ax.set_ylabel("Execution Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# filters", loc="best")

    core_labels = metrics["cores"].unique()  # Get the unique cores values
    ax.set_xticks(range(len(core_labels)))  # Set ticks at equal intervals
    ax.set_xticklabels(
        core_labels, rotation=45
    )  # Use the actual filter values as tick labels

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_execution_time_edit_cores_2(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by filter
    unique_filters = metrics["filters"].unique()
    for i, filter_count in enumerate(unique_filters):
        subset = metrics[metrics["filters"] == filter_count]
        x = subset["cores"]
        y = subset["totaltime"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{filter_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# cores")
    ax.set_ylabel("Execution Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# filters", loc="best")

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_mrt_edit_1(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        # x = subset["filters"]
        x = range(len(subset))
        y = subset["mrt"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel("Mean Response Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    filter_labels = metrics["filters"].unique()  # Get the unique filter values
    ax.set_xticks(range(len(filter_labels)))  # Set ticks at equal intervals
    ax.set_xticklabels(
        filter_labels, rotation=45
    )  # Use the actual filter values as tick labels

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_mrt_edit_2(
    test_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        x = subset["filters"]
        y = subset["mrt"]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel("Mean Response Time [s]")
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_edit_1(
    test_name,
    metric_name,
    label_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        # x = subset["filters"]
        x = range(len(subset))
        y = subset[metric_name]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel(label_name)
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    filter_labels = metrics["filters"].unique()  # Get the unique filter values
    ax.set_xticks(range(len(filter_labels)))  # Set ticks at equal intervals
    ax.set_xticklabels(
        filter_labels, rotation=45
    )  # Use the actual filter values as tick labels

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def plot_edit_2(
    test_name,
    metric_name,
    label_name,
    metrics: pd.DataFrame,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    # Create a figure and axis
    fig, ax = plt.subplots(figsize=(10, 6))

    # Group data by core
    unique_cores = metrics["cores"].unique()
    for i, core_count in enumerate(unique_cores):
        subset = metrics[metrics["cores"] == core_count]
        x = subset["filters"]
        y = subset[metric_name]
        color = colors[i % len(colors)]
        ax.plot(x, y, label=f"{core_count}", color=color, marker="o")

    if log_scale:
        ax.set_xscale("log")
        ax.set_yscale("log")

    ax.set_xlabel("# filters")
    ax.set_ylabel(label_name)
    ax.set_title(test_name)
    ax.legend(title="# cores", loc="best")

    ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


def dieft_edit(
    inputtrace: np.ndarray,
    inputtest: str,
    t: float = -1.0,
    continue_to_end: bool = True,
) -> np.ndarray:
    """
    Computes the **dief@t** metric for a specific test at a given time point *t*.

    **dief@t** measures the diefficiency during an exlapsed time period *t* by computing
    the area under the curve of the answer traces.
    By default, the function computes the maximum of the execution time among the approaches
    in the answer trace, i.e., until the point in time when the slowest approach finishes.

    :param inputtrace: Dataframe with the answer trace. Attributes of the dataframe: test, approach, answer, time.
    :param inputtest: Specifies the specific test to analyze from the answer trace.
    :param t: Point in time to compute dief@t for. By default, the function computes the maximum of the execution time
              among the approaches in the answer trace.
    :param continue_to_end: Indicates whether the AUC should be continued until the end of the time frame
    :return: Dataframe with the dief@t values for each approach. Attributes of the dataframe: test, approach, dieft.

    **Examples**

    >>> dieft(traces, "Q9.sparql")
    >>> dieft(traces, "Q9.sparql", 7.5)
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("test", inputtrace["test"].dtype),
            ("approach", inputtrace["approach"].dtype),
            ("dieft", float),
        ],
    )

    # Obtain test and approaches to compare.
    results = inputtrace
    approaches = np.unique(results["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    # Obtain maximum t over all approaches if t is not set.
    if t == -1:
        t = np.max(results["time"])

    # Compute dieft per approach.
    for a in sorted_approaches:
        dief = 0
        subtrace = results[(results["approach"] == a) & (results["time"] <= t)]

        if continue_to_end:
            com = np.array(
                [(inputtest, a, len(subtrace), t)],
                dtype=[
                    ("test", subtrace["test"].dtype),
                    ("approach", subtrace["approach"].dtype),
                    ("answer", int),
                    ("time", float),
                ],
            )

            if len(subtrace) == 1 and subtrace["answer"] == 0:
                pass
            else:
                subtrace = np.concatenate((subtrace, com), axis=0)

        if len(subtrace) > 1:
            dief = np.trapz(subtrace["answer"], subtrace["time"])

        res = np.array(
            [(inputtest, a, dief)],
            dtype=[
                ("test", subtrace["test"].dtype),
                ("approach", subtrace["approach"].dtype),
                ("dieft", float),
            ],
        )
        df = np.append(df, res, axis=0)

    return df


####################################################################################################################


if len(sys.argv) < 5:
    print(
        "Error, run like: $>python dieffpy.py resultsDirectoryPath TEST(name) DO_JOIN(0:no,1:yes) num_interactions"
    )
    exit(1)

# Read name of the directory
input_dir = sys.argv[1]
test_name = sys.argv[2]
do_join = sys.argv[3] == "1"
num_interactions = int(sys.argv[4])

if do_join:

    metrics_all = []
    header_metrics = False
    traces_all = []
    header_traces = False

    subdirs = [
        d for d in os.listdir(input_dir) if os.path.isdir(os.path.join(input_dir, d))
    ]

    # Sort subdirectories alphanumerically, respecting numerical order
    sorted_subdirs = sorted(
        subdirs,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    # Subdirectories
    for subdir in sorted_subdirs:
        if os.path.isdir(os.path.join(input_dir, subdir)):
            # print(subdir)
            metrics_file = os.path.join(input_dir, subdir, "metrics.csv")
            # print(metrics_file)

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

                    # print(df)
                    metrics_all.append(df)
                    # print(metrics_all)
                except Exception as e:
                    print(f"Error reading {metrics_file}: {e}")

            traces_file = os.path.join(input_dir, subdir, "trace.csv")
            # print(traces_file)
            if os.path.isfile(traces_file):
                try:
                    # Read the CSV file
                    if not header_traces:
                        # Read with the header for the first file
                        df = pd.read_csv(traces_file)
                        header_traces = True
                    else:
                        # Skip the header for subsequent files
                        df = pd.read_csv(traces_file, header=0)

                    traces_all.append(df)
                except Exception as e:
                    print(f"Error reading {traces_file}: {e}")

    if metrics_all:
        metrics_all_df = pd.concat(metrics_all, ignore_index=True)
        # output metrics all csv
        output_metrics = input_dir + "/metrics.csv"
        metrics_all_df.to_csv(output_metrics, index=False, header=header_metrics)
    else:
        print("No metrics.csv files found to combine.")

    if traces_all:
        traces_all_df = pd.concat(traces_all, ignore_index=True)
        # output metrics all csv
        output_traces = input_dir + "/trace.csv"
        traces_all_df.to_csv(output_traces, index=False, header=header_traces)
    else:
        print("No traces.csv files found to combine.")

############################ RESULT PLOTS AND METRICS ############################

outputPlotDir = input_dir + "/plots/"
if not os.path.exists(outputPlotDir):
    os.makedirs(outputPlotDir)


# total execution time plot
metrics = load_metrics(input_dir + "/metrics.csv")
# F: Conversions
# tfft (ns) -> ms
# mrt (ns)  -> s
metrics["mrt"] = metrics["mrt"] / 1_000_000_000
metrics["tfft"] = metrics["tfft"] / 1_000_000

columns = ["test", "approach", "tfft", "totaltime", "mrt", "checks", "alerts"]

df = pd.DataFrame(metrics, columns=columns)

# extract # cores and # filters
df["cores"] = df["test"].str.extract(r"-(\d+)c$").astype(int)
df["filters"] = df["approach"].str.extract(r"-(\d+)f$").astype(int)

plot_execution_time_edit_1(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "execTime-1.png")

plot_execution_time_edit_2(test_name, df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "execTime-log.png")

plot_execution_time_edit_2(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "execTime-2.png")

plot_execution_time_edit_cores_1(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "execTime-cores-1.png")

plot_execution_time_edit_cores_2(test_name, df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "execTime-cores-log.png")

plot_execution_time_edit_cores_2(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "execTime-cores-2.png")


# mean response time plot
plot_mrt_edit_1(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "mrt-1.png")

plot_mrt_edit_2(test_name, df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "mrt-log.png")

plot_mrt_edit_2(test_name, df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "mrt-2.png")


# tfft
# throughput: number of checks / second
df["throughput"] = df["checks"] / df["totaltime"]
# interactions/second
df["interactions/s"] = num_interactions / df["totaltime"]
# dieft
traces = load_trace_reduced(input_dir + "/trace.csv")
dt = dieft_edit(traces, test_name)
print("dief@t until the time unit when the slowest approach finalizes its execution")
print(
    "NOTE! - Now it is among all the possible included variations! (not only among the ones with the same # filters or # cores)"
)
df["dieft"] = dt["dieft"]

with pd.option_context("display.max_rows", None, "display.max_columns", None):
    print(pd.DataFrame(df))
print("____________________________________________________________________________")
print()

# plots

# tfft
plot_edit_1(test_name, "tfft", "tfft [ms]", df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "tfft-1.png")
plot_edit_2(test_name, "tfft", "tfft [ms]", df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "tfft-log.png")

# throughput: number of checks / second
plot_edit_1(
    test_name, "throughput", "throughput (checks/s)", df, COLORS, log_scale=False
)
plt.savefig(outputPlotDir + "throughput-1.png")
plot_edit_2(
    test_name, "throughput", "throughput (checks/s)", df, COLORS, log_scale=True
)
plt.savefig(outputPlotDir + "throughput-log.png")

# interactions/s
plot_edit_1(test_name, "interactions/s", "interactions/s", df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "interactions-1.png")
plot_edit_2(test_name, "interactions/s", "interactions/s", df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "interactions-log.png")

# dieft
plot_edit_1(test_name, "dieft", "dieft", df, COLORS, log_scale=False)
plt.savefig(outputPlotDir + "dieft-1.png")
plot_edit_2(test_name, "dieft", "dieft", df, COLORS, log_scale=True)
plt.savefig(outputPlotDir + "dieft-log.png")
