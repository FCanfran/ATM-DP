import pandas as pd
import numpy as np
import random
import datetime
from geopy.distance import geodesic, great_circle

def calculate_distance(atm_row, point):
    atm_loc = (atm_row['loc_latitude'], atm_row['loc_longitude'])
    distance = great_circle(atm_loc, point).kilometers
    return round(distance,3) # limit to 3 decimals only, km and meters

# get ordered ascending list by distance of the atms wrt card location coordinates
# Optional: limit to the ones that lie inside a specific distance threshold
# 2 approaches for the distance:
# - Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
# - Vicenty: Earth as a ellipsoid (oblate spheroid). More accurate. More expensive computation.
# NOTE that: Earth is neither perfectly spherical nor ellipse hence calculating the distance on its surface is a challenging task.
# https://www.neovasolutions.com/2019/10/04/haversine-vs-vincenty-which-is-the-best/
def get_ordered_atms(card_loc_latitude, card_loc_longitude, atm_df, threshold=None):
    # Create a copy of the original DataFrame to avoid modifying it - dataframes are mutable objects!
    atm_df_ordered = atm_df.copy()
    card_loc = (card_loc_latitude, card_loc_longitude)
    print(card_loc)
    # Calculate distances and add as a new column
    atm_df_ordered['distance'] = atm_df_ordered.apply(calculate_distance, point=card_loc, axis=1)
    # Sort DataFrame based on distance
    atm_df_ordered = atm_df_ordered.sort_values(by='distance', ascending=True).reset_index(drop=True)
    return atm_df_ordered

def transaction_generator(card, atm_df, start_date, tx_id):

    # create transaction dataframe
    cols = ['transaction_id', 
            'number_id', 
            'ATM_id', 
            'transaction_start', 
            'transaction_end', 
            'transaction_amount'
            ]
    transaction_df = pd.DataFrame(columns=cols)

    start_datetime = datetime.datetime.strptime(start_date, "%Y-%m-%d")

    # fix a constant seed so that experiments are reproducible
    key = int(str(card['number_id']) + str(card['client_id']))
    random.seed(int(key))
    np.random.seed(int(key))

    # 1. Ordered list of terminals by ascending distance to the client card location
    atm_df_ordered = get_ordered_atms(card['loc_latitude'], card['loc_longitude'], atm_df)
    # 2. ATMs subset - select a maximum of MAX_SIZE_ATM_SUBSET of ATMs that are at a distance
    # inferior or equal to MAX_DISTANCE to the residence of the client
    # TODO: Improve this - values are "testing" values
    MAX_SIZE_ATM_SUBSET = 10
    MAX_DISTANCE = 30 # km
    
    """
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

                    end_time_tx = start_time_tx + diff_end
                    end_time_delta = datetime.timedelta(seconds=end_time_tx)
                    # Add the timedelta to the start date
                    transaction_end = start_datetime + end_time_delta
                    print(transaction_end)

                    # -----------------------------------------------------------
                    # transaction_amount 
                    # based on card behavior params: amount_avg & amount_std
                    # normal distribution: mean = amount_avg, std = amount_std
                    print(f"amount_avg: {card['amount_avg']}")
                    print(f"amount_avg: {card['amount_std']}")
                    transaction_amount = np.random.normal(card['amount_avg'], card['amount_std'])
                    # If negative amount, draw from a uniform distribution
                    if transaction_amount < 0:
                        transaction_amount = np.random.uniform(0,card['amount_avg']*2)
                    
                    transaction_amount = np.round(transaction_amount,decimals=2)
                    print(transaction_amount)
                    new_tx = {
                        'transaction_id': tx_id,
                        'number_id': card['number_id'], # card id
                        'ATM_id': 1, # TODO 
                        'transaction_start': transaction_start, 
                        'transaction_end': transaction_end, 
                        'transaction_amount': transaction_amount
                    }

                    new_tx_df = pd.DataFrame([new_tx])
                    transaction_df = pd.concat([transaction_df, new_tx_df], ignore_index=True)
                    tx_id += 1 

    return transaction_df, tx_id
    """

def main():
    # Read the card and atm datasets
    atm_df = pd.read_csv('atm.csv')
    card_df = pd.read_csv('card.csv')

    print(atm_df)
    print(card_df)

    # create the transaction dataframe
    cols = ['transaction_id', 
            'number_id', 
            'ATM_id', 
            'transaction_start', 
            'transaction_end', 
            'transaction_amount'
            ]
    transaction_df = pd.DataFrame(columns=cols)

    # TODO: Define the start date
    start_date = "2018-04-01"
    tx_id = 0
    # 1st card
    tx_card, tx_id = transaction_generator(card_df.iloc[0], atm_df, start_date, tx_id)
    print(tx_card)
    print(tx_id) # to continue the next transaction_id on this value

    



if __name__ == "__main__":
    main()