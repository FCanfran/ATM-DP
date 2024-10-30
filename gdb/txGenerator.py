import pandas as pd
import numpy as np
import datetime
from geopy.distance import geodesic, great_circle
import sys
from bitarray import bitarray

# Transaction generator with anomalous transaction generation, given by parameter ratio [0,1], which defines
# the number of anomalous tx introduced per card (# anomalous tx of card_i = ratio * # tx of card_i)
# Every transaction consists of 2 "edges"/"tx": the starting and the ending transaction.
# NOTE: Have the same tx_id for the 2 edges (the start and the end one)

# Parameters
# --------------------------------------------------------------------------
start_date = "2018-04-01"  # start date, from which the first transaction is generated
num_days = 5  # num of days for which transactions are generated (init start_date)
# TODO: Improve this - values are "testing" values
max_size_atm_subset = 10  # maximum size of the ATM subset
# TODO: Calculate the max distance of each of the ATMs subsets
max_distance = 30  # maximum distance of the atms in the ATM subset to client residence
# duration of a transaction
# TODO: Define this better
max_duration = 600  # max duration - 600s (10min)
mean_duration = 300  # mean duration - 300s (5min)
std_duration = 120  # std duration - 120s  (2min)

# TODO: Adjust
# anomalous_tx_duration
anomalous_tx_duration = 5  # segs

########### Speeds ###########
# SPEED: for the creation of the regular tx
# MAX_SPEED: for the creation of the anomalous tx
# Needed to calculate the t_min: time needed to traverse the distance between 2 geographical points at SPEED km/h
# -> speed at which we consider the client travels NORMALY (by any means of transport) between 2 points
SPEED = 50  # km/h
# Assumption on the maximum speed (km/h) at which the distance between two geographical points
# can be traveled
MAX_SPEED = 500  # km/h
# --------------------------------------------------------------------------

# ------------------
# number of cards for which no transactions can be generated due to the specific required conditions
# -> for example: empty ATM subset (since the distance of the residence to the closest ATM is > max_distance)
fail_cards = 0
success_cards = 0
# regular and anomalous tx counters
total_regular = 0
total_anomalous = 0


# 2 approaches for the distance:
# - Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
# - Vicenty: Earth as a ellipsoid (oblate spheroid). More accurate. More expensive computation.
# NOTE that: Earth is neither perfectly spherical nor ellipse hence calculating the distance on its surface is a challenging task.
# https://www.neovasolutions.com/2019/10/04/haversine-vs-vincenty-which-is-the-best/
# Haversine
# Specific function for the ATM dataframe
def calculate_distance(atm_row, point):
    atm_loc = (atm_row["loc_latitude"], atm_row["loc_longitude"])
    distance = great_circle(atm_loc, point).kilometers
    return round(distance, 3)  # limit to 3 decimals only, km and meters


# - Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
# point: (latitude, longitude)
def calculate_distance_points(point_1, point_2):
    distance = great_circle(point_1, point_2).kilometers
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
                    global total_regular
                    total_regular += 1

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


# Generation of anomalous tx to cause the fraud pattern 1
# Per each of the generated card tx
def introduce_anomalous_fp_1(
    regular_tx_card, ratio, atm_regular, atm_non_regular, tx_id
):
    num_regular = len(regular_tx_card)
    num_anomalous = round(num_regular * ratio)
    print("..........................................")
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
    # create anomalous transaction dataframe
    cols = [
        "transaction_id",
        "number_id",
        "ATM_id",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    anomalous_df = pd.DataFrame(columns=cols)
    while anomalous < num_anomalous:
        print("................... ANOMALOUS: ", anomalous, "...................")
        # random hole selection in [0, num_regular-1]
        hole_index = np.random.randint(0, num_regular)
        if holes[hole_index] == 0:
            # not occupied, mark as occupied
            holes[hole_index] = 1
            # ................................ #
            # TODO: While loop in case we cant meet the conditions for the first selected random ATM from the atm_non_regular subset?
            # introduce anomalous tx in this position: after the tx[index] (and before tx[index+1], in case it exists tx[index+1] (tx_next) ????)
            tx_prev = regular_tx_card.iloc[hole_index]

            print("----------------------- prev -----------------------")
            print(tx_prev)
            if hole_index + 1 < num_regular:
                print("----------------------- next -----------------------")
                print(regular_tx_card.iloc[hole_index + 1])

            # select one ATM at random from atm_non_regular
            rand_index = np.random.choice(atm_non_regular.index)
            ATM_new = atm_non_regular.loc[rand_index]

            ATM_prev = (
                atm_regular.loc[atm_regular["ATM_id"] == tx_prev["ATM_id"]]
            ).iloc[0]

            ATM_prev_loc = (ATM_prev["loc_latitude"], ATM_prev["loc_longitude"])
            ATM_new_loc = (ATM_new["loc_latitude"], ATM_new["loc_longitude"])

            # Calculate t_min(ATM_prev, ATM_new)
            # 1. Calculate the distance between the 2 ATM locations (Haversine distance)
            # 2. t = e / v ---> (km)/(km/h) --> in seconds (*60*60)
            distance_km = calculate_distance_points(ATM_prev_loc, ATM_new_loc)
            t_min = int((distance_km / MAX_SPEED) * 60 * 60)  # in seconds
            # Make t_diff(tx_prev, tx_new) < t_min(ATM_prev, ATM_new)
            # tx_new.start = tx_prev.end + s_time s.t. s_time < t_min (in seconds) & s_time > 0
            # (so that tx_new.start > tx_prev.end)
            # take a random number of seconds s_time in [1, t_min)
            s_time = np.random.randint(1, t_min)
            tx_new_start = tx_prev["transaction_end"] + datetime.timedelta(
                seconds=s_time
            )

            # transaction_end:
            # tx_end = tx_start + anomalous_tx_duration segs, for all the anomalous tx
            tx_new_end = tx_new_start + datetime.timedelta(
                seconds=anomalous_tx_duration
            )

            # create the tx and insert it in the dataframe
            tx_new = {
                "transaction_id": tx_id,
                "number_id": tx_prev["number_id"],  # card id
                "ATM_id": ATM_new["ATM_id"],
                "transaction_start": tx_new_start,
                "transaction_end": tx_new_end,
                "transaction_amount": tx_prev["transaction_amount"] * 2,
            }

            print("==================== tx_new ======================")
            print(tx_new)

            tx_new_df = pd.DataFrame([tx_new])
            anomalous_df = (
                tx_new_df.copy()
                if anomalous_df.empty
                else pd.concat([anomalous_df, tx_new_df], ignore_index=True)
            )
            tx_id += 1
            anomalous += 1
            global total_anomalous
            total_anomalous += 1

    return anomalous_df, tx_id


# Splits the tx of a dataframe, so that from each tx, 2 edges are generated: tx_start & tx_end
def split_tx(tx_df):

    new_tx = []  # list of new rows, later converted to df

    # Create a new df, such that, for each tx we generate 2, 1 for tx_start and 1 for tx_end
    for _, tx in tx_df.iterrows():
        tx_start = tx.copy()
        tx_start["transaction_end"] = None
        tx_start["transaction_amount"] = None

        tx_end = tx.copy()

        new_tx.append(tx_start)
        new_tx.append(tx_end)

    new_tx_df = pd.DataFrame(new_tx)
    return new_tx_df


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
    anomalous_df = pd.DataFrame(columns=cols)
    tx_id = 0

    for card_index in card_df.index:
        # atm_non_rgular: is the set of atms not selected for the generated tx of the card since distance <= max_distance
        tx_card, tx_id, atm_regular, atm_non_regular = transaction_generator(
            card_df.iloc[card_index], atm_df, start_date, tx_id
        )
        if len(tx_card) > 0:
            # Generation of anomalous tx for this card
            card_anomalous_df, tx_id = introduce_anomalous_fp_1(
                tx_card, anomalous_ratio, atm_regular, atm_non_regular, tx_id
            )

            # Ensure the df is not empty and does not contain only NaN values, to avoid warnings
            if not card_anomalous_df.dropna(how="all").empty:
                anomalous_df = (
                    card_anomalous_df.copy()
                    if anomalous_df.empty
                    else pd.concat([anomalous_df, card_anomalous_df], ignore_index=True)
                )

            print(
                "########################################################################################################################################"
            )

        # if transaction_df is empty (on first iteration) then directly assign the returned df, otherwise an ordinary concat
        # Drop all-NaN rows from tx_card before concatenation:
        # Ensure the df is not empty and does not contain only NaN values, to avoid warnings
        if not tx_card.dropna(how="all").empty:
            transaction_df = (
                tx_card.copy()
                if transaction_df.empty
                else pd.concat([transaction_df, tx_card], ignore_index=True)
            )

    # 3 csv generated:
    # - regular tx
    # - anomalous tx
    # - all tx

    # Custom sorting logic:
    # - If tx_end is None use tx_start.
    # - Otherwise, use tx_end.

    if total_regular > 0:

        # Split the tx in 2: tx_start and tx_end
        transaction_df_ext = split_tx(transaction_df)
        transaction_df_ext["sort_key"] = transaction_df_ext.apply(
            lambda row: (
                row["transaction_end"]
                if pd.notna(row["transaction_end"])
                else row["transaction_start"]
            ),
            axis=1,
        )
        # Sort based on the custom sort_key column
        transaction_df_ext = transaction_df_ext.sort_values(
            by="sort_key", ascending=True
        ).reset_index(drop=True)

        if total_anomalous > 0:
            # Split the tx in 2: tx_start and tx_end
            anomalous_df_ext = split_tx(anomalous_df)
            anomalous_df_ext["sort_key"] = anomalous_df_ext.apply(
                lambda row: (
                    row["transaction_end"]
                    if pd.notna(row["transaction_end"])
                    else row["transaction_start"]
                ),
                axis=1,
            )

            anomalous_df_ext = anomalous_df_ext.sort_values(
                by="sort_key", ascending=True
            ).reset_index(drop=True)

            # Join regular & anomalous
            all_tx_ext = pd.concat(
                [transaction_df_ext, anomalous_df_ext], ignore_index=True
            )
            # sort after joining both
            all_tx_ext = all_tx_ext.sort_values(
                by=["sort_key"], ascending=True
            ).reset_index(drop=True)
            # Drop the sort_key column
            anomalous_df_ext = anomalous_df_ext.drop(columns=["sort_key"])
            # Write csv
            anomalous_df_ext.to_csv(
                "csv/tx/" + output_file_name + "-anomalous.csv", index=False
            )
        else:
            all_tx_ext = transaction_df_ext

        all_tx_ext = all_tx_ext.drop(columns=["sort_key"])
        all_tx_ext.to_csv("csv/tx/" + output_file_name + "-all.csv", index=False)

        transaction_df_ext = transaction_df_ext.drop(columns=["sort_key"])
        transaction_df_ext.to_csv(
            "csv/tx/" + output_file_name + "-regular.csv", index=False
        )

    else:
        print("No transactions generated\n")

    print("\n")
    print("~~~~~~~~~~~~~~~~~~~ Summary ~~~~~~~~~~~~~~~~~~~~~")
    print(
        f"Total number of transactions created:                                 {tx_id}"
    )
    print(
        f"Number of Regular | Anomalous transactions created:                   {total_regular, total_anomalous}"
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
