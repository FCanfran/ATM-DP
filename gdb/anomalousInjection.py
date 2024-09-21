import pandas as pd
import numpy as np
import datetime
from geopy.distance import geodesic, great_circle
import sys
from bitarray import bitarray

# Parameters
# --------------------------------------------------------------------------
start_date = "2018-04-01"  # start date, from which the first transaction is generated
num_days = 5  # num of days for which transactions are generated (init start_date)
# TODO: Improve this - values are "testing" values
max_size_atm_subset = 10  # maximum size of the ATM subset
max_distance = 30  # maximum distance of the atms in the ATM subset to client residence
# duration of a transaction
# TODO: Define this better
max_duration = 600  # max duration - 600s (10min)
mean_duration = 300  # mean duration - 300s (5min)
std_duration = 120  # std duration - 120s  (2min)
# Needed to calculate the t_min: time needed to traverse the distance between 2 geographical points at SPEED km/h
# -> speed at which we consider the client travels NORMALY (by any means of transport) between 2 points
SPEED = 50
# --------------------------------------------------------------------------

# ------------------
# number of cards for which no transactions can be generated due to the specific required conditions
# -> for example: empty ATM subset (since the distance of the residence to the closest ATM is > max_distance)
fail_cards = 0
success_cards = 0


# 2 approaches for the distance:
# - Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
# - Vicenty: Earth as a ellipsoid (oblate spheroid). More accurate. More expensive computation.
# NOTE that: Earth is neither perfectly spherical nor ellipse hence calculating the distance on its surface is a challenging task.
# https://www.neovasolutions.com/2019/10/04/haversine-vs-vincenty-which-is-the-best/
def calculate_distance(atm_row, point):
    atm_loc = (atm_row["loc_latitude"], atm_row["loc_longitude"])
    distance = great_circle(atm_loc, point).kilometers
    return round(distance, 3)  # limit to 3 decimals only, km and meters


# Get ordered ascending list by distance of the atms wrt card location coordinates
# Optional: limit to the ones that lie inside a specific distance threshold
def get_ordered_atms(
    card_loc_latitude, card_loc_longitude, atm_df, max_size_subset, max_distance
):
    # Create a copy of the original DataFrame to avoid modifying it - dataframes are mutable objects!
    atm_df_ordered = atm_df.copy()
    card_loc = (card_loc_latitude, card_loc_longitude)
    # Calculate distances and add as a new column
    atm_df_ordered["distance"] = atm_df_ordered.apply(
        calculate_distance, point=card_loc, axis=1
    )

    # Sort DataFrame based on distance
    atm_df_ordered = atm_df_ordered.sort_values(
        by="distance", ascending=True
    ).reset_index(drop=True)

    # The "regular" subset: select those with distance <= max_distance
    atm_df_regular = atm_df_ordered[atm_df_ordered["distance"] <= max_distance]
    # The "non-regular" subset: the rest
    atm_df_non_regular = atm_df_ordered[atm_df_ordered["distance"] > max_distance]

    # Subset of max size of max_size_subset
    atm_df_regular = atm_df_regular.head(max_size_subset)

    # TODO: Give priority to the ATMs belonging to the same bank company as the card

    return atm_df_regular, atm_df_non_regular


# Distribute n transactions on a day [tmin/2, 86400-(tmin/2)]
# Returns a ordered list of start moments in seconds, respecting that all of the moments
# are at a minimum time distance of TMIN
def distribute_tx(n, t_min):
    # in seconds of a day: (86400s in a day) -> [tmin/2, 86400-(tmin/2)]
    lower_bound = t_min / 2
    upper_bound = 86400 - (t_min / 2) - max_duration

    if (upper_bound - lower_bound) < (n - 1) * t_min:
        raise ValueError(
            f"Impossible to distribute {n} transactions over a day with tmin = {t_min}"
        )

    moments = []
    while len(moments) < n:
        candidate = int(np.random.uniform(lower_bound, upper_bound))
        # to add this new moment of transaction, it is required that it respects
        # the time distance constraint wrt all the other added moments
        if all(abs(candidate - second) >= (t_min + max_duration) for second in moments):
            moments.append(candidate)

    moments.sort()
    return moments


def transaction_generator(card, atm_df, start_date, tx_id):

    print(f"------------- Generation for card: {card['number_id']} -------------")
    # create transaction dataframe
    cols = [
        "transaction_id",
        "number_id",
        "ATM_id",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    transaction_df = pd.DataFrame(columns=cols)

    start_datetime = datetime.datetime.strptime(start_date, "%Y-%m-%d")

    # 1. Ordered list of terminals by ascending distance to the client card location
    # selecting a maximum of max_size_atm_subset of ATMs that are at a distance
    # inferior or equal to max_distance to the residence of the client
    atm_df_regular, atm_df_non_regular = get_ordered_atms(
        card["loc_latitude"],
        card["loc_longitude"],
        atm_df,
        max_size_atm_subset,
        max_distance,
    )

    if len(atm_df_regular) > 0:
        # T_MIN: Minimum threshold time in between 2 transactions of this client
        # TODO: Calculate t_min? - based on the max distance between 2 atms of the subset list
        # NOTE: Approx -> 2 x MAX_DISTANCE kms is the upper bound on this max distance btw 2 atms of the subset list
        # Therefore we set the t_min approx to be the time needed to traverse that distance at SPEED km/h
        t_min = ((max_distance * 2) / SPEED) * 60 * 60  # in seconds

        # 3. Generation of transactions
        for day in range(num_days):
            # random number of transactions on this day:
            # poisson of lambda = withdrawal_day (= avg number of withdrawals per day)
            num_tx = np.random.poisson(card["withdrawal_day"])
            if num_tx > 0:
                # distributed transaction start moments on a day (in seconds)
                moments = distribute_tx(num_tx, t_min)
                for moment in moments:
                    # 0. ATM id
                    # randomly among the subset of ATMs -> all of them satisfy the constraints
                    # of the min threshold time TMIN etc...
                    rand_index = np.random.choice(atm_df_regular.index)
                    ATM_id = atm_df_regular.loc[rand_index]["ATM_id"]
                    # 1. transaction_start
                    # shift based on the number of day
                    start_time_tx = (86400 * day) + moment
                    start_time_delta = datetime.timedelta(seconds=start_time_tx)
                    # Add the timedelta to the start date
                    transaction_start = start_datetime + start_time_delta
                    # 2. transaction_end
                    # -> increment some diff time based on the normal duration of a transaction
                    diff_end = int(np.random.normal(mean_duration, std_duration))
                    if diff_end < 0:
                        diff_end = (
                            mean_duration  # if negative -> then it is = to the mean
                        )
                    if diff_end > max_duration:
                        diff_end = max_duration  # if above 10 mins -> then 10 min

                    end_time_tx = start_time_tx + diff_end
                    end_time_delta = datetime.timedelta(seconds=end_time_tx)
                    transaction_end = start_datetime + end_time_delta

                    # 3. transaction_amount
                    # based on card behavior params: amount_avg & amount_std
                    # normal distribution: mean = amount_avg, std = amount_std
                    transaction_amount = np.random.normal(
                        card["amount_avg"], card["amount_std"]
                    )
                    # If negative amount, draw from a uniform distribution
                    if transaction_amount < 0:
                        transaction_amount = np.random.uniform(
                            0, card["amount_avg"] * 2
                        )

                    transaction_amount = np.round(transaction_amount, decimals=2)

                    new_tx = {
                        "transaction_id": tx_id,
                        "number_id": card["number_id"],  # card id
                        "ATM_id": ATM_id,
                        "transaction_start": transaction_start,
                        "transaction_end": transaction_end,
                        "transaction_amount": transaction_amount,
                    }

                    new_tx_df = pd.DataFrame([new_tx])
                    transaction_df = (
                        new_tx_df.copy()
                        if transaction_df.empty
                        else pd.concat([transaction_df, new_tx_df], ignore_index=True)
                    )
                    tx_id += 1

    else:
        # if ATM subset size = 0 -> then
        print(f"Empty ATM subset for card: {card['number_id']}")

    if len(transaction_df) == 0:
        global fail_cards
        fail_cards += 1
    else:
        global success_cards
        success_cards += 1

    return transaction_df, tx_id, atm_df_regular, atm_df_non_regular


# Introduction of anomalous tx to cause the fraud pattern 1
def introduce_anomalous_fp_1(regular_tx_card, ratio, atm_regular, atm_non_regular):
    num_regular = len(regular_tx_card)
    num_anomalous = round(num_regular * ratio)
    print(num_regular, num_anomalous)

    # randomly select in between which tx the anomalous are introduced

    # bit array to mark occupied and free tx "holes" - python bitarray
    # - holes indicate the position i in between which tx the anomalous tx is to be inserted:
    # i in [0,num_regular-1]
    # - i = 2 -> indicates that the anomalous tx is to be inserted in between the tx 2 and 3
    # after the tx 2
    holes = bitarray(num_regular)
    holes.setall(0)
    anomalous = 0
    # :::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: #
    while anomalous < num_anomalous:
        # random hole selection in [0, num_regular-1]
        index = np.random.randint(0, num_regular)
        if holes[index] == 0:
            # not occupied, mark as occupied
            holes[index] = 1
            print(regular_tx_card)
            print("index", index)
            # ................................ #
            # TODO: While loop in case we cant meet the conditions for the first selected random ATM from the atm_non_regular subset?
            # introduce anomalous tx in this position: after the tx[index] and before tx[index+1]
            tx_prev = regular_tx_card.iloc[index]
            tx_next = regular_tx_card.iloc[index + 1]
            print(tx_prev)
            print(tx_next)
            # select one ATM at random from atm_non_regular
            rand_index = np.random.choice(atm_non_regular.index)
            ATM_new = atm_non_regular.loc[rand_index]

            # Calculate t_min(ATM_prev, ATM_new)

            # ................................ #
            anomalous += 1
    # :::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::: #


def main():

    if len(sys.argv) < 3:
        print(
            "Usage: python transactionGenerator.go <outputFileName> <m: ratio of anomalous tx (per card)>"
        )
        sys.exit(1)

    output_file_name = sys.argv[1]
    try:
        anomalous_ratio = float(sys.argv[2])
        if not 0 <= anomalous_ratio <= 1:
            raise ValueError
    except ValueError:
        print(
            "Error: The <m: ratio of anomalous tx (per card)> argument must be a float in [0,1]."
        )
        sys.exit(1)

    # fix a constant seed so that experiments are reproducible
    key = 37
    np.random.seed(int(key))

    # Read the card and atm datasets
    atm_df = pd.read_csv("csv/atm.csv")
    card_df = pd.read_csv("csv/card.csv")

    print(atm_df)
    print(card_df)

    # create the transaction dataframe
    cols = [
        "transaction_id",
        "number_id",
        "ATM_id",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    transaction_df = pd.DataFrame(columns=cols)
    tx_id = 0

    for card_index in card_df.index:
        # atm_non_rgular: is the set of atms not selected for the generated tx of the card since distance <= max_distance
        tx_card, tx_id, atm_regular, atm_non_regular = transaction_generator(
            card_df.iloc[card_index], atm_df, start_date, tx_id
        )
        if len(tx_card) > 0:
            # Introduction of anomalous tx
            introduce_anomalous_fp_1(
                tx_card, anomalous_ratio, atm_regular, atm_non_regular
            )
            print(
                "########################################################################################################################################"
            )

        # if transaction_df is empty (on first iteration) then directly assign the returned df, otherwise an ordinary concat
        # Drop all-NaN rows from tx_card before concatenation
        tx_card_cleaned = tx_card.dropna(how="all").dropna(axis=1, how="all")
        transaction_df = (
            tx_card_cleaned.copy()
            if transaction_df.empty
            else pd.concat([transaction_df, tx_card_cleaned], ignore_index=True)
        )

    # print(transaction_df)
    # NOTE: We want the stream of transactions to come ordered by tx_end, which is the time
    # in which the tx finished and therefore when we simulate that it reaches the query engine
    # - order by the times they finished and therefore reached the system
    # sort by transaction_end, and if equal (if ties) by transaction_start in ascending order
    transaction_df = transaction_df.sort_values(
        by=["transaction_end", "transaction_start"], ascending=True
    ).reset_index(drop=True)

    # print(transaction_df)
    transaction_df.to_csv("csv/" + output_file_name + ".csv", index=False)

    print("\n")
    print("~~~~~~~~~~~~~~~~~~~ Summary ~~~~~~~~~~~~~~~~~~~~~")
    print(
        f"Number of transactions created:                                       {tx_id}"
    )
    print(
        f"Number of Cards with success:                                         {success_cards}"
    )
    print(
        f'Number of Cards with "failure" (no transactions could be generated):  {fail_cards}'
    )
    print("~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")


if __name__ == "__main__":
    main()
