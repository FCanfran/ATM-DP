import pandas as pd
import numpy as np
import random
import datetime

# months: number of months for which we generate transactions
def transaction_generator(card, atm_df):

    # Define the start date
    start_date = "2018-04-01"
    start_datetime = datetime.datetime.strptime(start_date, "%Y-%m-%d")

    # fix a constant seed so that experiments are reproducible
    key = int(str(card['number_id']) + str(card['client_id']))
    random.seed(int(key))
    np.random.seed(int(key))
    num_days = 10
    for day in range(num_days):
        
        print(f"day {day}")
        # random number of transactions on this day:
        # poisson of lambda = withdrawal_day (= avg number of withdrawals per day)
        num_tx = np.random.poisson(card['withdrawal_day'])
        print(num_tx)

        if num_tx > 0:
            for tx in range(num_tx):
                # transaction_start & transaction_end
                # 1. transaction_start
                # Time of transaction: Around noon, std 20000 seconds. 
                # This choice aims at simulating the fact that most transactions occur during the day.
                # 24h x 60 min x 60 s = 86400s in a day
                # half day (noon) = 86400s / 2
                start_time_tx = int(np.random.normal(86400/2, 20000))
                print(start_time_tx)
                # If transaction time between 0 and 86400, let us keep it, otherwise, let us discard it
                if (start_time_tx > 0) and (start_time_tx < 86400):
                    # shift based on the number of day
                    start_time_tx = (86400 * day) + start_time_tx 
                    start_time_delta = datetime.timedelta(seconds=start_time_tx)
                    # Add the timedelta to the start date
                    transaction_start = start_datetime + start_time_delta
                    print(transaction_start)

                    # 2. transaction_end
                    # -> increment some diff time based on the normal duration of a transaction
                    # TODO: Define this better
                    # -> for the moment: mean = 5min (300s), std = 2min (120s)
                    diff_end = int(np.random.normal(300,120))
                    if (diff_end < 0): diff_end = 300 # if negative -> then it is = to the mean

                    print(f"diff_end = {diff_end}")
                    end_time_tx = start_time_tx + diff_end
                    end_time_delta = datetime.timedelta(seconds=end_time_tx)
                    # Add the timedelta to the start date
                    transaction_end = start_datetime + end_time_delta
                    print(transaction_end)

def main():
    # Read the card and atm datasets
    atm_df = pd.read_csv('atm.csv')
    card_df = pd.read_csv('card.csv')

    print(atm_df)
    print(card_df)

    # 1st card
    transaction_generator(card_df.iloc[0], atm_df)

    



if __name__ == "__main__":
    main()