import pandas as pd
import numpy as np
import datetime
from geopy.distance import geodesic, great_circle
import sys
from bitarray import bitarray
import random
import math
import os
import csv
from tqdm import tqdm
import time


# Splits the tx of a dataframe, so that from each tx, 2 edges are generated: tx_start & tx_end
def split_tx(tx_df):

    new_tx = []  # list of new rows, later converted to df

    # Create a new df, such that, for each tx we generate 2, 1 for tx_start and 1 for tx_end
    for _, tx in tx_df.iterrows():
        tx_start = tx.copy()
        tx_start["transaction_end"] = None
        tx_start["transaction_amount"] = None

        tx_end = tx.copy()

        new_tx.append(tx_start)
        new_tx.append(tx_end)

    new_tx_df = pd.DataFrame(new_tx)
    return new_tx_df


def main():

    if len(sys.argv) < 4:
        print(
            "Usage: python txGenerator-split.py <outFileName> <txRegular.csv> <txAnomalous.csv>"
        )
        sys.exit(1)

    # Read the regular and the anomalous csvs
    # TODO: Assuming only 1 of each for the moment
    outFileName = sys.argv[1]
    regular_df = pd.read_csv(sys.argv[2])
    anomalous_df = pd.read_csv(sys.argv[3])
    # - all tx - split and sorted

    # Split the tx in two: tx_start and tx_end
    # Custom sorting logic:
    # - If tx_end is None use tx_start.
    # - Otherwise, use tx_end.

    # Split the tx in 2: tx_start and tx_end
    transaction_df_ext = split_tx(regular_df)
    transaction_df_ext["sort_key"] = transaction_df_ext.apply(
        lambda row: (
            row["transaction_end"]
            if pd.notna(
                row["transaction_end"]
            )  # if tx_end is not missing (not NaN) -> use tx_end to sort, otherwise use tx_start
            else row["transaction_start"]
        ),
        axis=1,
    )

    if len(anomalous_df) > 0:
        # Split the tx in 2: tx_start and tx_end
        anomalous_df_ext = split_tx(anomalous_df)
        anomalous_df_ext["sort_key"] = anomalous_df_ext.apply(
            lambda row: (
                row["transaction_end"]
                if pd.notna(row["transaction_end"])
                else row["transaction_start"]
            ),
            axis=1,
        )

        anomalous_df_ext = anomalous_df_ext.sort_values(
            by="sort_key", ascending=True
        ).reset_index(drop=True)

        # Join regular & anomalous
        all_tx_ext = pd.concat(
            [transaction_df_ext, anomalous_df_ext], ignore_index=True
        )
        # Drop the sort_key column
        anomalous_df_ext = anomalous_df_ext.drop(columns=["sort_key"])
        # Write csv
        anomalous_df_ext.to_csv(
            "tx/" + outFileName + "-anomalous-split.csv", index=False
        )
    else:
        all_tx_ext = transaction_df_ext

    # sort
    all_tx_ext = all_tx_ext.sort_values(by=["sort_key"], ascending=True).reset_index(
        drop=True
    )
    all_tx_ext = all_tx_ext.drop(columns=["sort_key"])
    all_tx_ext.to_csv("tx/" + outFileName + "-all-split.csv", index=False)


if __name__ == "__main__":
    start_time = time.time()
    main()
    end_time = time.time()
    print(f"Execution time: {end_time - start_time:.4f} seconds")
