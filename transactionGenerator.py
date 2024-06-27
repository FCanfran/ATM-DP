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
def get_ordered_atms(card_loc_latitude, card_loc_longitude, atm_df, max_size_subset, max_distance):
    # Create a copy of the original DataFrame to avoid modifying it - dataframes are mutable objects!
    atm_df_ordered = atm_df.copy()
    card_loc = (card_loc_latitude, card_loc_longitude)
    # Calculate distances and add as a new column
    atm_df_ordered['distance'] = atm_df_ordered.apply(calculate_distance, point=card_loc, axis=1)

    # Subset that has distnace <= max_distance
    atm_df_ordered = atm_df_ordered[atm_df_ordered['distance'] <= max_distance]

    # Sort DataFrame based on distance
    atm_df_ordered = atm_df_ordered.sort_values(by='distance', ascending=True).reset_index(drop=True)
   
    # Subset of max size of max_size_subset
    atm_df_ordered = atm_df_ordered.head(max_size_subset)

    return atm_df_ordered

# Distribute n transactions on a day [tmin/2, 86400-(tmin/2)]
# Returns a ordered list of start moments in seconds, respecting that all of the moments
# are at a minimum time distance of TMIN
# NOTE: max duration of a transaction defined by max_duration
def distribute_tx(n, max_duration):
    # TODO: pass by param 
    TMIN = 3600 # assume is 1h 
    # in seconds of a day: (86400s in a day) -> [tmin/2, 86400-(tmin/2)]
    lower_bound = TMIN/2
    upper_bound = 86400 - (TMIN/2) - max_duration

    if (upper_bound - lower_bound) < (n-1) * TMIN:
        raise ValueError(f"Impossible to distribute {n} transactions over a day with tmin = {TMIN}")

    moments = []
    while len(moments) < n:
        candidate = int(random.uniform(lower_bound, upper_bound))
        # to add this new moment of transaction, it is required that it respects 
        # the time distance constraint wrt all the other added moments
        if (all(abs(candidate - second) >= (TMIN + max_duration) for second in moments)):
            moments.append(candidate)
    
    moments.sort()
    return moments

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
    # & ATMs subset - select a maximum of MAX_SIZE_ATM_SUBSET of ATMs that are at a distance
    # inferior or equal to MAX_DISTANCE to the residence of the client
    # TODO: Improve this - values are "testing" values
    # NOTE
    MAX_SIZE_ATM_SUBSET = 10
    MAX_DISTANCE = 30 # km
    atm_df_ordered = get_ordered_atms(card['loc_latitude'], card['loc_longitude'], atm_df, MAX_SIZE_ATM_SUBSET, MAX_DISTANCE)

    #print(atm_df_ordered)

    # T_MIN: Minimum threshold time in between 2 transactions of this client
    # TODO: Calculate T_MIN? - based on the max distance between 2 atms of the subset list
    # NOTE: Approx -> 2 x MAX_DISTANCE kms is the upper bound on this max distance btw 2 atms of the subset list
    # Therefore we set the t_min approx to be the time needed to traverse that distance at 50km/h
    T_MIN = ((MAX_DISTANCE * 2) / 50) * 60 * 60 # in seconds

    # max_duration of a transaction
    max_duration = 600 # 600s - 10min
    
    # 3. Generation of transactions
    num_days = 10 # NOTE
    for day in range(num_days):
        # random number of transactions on this day:
        # poisson of lambda = withdrawal_day (= avg number of withdrawals per day)
        num_tx = np.random.poisson(card['withdrawal_day'])

        if num_tx > 0:
            # distributed transaction start moments on a day (in seconds)
            moments = distribute_tx(num_tx, max_duration)
            for moment in moments:
                # 0. ATM id
                # randomly among the subset of ATMs -> all of them satisfy the constraints
                # of the min threshold time TMIN etc...
                rand_index = random.choice(atm_df_ordered.index)
                ATM_id = atm_df_ordered.loc[rand_index]['ATM_id']
                # transaction_start & transaction_end
                # 1. transaction_start
                # shift based on the number of day
                start_time_tx = (86400 * day) + moment 
                start_time_delta = datetime.timedelta(seconds=start_time_tx)
                # Add the timedelta to the start date
                transaction_start = start_datetime + start_time_delta
                # 2. transaction_end
                # -> increment some diff time based on the normal duration of a transaction
                # TODO: Define this better
                # -> for the moment: mean = 5min (300s), std = 2min (120s)
                # upper bound to 10min -> 600
                diff_end = int(np.random.normal(300,120))
                if (diff_end < 0): diff_end = 300 # if negative -> then it is = to the mean
                if (diff_end > max_duration): diff_end = max_duration # if above 10 mins -> then 10 min

                end_time_tx = start_time_tx + diff_end
                end_time_delta = datetime.timedelta(seconds=end_time_tx)
                transaction_end = start_datetime + end_time_delta

                # -----------------------------------------------------------
                # transaction_amount 
                # based on card behavior params: amount_avg & amount_std
                # normal distribution: mean = amount_avg, std = amount_std
                transaction_amount = np.random.normal(card['amount_avg'], card['amount_std'])
                # If negative amount, draw from a uniform distribution
                if transaction_amount < 0:
                    transaction_amount = np.random.uniform(0,card['amount_avg']*2)
                
                transaction_amount = np.round(transaction_amount,decimals=2)

                new_tx = {
                    'transaction_id': tx_id,
                    'number_id': card['number_id'], # card id
                    'ATM_id': ATM_id,
                    'transaction_start': transaction_start, 
                    'transaction_end': transaction_end, 
                    'transaction_amount': transaction_amount
                }

                new_tx_df = pd.DataFrame([new_tx])
                transaction_df = new_tx_df.copy() if transaction_df.empty else pd.concat([transaction_df, new_tx_df], ignore_index=True)
                tx_id += 1

    return transaction_df, tx_id

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

    for card_index in card_df.index:
        tx_card, tx_id = transaction_generator(card_df.iloc[card_index], atm_df, start_date, tx_id)
        # if transaction_df is empty (first iter) then directly assign the returned df, otherwise an ordinary concat
        transaction_df = tx_card.copy() if transaction_df.empty else pd.concat([transaction_df, tx_card], ignore_index=True)


    print(transaction_df)
    # sort by transaction_start, and if equal (if ties) by transaction_end in ascending order
    transaction_df = transaction_df.sort_values(by=["transaction_start", "transaction_end"], ascending=True).reset_index(drop=True)
    print(transaction_df)
    transaction_df.to_csv('transaction.csv', index=False)


if __name__ == "__main__":
    main()