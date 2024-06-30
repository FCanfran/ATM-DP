import string
import random
import pandas as pd


# - name: Bank name.
# - code: Bank identifier code.
# - loc latitude: Bank headquarters GPS-location latitude.
# - loc longitude: Bank headquarters GPS-location longitude.
def main():
    # create the bank dataframe
    cols = ["name", "code", "loc_latitude", "loc_longitude"]
    bank_df = pd.DataFrame(columns=cols)

    bank_row = {
        "name": "Lagos Bank",
        "code": "LAGO",
        "loc_latitude": 6.478685,
        "loc_longitude": 3.368442,
    }

    bank_row_df = pd.DataFrame([bank_row])
    bank_df = pd.concat([bank_df, bank_row_df], ignore_index=True)

    bank_row = {
        "name": "Kano Bank",
        "code": "KANO",
        "loc_latitude": 11.994949,
        "loc_longitude": 8.520313,
    }

    bank_row_df = pd.DataFrame([bank_row])
    bank_df = pd.concat([bank_df, bank_row_df], ignore_index=True)

    bank_row = {
        "name": "Abuya Bank",
        "code": "ABYA",
        "loc_latitude": 9.042977,
        "loc_longitude": 7.478564,
    }

    bank_row_df = pd.DataFrame([bank_row])
    bank_df = pd.concat([bank_df, bank_row_df], ignore_index=True)

    print(bank_df)
    bank_df.to_csv("csv/bank.csv", index=False)


if __name__ == "__main__":
    main()
