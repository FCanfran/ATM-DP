import os
import pandas as pd

# Bank instance generator.
# - Insertion of all the needed details of the bank instance and creation in the form of a csv.


# - name: Bank name.
# - code: Bank identifier code.
# - loc latitude: Bank headquarters GPS-location latitude.
# - loc longitude: Bank headquarters GPS-location longitude.
def main():
    # create the bank dataframe
    cols = ["name", "code", "loc_latitude", "loc_longitude"]
    bank_df = pd.DataFrame(columns=cols)

    print("Introduce the attributes of the bank instance")
    name = input("name: ")
    code = input("bank code: ")
    coordinates = input("bank location coordinates (latitude, longitude): ")
    latitude, longitude = map(float, coordinates.split(","))
    print(latitude, longitude)

    # Example case
    """
    bank_row = {
        "name": "Niger Bank",
        "code": "NIGER",
        "loc_latitude": 6.478685,
        "loc_longitude": 3.368442,
    }
    """

    bank = {
        "name": name,
        "code": code,
        "loc_latitude": latitude,
        "loc_longitude": longitude,
    }

    bank_row_df = pd.DataFrame([bank])
    bank_df = pd.concat(
        [bank_df.dropna(axis=1, how="all"), bank_row_df.dropna(axis=1, how="all")],
        ignore_index=True,
    )

    print(bank_df)

    os.makedirs("csv", exist_ok=True)  # Ensure the 'csv/' directory exists

    bank_df.to_csv("csv/bank.csv", index=False)


if __name__ == "__main__":
    main()
