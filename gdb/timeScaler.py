import pandas as pd
import sys
from datetime import datetime, timedelta

# Pre:
# - input stream comes ordered by timestamp
format = "%Y-%m-%d %H:%M:%S"


def main():

    if len(sys.argv) < 2:
        print("Usage: python timeScaler.py <inputFileToScale>")
        sys.exit(1)

    tx_file = sys.argv[1]
    # read stream input file to be scaled
    tx_df = pd.read_csv(tx_file)

    # convert the timestamp fields to datetime type
    tx_df["transaction_start"] = pd.to_datetime(
        tx_df["transaction_start"], errors="coerce"
    )
    tx_df["transaction_end"] = pd.to_datetime(tx_df["transaction_end"], errors="coerce")

    # calculate the t: size of time interval of the given file
    first_timestamp = tx_df.iloc[0]["transaction_start"]  # first tx
    last_timestamp = tx_df.iloc[-1]["transaction_end"]  # last tx
    # Calculate the difference in seconds - size of the original time interval T
    time_difference = last_timestamp - first_timestamp
    T_original = int(time_difference.total_seconds())
    print(f"The original time interval size is T = {T_original} s")
    print(f"\t~ {time_difference.days} days")
    print(f"\t~ {round(T_original / 3600, 2)} hours")

    validInt = False
    while not validInt:  # Loop until a valid integer is entered
        try:
            T_new = int(
                input("Introduce the T' (in seconds) to scale the input stream: ")
            )
            validInt = True
        except ValueError:
            print("Invalid input. Please enter a valid integer as T'")

    # scaling timestamps
    # t' = t0 + ((t' - t0) x scale_factor)
    scale_factor = round(T_new / T_original, 6)
    print(scale_factor)

    tx_df["transaction_start"] = tx_df["transaction_start"].apply(
        lambda t_start: first_timestamp
        + timedelta(
            seconds=((t_start - first_timestamp).total_seconds() * scale_factor)
        )
    )

    # only if transaction_end is not NaN
    tx_df["transaction_end"] = tx_df["transaction_end"].apply(
        lambda t_end: (
            first_timestamp
            + timedelta(
                seconds=((t_end - first_timestamp).total_seconds() * scale_factor)
            )
            if pd.notna(t_end)
            else pd.NaT
        )  # Handle NaN (missing values)
    )

    tx_df["transaction_start"] = tx_df["transaction_start"].dt.strftime(format)
    tx_df["transaction_end"] = tx_df["transaction_end"].dt.strftime(format)

    outfilename = tx_file.replace(".csv", f"-scaled-{scale_factor}.csv")
    tx_df.to_csv(outfilename, index=False)
    print(f"Scaled stream file saved to {outfilename}")


if __name__ == "__main__":
    main()
