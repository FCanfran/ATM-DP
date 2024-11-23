import pandas as pd
import sys
from datetime import datetime

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

    # calculate the t: size of time interval of the given file
    interval_start = tx_df.iloc[0]["transaction_start"]  # first tx
    interval_end = tx_df.iloc[-1]["transaction_end"]  # last tx
    # convert to datetime
    dt1 = datetime.strptime(interval_start, format)
    dt2 = datetime.strptime(interval_end, format)
    # Calculate the difference in seconds - size of the original time interval t
    t_original = int((dt2 - dt1).total_seconds())
    print(f"The original time interval size is t = {t_original} s")

    validInt = False
    while not validInt:  # Loop until a valid integer is entered
        try:
            t_prime = int(
                input("Introduce the t' (in seconds) to scale the input stream: ")
            )
            validInt = True
        except ValueError:
            print("Invalid input. Please enter a valid integer as t'")


if __name__ == "__main__":
    main()
