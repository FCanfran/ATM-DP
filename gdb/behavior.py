import pandas as pd
import csv
from tqdm import tqdm


# Different types of transactions:
# - 1: Withdrawal       (Retirada de dinero)
# - 2: Deposit          (Ingreso)
# - 3: Balance Inquiry  (Consulta de saldo/balance)
# - 4: Transfer         (Transferencia)
# The behavior gathers metrics for each of the kind of operations
def get_client_behavior_wisabi(customer):
    # CardholderID to locate the transactions of the customer in the wisabi dataset
    # for a customer, all the transactions take place in the same atm (in the wisabi dataset)

    # CardholderID
    # -> to gather the transactions of this client
    # -> also indicates in which transaction csv we have to look into
    behavior = {}
    cardholderid = customer["CardholderID"]
    csv_code = cardholderid.split("-")[
        0
    ]  # to read the transactions from the corresponding CSV

    if csv_code == "EN":
        csv_file = "enugu_transactions.csv"
    elif csv_code == "FC":
        csv_file = "fct_transactions.csv"
    elif csv_code == "KN":
        csv_file = "kano_transactions.csv"
    elif csv_code == "LA":
        csv_file = "lagos_transactions.csv"
    elif csv_code == "RI":
        csv_file = "rivers_transactions.csv"
    else:
        print("No matching transaction file, csv code was:", csv_code)
        return

    all_transactions_df = pd.read_csv("wisabi/" + csv_file)

    # obtain all the transactions of the customer by the cardholderid
    transactions = all_transactions_df[
        (all_transactions_df["CardholderID"] == cardholderid)
    ]
    # print(f"# of transactions: {len(transactions)}")

    withdrawals = transactions[(transactions["TransactionTypeID"] == 1)]
    deposits = transactions[(transactions["TransactionTypeID"] == 2)]
    inquiries = transactions[(transactions["TransactionTypeID"] == 3)]
    transfers = transactions[(transactions["TransactionTypeID"] == 4)]

    # print(f"# of withdrawals: {len(withdrawals)}")
    # print(f"# of deposits: {len(deposits)}")
    # print(f"# of balance_inquiries: {len(inquiries)}")
    # print(f"# of transfers: {len(transfers)}")

    # Metrics - Withdrawals
    if not withdrawals.empty:
        amount_avg = round(withdrawals["TransactionAmount"].mean(), 2)
        amount_std = round(withdrawals["TransactionAmount"].std(), 2)
        # Number of withdrawals per day - we have transactions of the year 2022 - 365 days
        num_transacc_per_day = round(len(withdrawals) / 365, 4)
        behavior["amount_avg_withdrawal"] = amount_avg
        behavior["amount_std_withdrawal"] = amount_std
        behavior["withdrawal_day"] = num_transacc_per_day
    else:
        print("No matching withdrawals with CardholderID found in transactions table")

    # Metrics - Deposits
    if not deposits.empty:
        amount_avg = round(deposits["TransactionAmount"].mean(), 2)
        amount_std = round(deposits["TransactionAmount"].std(), 2)
        # Number of Deposits per day - we have transactions of the year 2022 - 365 days
        num_transacc_per_day = round(len(deposits) / 365, 4)
        behavior["amount_avg_deposit"] = amount_avg
        behavior["amount_std_deposit"] = amount_std
        behavior["deposit_day"] = num_transacc_per_day
    else:
        print("No matching deposits with CardholderID found in transactions table")

    # Metrics - Inquiries
    if not inquiries.empty:
        # Number of inquiries per day - we have transactions of the year 2022 - 365 days
        num_transacc_per_day = round(len(inquiries) / 365, 4)
        behavior["inquiry_day"] = num_transacc_per_day
    else:
        print("No matching inquiries with CardholderID found in transactions table")

    # Metrics - Transfers
    if not transfers.empty:
        amount_avg = round(transfers["TransactionAmount"].mean(), 2)
        amount_std = round(transfers["TransactionAmount"].std(), 2)
        # Number of transfers per day - we have transactions of the year 2022 - 365 days
        num_transacc_per_day = round(len(transfers) / 365, 4)
        behavior["amount_avg_transfer"] = amount_avg
        behavior["amount_std_transfer"] = amount_std
        behavior["transfer_day"] = num_transacc_per_day
    else:
        print("No matching transfers with CardholderID found in transactions table")

    return behavior


def main():

    # wisabi customers
    customers_file = "wisabi/customers_lookup.csv"
    customers_df_wisabi = pd.read_csv(customers_file)

    # obtain the behavior of all the customers of wisabi - csv file with the gathered
    # behavior for each of the customers
    # header
    cols = [
        "cardholderid",
        "amount_avg_withdrawal",
        "amount_std_withdrawal",
        "withdrawal_day",
        "amount_avg_deposit",
        "amount_std_deposit",
        "deposit_day",
        "inquiry_day",
        "amount_avg_transfer",
        "amount_std_transfer",
        "transfer_day",
    ]
    with open("behavior.csv", mode="w", newline="") as file:
        writer = csv.DictWriter(file, fieldnames=cols)
        writer.writeheader()
        for i in tqdm(
            range(0, len(customers_df_wisabi)),
            desc="Creating customer behavior dictionary of the wisabi dataset",
        ):
            customer = customers_df_wisabi.iloc[i]
            cardholderid = customer["CardholderID"]
            behavior = get_client_behavior_wisabi(customer)
            behavior["cardholderid"] = cardholderid
            writer.writerow(behavior)


if __name__ == "__main__":
    main()
