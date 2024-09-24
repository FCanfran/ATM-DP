import pandas as pd
import numpy as np

# Sample DataFrame
data = {
    "transaction_id": [18, 0],
    "number_id": ["c-NIGER-1", "c-NIGER-0"],
    "ATM_id": ["OGUN-3", "BENUE-4"],
    "transaction_start": ["2018-04-01 00:40:53", "2018-04-01 13:46:12"],
    "transaction_end": ["2018-04-01 00:46:59", "2018-04-01 13:54:06"],
    "transaction_amount": [29128.2, 56202.66],
}

df = pd.DataFrame(data)

# Convert transaction_start and transaction_end to datetime
df["transaction_start"] = pd.to_datetime(df["transaction_start"])
df["transaction_end"] = pd.to_datetime(df["transaction_end"])

# Initialize a new DataFrame to hold the modified rows
new_rows = []

# Iterate through each row in the original DataFrame
for i, row in df.iterrows():
    # First row with None for transaction_end and transaction_amount
    row_first = row.copy()
    row_first["transaction_end"] = None
    row_first["transaction_amount"] = None

    # Second row with original values but incremented transaction_id
    row_second = row.copy()
    row_second["transaction_id"] += 1

    # Append both new rows
    new_rows.append(row_first)
    new_rows.append(row_second)

# Create a new DataFrame from the list of new rows
df_expanded = pd.DataFrame(new_rows)

# Print the resulting DataFrame
print(df_expanded)
