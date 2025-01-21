import pandas as pd

# Load the CSV file
file_path = "behavior.csv"
df = pd.read_csv(file_path)

average_withdrawal_day = df["withdrawal_day"].mean()
average_deposit_day = df["deposit_day"].mean()
average_inquiry_day = df["inquiry_day"].mean()
average_transfer_day = df["transfer_day"].mean()

df["num_ops_day"] = (
    df["withdrawal_day"] + df["deposit_day"] + df["inquiry_day"] + df["transfer_day"]
)

average_ops_day = df["num_ops_day"].mean()


print(f"Average ops_day: {average_ops_day:.4f}")
print(f"Average withdrawal_day: {average_withdrawal_day:.4f}")
print(f"average_deposit_day: {average_deposit_day:.4f}")
print(f"average_inquiry_day: {average_inquiry_day:.4f}")
print(f"average_transfer_day: {average_transfer_day:.4f}")

print(
    average_withdrawal_day
    + average_deposit_day
    + average_inquiry_day
    + average_transfer_day
)
