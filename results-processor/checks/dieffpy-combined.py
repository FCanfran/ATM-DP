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
    * *comp*: number of answers produced

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
    return df[["test", "approach", "tfft", "totaltime", "mrt", "comp"]]


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

    # ax.grid(True, which="both", linestyle="--", linewidth=0.5)
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

    # ax.grid(True, which="both", linestyle="--", linewidth=0.5)
    plt.tight_layout()

    return fig


# mrt is read in ns -> we plot it in ms
def plot_mean_response_time_single_test(
    test_name,
    metrics: np.ndarray,
    colors: list = DEFAULT_COLORS,
    log_scale: bool = False,
) -> Figure:

    submetrics = metrics[metrics["test"] == test_name]
    approaches = np.unique(metrics["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    color_map = dict(zip(sorted_approaches, colors))

    results = [
        submetrics[submetrics["approach"] == a]["mrt"][0] for a in sorted_approaches
    ]

    edited_labels = [a.split("-")[-1] for a in sorted_approaches]
    edited_labels = [re.search(r"\d+", label).group() for label in edited_labels]

    fig = plt.figure(figsize=(0.6 * len(approaches), 5), dpi=100)

    # Plot each bar with its respective label
    for approach, result, color, label in zip(
        sorted_approaches,
        results,
        [color_map[a] for a in sorted_approaches],
        edited_labels,
    ):
        plt.bar(approach, result, color=color, label=label, width=0.7)

    # Customizing the chart
    plt.xlabel("# filters", fontsize="large", labelpad=10)
    plt.ylabel("Mean Response Time [ms]", fontsize="large")
    plt.xticks(
        range(len(sorted_approaches)), edited_labels, rotation=90, fontsize="medium"
    )
    plt.legend(
        edited_labels,
        bbox_to_anchor=(1, 1),
        loc="upper left",
        labelspacing=0.1,
        fontsize="medium",
        frameon=False,
        title="# filters",
    )

    title = test_name.split("-")[-1]

    plt.title(f"{title}", fontsize=16, loc="center", pad=10)
    if log_scale:
        plt.yscale("log")

    # Display the chart
    plt.tight_layout()

    return fig


def plot_performance_of_approaches_with_dieft_edit(
    allmetrics: np.ndarray, q: str, colors: list = DEFAULT_COLORS
) -> Figure:
    """
    Generates a radar plot that compares **dief@t** with conventional metrics for a specific test.

    This function plots the results reported for a single given test in "Experiment 1" (see :cite:p:`dief`).
    "Experiment 1" compares the performance of testing approaches when using metrics defined in the literature
    (*total execution time*, *time for the first tuple*, *throughput*, and *completeness*) and the metric **dieft@t**.

    :param allmetrics: Dataframe with all the metrics from "Experiment 1".
    :param q: ID of the selected test to plot.
    :param colors: List of colors to use for the different approaches.
    :return: Matplotlib radar plot for the specified test over the provided metrics.

    **Examples**

    >>> plot_performance_of_approaches_with_dieft(extended_metrics, "Q9.sparql")
    >>> plot_performance_of_approaches_with_dieft(extended_metrics, "Q9.sparql", ["#ECC30B","#D56062","#84BCDA"])
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("invtfft", allmetrics["invtfft"].dtype),
            ("invtotaltime", allmetrics["invtotaltime"].dtype),
            ("comp", float),
            ("throughput", allmetrics["throughput"].dtype),
            ("dieft", allmetrics["dieft"].dtype),
        ],
    )

    # Obtain approaches.
    approaches = np.unique(allmetrics["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )
    color_map = dict(zip(sorted_approaches, colors))
    labels = []
    for a in sorted_approaches:
        submetric_approaches = allmetrics[
            (allmetrics["approach"] == a) & (allmetrics["test"] == q)
        ]

        if submetric_approaches.size == 0:
            continue
        else:
            labels.append(a)

        res = np.array(
            [
                (
                    (submetric_approaches["invtfft"]),
                    (submetric_approaches["invtotaltime"]),
                    (submetric_approaches["comp"]),
                    (submetric_approaches["throughput"]),
                    (submetric_approaches["dieft"]),
                )
            ],
            dtype=[
                ("invtfft", submetric_approaches["invtfft"].dtype),
                ("invtotaltime", submetric_approaches["invtotaltime"].dtype),
                ("comp", float),
                ("throughput", submetric_approaches["throughput"].dtype),
                ("dieft", submetric_approaches["dieft"].dtype),
            ],
        )
        df = np.append(df, res, axis=0)

    # Get maximum values
    maxs = [
        df["invtfft"].max(),
        df["invtotaltime"].max(),
        df["comp"].max(),
        df["throughput"].max(),
        df["dieft"].max(),
    ]

    # Normalize the data
    for row in df:
        row["invtfft"] = row["invtfft"] / maxs[0]
        row["invtotaltime"] = row["invtotaltime"] / maxs[1]
        row["comp"] = row["comp"] / maxs[2]
        row["throughput"] = row["throughput"] / maxs[3]
        row["dieft"] = row["dieft"] / maxs[4]

    # Plot metrics using spider plot.
    df = df.tolist()
    N = len(df[0])
    theta = radar_factory(N, frame="polygon")
    spoke_labels = ["(TFFT)^-1", "(ET)^-1       ", "Comp", "T", "     dief@t"]
    case_data = df
    fig, ax = plt.subplots(figsize=(6, 6), subplot_kw=dict(projection="radar"))
    fig.subplots_adjust(top=0.85, bottom=0.05)
    ax.set_ylim(0, 1)
    ticks_loc = ax.get_yticks()
    ax.yaxis.set_major_locator(mticker.FixedLocator(ticks_loc))
    ax.set_yticklabels("" for _ in ticks_loc)
    legend_handles = []
    for d, label in zip(case_data, labels):
        legend_handles.append(
            mlines.Line2D([], [], color=color_map[label], ls="-", label=label)
        )
        ax.plot(theta, d, label=label, color=color_map[label], zorder=10, clip_on=False)
        ax.fill(theta, d, label=label, facecolor=color_map[label], alpha=0.15)

    ax.set_varlabels(spoke_labels)
    ax.tick_params(labelsize=14)
    ax.legend(
        handles=legend_handles,
        loc=(0.80, 0.90),
        labelspacing=0.1,
        fontsize="medium",
        frameon=False,
    )

    plt.setp(ax.spines.values(), color="grey")
    title = q.split("-")[-1]
    plt.title(title, fontsize=16, loc="center", pad=30)
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
    results = inputtrace[inputtrace["test"] == inputtest]
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


def performance_of_approaches_with_dieft_edit(
    traces: np.ndarray, metrics: np.ndarray, continue_to_end: bool = True
) -> np.ndarray:
    """
    Compares **dief@t** with other conventional metrics used in query performance analysis.

    This function repeats the results reported in "Experiment 1" of :cite:p:`dief`.
    "Experiment 1" compares the performance of testing approaches when using metrics defined in the
    literature (*total execution time*, *time for the first tuple*, *throughput*, and *completeness*) and the metric **dieft@t**.

    :param traces: Dataframe with the answer trace. Attributes of the dataframe: test, approach, answer, time.
    :param metrics: Metrics dataframe with the result of the other metrics.
                    The structure is as follows: test, approach, tfft, totaltime, comp.
    :param continue_to_end: Indicates whether the AUC should be continued until the end of the time frame
    :return: Dataframe with all the metrics.
             The structure is: test, approach, tfft, totaltime, comp, throughput, invtfft, invtotaltime, dieft

    **Examples**

    >>> performance_of_approaches_with_dieft(traces, metrics)
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("test", traces["test"].dtype),
            ("approach", traces["approach"].dtype),
            ("tfft (ms)", metrics["tfft"].dtype),
            ("totaltime (s)", metrics["totaltime"].dtype),
            ("mrt (ms)", metrics["mrt"].dtype),
            ("checks", metrics["comp"].dtype),
            ("throughput", float),
            ("invtfft", float),
            ("invtotaltime", float),
            ("invmrt", float),
            ("dieft", float),
        ],
    )

    # Obtain tests and approaches.
    tests = np.unique(metrics["test"])
    approaches = np.unique(metrics["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    # Compute metrics: dieft, throughput, inverse of execution time, inverse of time for the first tuple.
    for t in tests:
        subtrace = traces[traces["test"] == t]
        dieft_res = dieft_edit(subtrace, t, continue_to_end=continue_to_end)

        for a in sorted_approaches:
            if a not in np.unique(dieft_res["approach"]):
                continue

            dieft_ = dieft_res[(dieft_res["approach"] == a) & (dieft_res["test"] == t)][
                "dieft"
            ][0]
            submetric = metrics[(metrics["approach"] == a) & (metrics["test"] == t)]

            throughput = submetric["comp"][0] / submetric["totaltime"][0]
            invtfft = 1 / submetric["tfft"][0]
            invtotaltime = 1 / submetric["totaltime"][0]
            invmrt = 1 / submetric["mrt"][0]

            res = np.array(
                [
                    (
                        t,
                        a,
                        submetric["tfft"][0],
                        submetric["totaltime"][0],
                        submetric["mrt"][0],
                        submetric["comp"][0],
                        throughput,
                        invtfft,
                        invtotaltime,
                        invmrt,
                        dieft_,
                    )
                ],
                dtype=[
                    ("test", submetric["test"].dtype),
                    ("approach", submetric["approach"].dtype),
                    ("tfft (ms)", submetric["tfft"].dtype),
                    ("totaltime (s)", submetric["totaltime"].dtype),
                    ("mrt (ms)", submetric["mrt"].dtype),
                    ("checks", submetric["comp"].dtype),
                    ("throughput", float),
                    ("invtfft", float),
                    ("invtotaltime", float),
                    ("invmrt", float),
                    ("dieft", float),
                ],
            )
            df = np.append(df, res, axis=0)

    return df


def plot_performance_of_approaches_with_dieft_edit(
    allmetrics: np.ndarray, q: str, colors: list = DEFAULT_COLORS
) -> Figure:
    """
    Generates a radar plot that compares **dief@t** with conventional metrics for a specific test.

    This function plots the results reported for a single given test in "Experiment 1" (see :cite:p:`dief`).
    "Experiment 1" compares the performance of testing approaches when using metrics defined in the literature
    (*total execution time*, *time for the first tuple*, *throughput*, and *completeness*) and the metric **dieft@t**.

    :param allmetrics: Dataframe with all the metrics from "Experiment 1".
    :param q: ID of the selected test to plot.
    :param colors: List of colors to use for the different approaches.
    :return: Matplotlib radar plot for the specified test over the provided metrics.

    **Examples**

    >>> plot_performance_of_approaches_with_dieft(extended_metrics, "Q9.sparql")
    >>> plot_performance_of_approaches_with_dieft(extended_metrics, "Q9.sparql", ["#ECC30B","#D56062","#84BCDA"])
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("invtfft", allmetrics["invtfft"].dtype),
            ("invtotaltime", allmetrics["invtotaltime"].dtype),
            ("invmrt", allmetrics["invmrt"].dtype),
            ("throughput", allmetrics["throughput"].dtype),
            ("dieft", allmetrics["dieft"].dtype),
        ],
    )

    # Obtain approaches.
    approaches = np.unique(allmetrics["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )
    color_map = dict(zip(sorted_approaches, colors))
    labels = []
    for a in sorted_approaches:
        submetric_approaches = allmetrics[
            (allmetrics["approach"] == a) & (allmetrics["test"] == q)
        ]

        if submetric_approaches.size == 0:
            continue
        else:
            labels.append(a)

        res = np.array(
            [
                (
                    (submetric_approaches["invtfft"]),
                    (submetric_approaches["invtotaltime"]),
                    (submetric_approaches["invmrt"]),
                    (submetric_approaches["throughput"]),
                    (submetric_approaches["dieft"]),
                )
            ],
            dtype=[
                ("invtfft", submetric_approaches["invtfft"].dtype),
                ("invtotaltime", submetric_approaches["invtotaltime"].dtype),
                ("invmrt", submetric_approaches["invmrt"].dtype),
                ("throughput", submetric_approaches["throughput"].dtype),
                ("dieft", submetric_approaches["dieft"].dtype),
            ],
        )
        df = np.append(df, res, axis=0)

    # Get maximum values
    maxs = [
        df["invtfft"].max(),
        df["invtotaltime"].max(),
        df["invmrt"].max(),
        df["throughput"].max(),
        df["dieft"].max(),
    ]

    # Normalize the data
    for row in df:
        row["invtfft"] = row["invtfft"] / maxs[0]
        row["invtotaltime"] = row["invtotaltime"] / maxs[1]
        row["invmrt"] = row["invmrt"] / maxs[2]
        row["throughput"] = row["throughput"] / maxs[3]
        row["dieft"] = row["dieft"] / maxs[4]

    # Plot metrics using spider plot.
    df = df.tolist()
    N = len(df[0])
    theta = radar_factory(N, frame="polygon")
    spoke_labels = ["(TFFT)^-1", "(ET)^-1       ", "(MRT)^-1", "T", "     dief@t"]
    case_data = df
    fig, ax = plt.subplots(figsize=(6, 6), subplot_kw=dict(projection="radar"))
    fig.subplots_adjust(top=0.85, bottom=0.05)
    ax.set_ylim(0, 1)
    ticks_loc = ax.get_yticks()
    ax.yaxis.set_major_locator(mticker.FixedLocator(ticks_loc))
    ax.set_yticklabels("" for _ in ticks_loc)
    legend_handles = []
    for d, label in zip(case_data, labels):
        legend_handles.append(
            mlines.Line2D([], [], color=color_map[label], ls="-", label=label)
        )
        ax.plot(theta, d, label=label, color=color_map[label], zorder=10, clip_on=False)
        ax.fill(theta, d, label=label, facecolor=color_map[label], alpha=0.15)

    ax.set_varlabels(spoke_labels)
    ax.tick_params(labelsize=14)
    ax.legend(
        handles=legend_handles,
        loc=(0.80, 0.90),
        labelspacing=0.1,
        fontsize="medium",
        frameon=False,
    )

    plt.setp(ax.spines.values(), color="grey")
    title = q.split("-")[-1]
    plt.title(title, fontsize=16, loc="center", pad=30)
    plt.tight_layout()

    return fig


def diefk_edit(inputtrace: np.ndarray, inputtest: str, k: int = -1) -> np.ndarray:
    """
    Computes the **dief@k** metric for a specific test at a given number of answers *k*.

    **dief@k** measures the diefficiency while *k* answers are produced by computing
    the area under the curve of the answer traces.
    By default, the function computes the minimum of the total number of answer produces by the approaches.

    :param inputtrace: Dataframe with the answer trace. Attributes of the dataframe: test, approach, answer, time.
    :param inputtest: Specifies the specific test to analyze from the answer trace.
    :param k: Number of answers to compute dief@k for. By default, the function computes the minimum of the total number
              of answers produced by the approaches.
    :return: Dataframe with the dief@k values for each approach. Attributes of the dataframe: test, approach, diefk.

    **Examples**

    >>> diefk(traces, "Q9.sparql")
    >>> diefk(traces, "Q9.sparql", 1000)
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("test", inputtrace["test"].dtype),
            ("approach", inputtrace["approach"].dtype),
            ("diefk", float),
        ],
    )

    # Obtain test and approaches to compare.
    results = inputtrace[inputtrace["test"] == inputtest]
    approaches = np.unique(results["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    # Obtain k per approach.
    if k == -1:
        n = []
        for a in sorted_approaches:
            x = results[results["approach"] == a]
            n.append(len(x))
        k = min(n)

    # Compute diefk per approach.
    for a in sorted_approaches:
        dief = 0
        subtrace = results[(results["approach"] == a) & (results["answer"] <= k)]
        if len(subtrace) > 1:
            dief = np.trapz(subtrace["answer"], subtrace["time"])
        res = np.array(
            [(inputtest, a, dief)],
            dtype=[
                ("test", inputtrace["test"].dtype),
                ("approach", inputtrace["approach"].dtype),
                ("diefk", float),
            ],
        )
        df = np.append(df, res, axis=0)

    return df


def diefk2_edit(inputtrace: np.ndarray, inputtest: str, kp: float = -1.0) -> np.ndarray:
    """
    Computes the **dief@k** metric for a specific test at a given percentage of answers *kp*.

    **dief@k** measures the diefficiency while the first *kp* percent of answers are produced
    by computing the area under the curve of the answer traces.
    By default, this function behaves the same as ``diefk``. This also holds for kp = 1.0.
    The function computes the portion *kp* of the minimum number of answers produces by the approaches.

    :param inputtrace: Dataframe with the answer trace. Attributes of the dataframe: test, approach, answer, time.
    :param inputtest: Specifies the specific test to analyze from the answer trace.
    :param kp: Ratio of answers to compute dief@k for (kp in [0.0;1.0]). By default and when kp=1.0, this function behaves
               the same as diefk. It computes the kp portion of the minimum number of answers produced by the approaches.
    :return: Dataframe with the dief@k values for each approach. Attributes of the dataframe: test, approach, diefk.

    **Examples**

    >>> diefk2(traces, "Q9.sparql")
    >>> diefk2(traces, "Q9.sparql", 0.25)
    """
    # Obtain test and approaches to compare.
    results = inputtrace[inputtrace["test"] == inputtest]
    approaches = np.unique(results["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )

    # Obtain k per approach.
    n = []
    for a in approaches:
        x = results[results["approach"] == a]
        n.append(len(x))
    k = min(n)
    if kp > -1:
        k = k * kp

    # Compute diefk.
    df = diefk_edit(inputtrace, inputtest, k)

    return df


def plot_continuous_efficiency_with_diefk_edit(
    diefkDF: np.ndarray, q: str, colors: list = DEFAULT_COLORS
) -> Figure:
    """
    Generates a radar plot that compares **dief@k** at different answer completeness percentages for a specific test.

    This function plots the results reported for a single given test in "Experiment 2"
    (see :cite:p:`dief`).
    "Experiment 2" measures the continuous efficiency of approaches when producing
    the first 25%, 50%, 75%, and 100% of the answers.

    :param diefkDF: Dataframe with the results from "Experiment 2".
    :param q: ID of the selected test to plot.
    :param colors: List of colors to use for the different approaches.
    :return: Matplotlib plot for the specified test over the provided metrics.

    **Examples**

    >>> plot_continuous_efficiency_with_diefk(diefkDF, "Q9.sparql")
    >>> plot_continuous_efficiency_with_diefk(diefkDF, "Q9.sparql", ["#ECC30B","#D56062","#84BCDA"])
    """
    # Initialize output structure.
    df = np.empty(
        shape=0,
        dtype=[
            ("diefk25", float),
            ("diefk50", float),
            ("diefk75", float),
            ("diefk100", float),
        ],
    )

    # Obtain approaches.
    approaches = np.unique(diefkDF["approach"])
    sorted_approaches = sorted(
        approaches,
        key=lambda x: [int(i) if i.isdigit() else i for i in re.split("([0-9]+)", x)],
    )
    labels = []
    color_map = dict(zip(sorted_approaches, colors))

    for a in sorted_approaches:
        submetric_approaches = diefkDF[
            (diefkDF["approach"] == a) & (diefkDF["test"] == q)
        ]

        if submetric_approaches.size == 0:
            continue
        else:
            labels.append(a)

        res = np.array(
            [
                (
                    (submetric_approaches["diefk25"]),
                    (submetric_approaches["diefk50"]),
                    (submetric_approaches["diefk75"]),
                    (submetric_approaches["diefk100"]),
                )
            ],
            dtype=[
                ("diefk25", submetric_approaches["diefk25"].dtype),
                ("diefk50", submetric_approaches["diefk50"].dtype),
                ("diefk75", submetric_approaches["diefk75"].dtype),
                ("diefk100", submetric_approaches["diefk100"].dtype),
            ],
        )

        df = np.append(df, res, axis=0)

    # Get maximum values
    maxs = [
        df["diefk25"].max(),
        df["diefk50"].max(),
        df["diefk75"].max(),
        df["diefk100"].max(),
    ]

    # Normalize the data
    for row in df:
        row["diefk25"] = row["diefk25"] / maxs[0]
        row["diefk50"] = row["diefk50"] / maxs[1]
        row["diefk75"] = row["diefk75"] / maxs[2]
        row["diefk100"] = row["diefk100"] / maxs[3]

    # Plot metrics using spider plot.
    df = df.tolist()
    N = len(df[0])
    theta = radar_factory(N, frame="polygon")
    spoke_labels = ["k=25%", "k=50%      ", "k=75%", "        k=100%"]
    case_data = df
    fig, ax = plt.subplots(figsize=(6, 6), subplot_kw=dict(projection="radar"))
    fig.subplots_adjust(top=0.85, bottom=0.05)
    ax.set_ylim(0, 1)
    ticks_loc = ax.get_yticks()
    ax.yaxis.set_major_locator(mticker.FixedLocator(ticks_loc))
    ax.set_yticklabels("" for _ in ticks_loc)
    legend_handles = []
    for d, label in zip(case_data, labels):
        legend_handles.append(
            mlines.Line2D([], [], color=color_map[label], ls="-", label=label)
        )
        ax.plot(theta, d, color=color_map[label], zorder=10, clip_on=False)
        ax.fill(theta, d, facecolor=color_map[label], alpha=0.15)
    ax.set_varlabels(spoke_labels)
    ax.tick_params(labelsize=14, zorder=0)

    ax.legend(
        handles=legend_handles,
        loc=(0.80, 0.90),
        labelspacing=0.1,
        fontsize="medium",
        frameon=False,
    )

    plt.setp(ax.spines.values(), color="grey")
    title = q.split("-")[-1]
    plt.title(title, fontsize=16, loc="center", pad=30)
    plt.tight_layout()

    return fig


####################################################################################################################


if len(sys.argv) < 4:
    print(
        "Error, run like: $>python dieffpy.py resultsDirectoryPath TEST(name) DO_JOIN(0:no,1:yes)"
    )
    exit(1)

# Read name of the directory
input_dir = sys.argv[1]
test_name = sys.argv[2]
do_join = sys.argv[3] == "1"

if do_join:

    metrics_all = []
    header_metrics = False

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

    if metrics_all:
        metrics_all_df = pd.concat(metrics_all, ignore_index=True)
        # output metrics all csv
        output_metrics = input_dir + "/metrics.csv"
        metrics_all_df.to_csv(output_metrics, index=False, header=header_metrics)
    else:
        print("No metrics.csv files found to combine.")

############################ RESULT PLOTS AND METRICS ############################

outputPlotDir = input_dir + "/plots/"
if not os.path.exists(outputPlotDir):
    os.makedirs(outputPlotDir)


# total execution time plot
metrics = load_metrics(input_dir + "/metrics.csv")
# F: Conversions
# tfft (ns) -> ms
# mrt (ns)  -> ms
metrics["mrt"] = metrics["mrt"] / 1_000_000
metrics["tfft"] = metrics["tfft"] / 1_000_000

columns = ["test", "approach", "tfft", "totaltime", "mrt", "comp"]

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
