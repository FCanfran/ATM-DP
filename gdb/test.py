import pandas as pd
import numpy as np  # For NaN values

# Example values
tx_id = 1
ATM_id = "ATM001"
transaction_start = "2024-09-23 15:00:00"
transaction_end = "2024-09-23 19:00:00"
card = {"number_id": "1234-5678-9012-3456"}

# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_start = {
    "transaction_id": tx_id,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": None,  # Leave empty using None or np.nan
    "transaction_amount": None,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_start])

# Assuming transaction_df is an existing DataFrame (empty or with data)
transaction_df = pd.DataFrame()  # Example empty DataFrame for demonstration

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)


# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_end = {
    "transaction_id": 2,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": transaction_end,  # Leave empty using None or np.nan
    "transaction_amount": 55.21,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_end])

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)

# -----------------------------------------------------------------------------------
ATM_id = "ATM001"
transaction_start = "2024-09-23 16:00:00"
transaction_end = "2024-09-23 17:00:00"
card = {"number_id": "1234-5678-9012-3456"}

# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_start = {
    "transaction_id": 3,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": None,  # Leave empty using None or np.nan
    "transaction_amount": None,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_start])

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)


# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_end = {
    "transaction_id": tx_id,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": transaction_end,  # Leave empty using None or np.nan
    "transaction_amount": 55.21,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_end])

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)

# -----------------------------------------------------------------------------------
ATM_id = "ATM001"
transaction_start = "2024-09-23 15:30:00"
transaction_end = "2024-09-23 15:50:00"
card = {"number_id": "1234-5678-9012-3456"}

# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_start = {
    "transaction_id": 3,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": None,  # Leave empty using None or np.nan
    "transaction_amount": None,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_start])

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)


# Create a new transaction with empty fields for transaction_end and transaction_amount
new_tx_end = {
    "transaction_id": tx_id,
    "number_id": card["number_id"],  # card id
    "ATM_id": ATM_id,
    "transaction_start": transaction_start,
    "transaction_end": transaction_end,  # Leave empty using None or np.nan
    "transaction_amount": 55.21,  # Leave empty using None or np.nan
}

# Increment transaction ID
tx_id += 1

# Create a DataFrame for the new transaction
new_tx_df = pd.DataFrame([new_tx_end])

tx_card_cleaned = new_tx_df.dropna(how="all").dropna(axis=1, how="all")
transaction_df = (
    tx_card_cleaned.copy()
    if transaction_df.empty
    else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
)


#######################################################################################

# Define custom sorting logic:
# - If tx_end is None use tx_start.
# - Otherwise, use tx_end.
transaction_df["sort_key"] = transaction_df.apply(
    lambda row: (
        row["transaction_end"]
        if pd.notna(row["transaction_end"])
        else row["transaction_start"]
    ),
    axis=1,
)

# Sort DataFrame based on the custom `sort_key` column
transaction_df = transaction_df.sort_values(by="sort_key", ascending=True).reset_index(
    drop=True
)

# Drop the `sort_key` column if you don't want it anymore
transaction_df = transaction_df.drop(columns=["sort_key"])


#######################################################################################


# Write the DataFrame to a CSV file
csv_file_path = "test.csv"
transaction_df.to_csv(csv_file_path, index=False)  # Don't write row numbers

print(f"DataFrame has been written to {csv_file_path}")
