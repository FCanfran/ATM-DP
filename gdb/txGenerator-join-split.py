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

    if len(sys.argv) < 3:
        print("Usage: python txGenerator-join-split.py <outFileName> <inputDir>")
        sys.exit(1)

    outFileName = sys.argv[1]
    inputDir = sys.argv[2]

    all_regular_df = []
    for file in os.listdir(inputDir):
        if file.endswith("-regular.csv"):
            file_path = os.path.join(inputDir, file)
            df = pd.read_csv(file_path)
            all_regular_df.append(df)

    regular_df = pd.concat(all_regular_df, ignore_index=True)
    # Assign a unique `transaction_id` for each row
    regular_df["transaction_id"] = range(len(regular_df))

    all_anomalous_df = []
    for file in os.listdir(inputDir):
        if file.endswith("anomalous.csv"):
            file_path = os.path.join(inputDir, file)
            df = pd.read_csv(file_path)
            all_anomalous_df.append(df)

    anomalous_df = pd.concat(all_anomalous_df, ignore_index=True)
    # start the ids of the anomalous on the last ids of the regulars
    last_id = len(regular_df)
    anomalous_df["transaction_id"] = range(last_id, last_id + len(anomalous_df))

    print(":::::::::::::: regular:")
    print(regular_df)
    print(":::::::::::::: anomalous:")
    print(anomalous_df)

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
    main()
