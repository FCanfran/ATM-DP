import pandas as pd
import numpy as np
import datetime
from geopy.distance import geodesic, great_circle
import sys
from bitarray import bitarray
import random
import bisect
import math
import os
from tqdm import tqdm

# Transaction generator with anomalous transaction generation, given by parameter ratio [0,1], which defines
# the number of anomalous tx introduced per card (# anomalous tx of card_i = ratio * # tx of card_i)
# Every transaction consists of 2 "edges"/"tx": the starting and the ending transaction.
# NOTE: Have the same tx_id for the 2 edges (the start and the end one)

# Operation types:
# 0: withdrawal
# 1: deposit
# 2: balance inquiry
# 3: transfer
OP_TYPES = [0, 1, 2, 3]

# Parameters
#############################################################################################################
START_DATE = "2018-04-01"  # start date, from which the first transaction is generated
NUM_DAYS = 5  # num of days for which transactions are generated (init START_DATE)

ANOMALOUS_RATIO_1 = (
    0.02  # ratio of anomalous tx (per card) over the total amount of generated transactions
    # argument must be a float in [0,1]
)

MAX_SIZE_ATM_SUBSET_RATIO = 0.2  # ratio [0,1] of the total size of the ATM global set - maximum size of the ATM subset: |ATM_subset| = ratio * |ATM|
MAX_DISTANCE_SUBSET_THRESHOLD = (
    70  # maximum distance of the atms in the ATM subset to client residence
)
MAX_DURATION = 600  # max duration of a transaction - 600s (10min)
MEAN_DURATION = 300  # mean duration of a transaction- 300s (5min)
STD_DURATION = 120  # std duration of a transaction - 120s  (2min)
REGULAR_SPEED = 50  # (km/h) REGULAR_SPEED: for the creation of the regular tx
# - needed to calculate the t_min_subset: time needed to traverse the distance between 2 geographical points at SPEED km/h
# - speed at which we consider the client travels NORMALY (by any means of transport) between 2 points
ANOMALOUS_SPEED = 500  # (km/h)  ANOMALOUS_SPEED: Assumption on the maximum ANOMALOUS speed (km/h) at which the distance between two geographical points
# can be traveled
ANOMALOUS_TX_DURATION = 5  # (segs)
#############################################################################################################


# Counters
#############################################################################################################
# number of cards for which no transactions can be generated due to the specific required conditions
# -> for example: empty ATM subset (since the distance of the residence to the closest ATM is > max_distance)
fail_cards = 0  # number of cards for which no transactions can be generated due to the specific required conditions
# -> for example: empty ATM subset (since the distance of the residence to the closest ATM is > max_distance)
success_cards = 0
total_regular = 0  # regular transactions counter
total_anomalous = 0  # anomalous transactions counter
#############################################################################################################


# 2 approaches for the distance:
# - (*) Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
# - Vicenty: Earth as a ellipsoid (oblate spheroid). More accurate. More expensive computation.
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


# max_distance between any pair of ATMs belonging to an ATM subset
def calculate_max_distance_subset(atm_df_regular):
    max_distance = 0.0

    for i in range(len(atm_df_regular)):
        for j in range(i + 1, len(atm_df_regular)):
            i_loc = (
                atm_df_regular.loc[i]["loc_latitude"],
                atm_df_regular.loc[i]["loc_longitude"],
            )
            j_loc = (
                atm_df_regular.loc[j]["loc_latitude"],
                atm_df_regular.loc[i]["loc_longitude"],
            )
            dist_i_j = calculate_distance_points(i_loc, j_loc)
            if dist_i_j > max_distance:
                max_distance = dist_i_j

    return math.ceil(max_distance)


# Get ordered ascending list by distance of the atms wrt. card location coordinates
# - Filter those that lie inside a specific distance threshold MAX_DISTANCE_SUBSET_THRESHOLD
# - Filter a maximum of ATMs: max size of the subset -> |ATM_subset| = MAX_SIZE_ATM_SUBSET_RATIO * |ATM|
def get_ordered_atms(card_loc_latitude, card_loc_longitude, atm_df):
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
    # The "regular" subset: select those with distance <= MAX_DISTANCE_SUBSET_THRESHOLD
    atm_df_regular = atm_df_ordered[
        atm_df_ordered["distance"] <= MAX_DISTANCE_SUBSET_THRESHOLD
    ]
    # The "non-regular" subset: the rest
    atm_df_non_regular = atm_df_ordered[
        atm_df_ordered["distance"] > MAX_DISTANCE_SUBSET_THRESHOLD
    ]

    # max size of the subset -> |ATM_subset| = MAX_SIZE_ATM_SUBSET_RATIO * |ATM|
    max_size_subset = math.ceil(MAX_SIZE_ATM_SUBSET_RATIO * len(atm_df))
    # print(f"max_size_subset {max_size_subset}, |ATM|={len(atm_df)}")
    atm_df_regular = atm_df_regular.head(max_size_subset)

    return atm_df_regular, atm_df_non_regular


# Distribute n transactions over the interval of all the given days
# Returns a ordered list of start moments in seconds, respecting that all of the moments
# are at a minimum time distance of t_min_subset
def distribute_tx(n, t_min_subset):
    # in seconds of a day: (86400s in a day)
    lower_bound = 0
    upper_bound = (86400 * NUM_DAYS) - 1

    num_holes = upper_bound - lower_bound
    needed_holes = (MAX_DURATION + t_min_subset) * (n - 1) + MAX_DURATION
    # print(f"num_holes: {num_holes}, needed_holes: {needed_holes}")

    if num_holes < needed_holes:
        raise ValueError(
            f"Impossible to distribute {n} transactions over the given interval time with t_min_subset = {t_min_subset}"
        )

    tx_ordered_times = []
    while len(tx_ordered_times) < n:
        start_time = int(np.random.uniform(lower_bound, upper_bound))
        diff_end = int(np.random.normal(MEAN_DURATION, STD_DURATION))
        if diff_end <= 0:
            diff_end = MEAN_DURATION  # if negative or 0 -> then it is = to the mean
        if diff_end > MAX_DURATION:
            diff_end = MAX_DURATION  # if above MAX_DURATION -> then MAX_DURATION

        end_time = start_time + diff_end
        candidate_tx = (start_time, end_time)

        def get_start(element):
            return element[0]

        def get_end(element):
            return element[1]

        # Check with previous and next
        # Find the insertion index
        index = bisect.bisect_left(
            tx_ordered_times, get_start(candidate_tx), key=get_start
        )
        # Access the previous element if it exists
        prev = tx_ordered_times[index - 1] if index > 0 else None
        # Access the next element if it exists
        next = tx_ordered_times[index] if index < len(tx_ordered_times) else None
        # Check if insertion is possible with prev and next
        if (
            prev == None or get_end(prev) + t_min_subset < get_start(candidate_tx)
        ) and (next == None or get_end(candidate_tx) + t_min_subset < get_start(next)):
            # Insert in this position
            bisect.insort(tx_ordered_times, candidate_tx)

    return tx_ordered_times


def transaction_generator(card, atm_df, tx_id):

    # print(f"------------- Generation for card: {card['number_id']} -------------")
    # create transaction dataframe
    cols = [
        "transaction_id",
        "number_id",
        "ATM_id",
        "transaction_type",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    transaction_df = pd.DataFrame(columns=cols)

    start_datetime = datetime.datetime.strptime(START_DATE, "%Y-%m-%d")

    # 1. Ordered list of terminals by ascending distance to the client card location
    # selecting a maximum of max_size_atm_subset of ATMs that are at a distance
    # inferior or equal to max_distance to the residence of the client
    atm_df_regular, atm_df_non_regular = get_ordered_atms(
        card["loc_latitude"],
        card["loc_longitude"],
        atm_df,
    )

    if len(atm_df_regular) > 0:
        # Calculate max_distance_subset for each specific ATM_subset
        max_distance_subset = calculate_max_distance_subset(atm_df_regular)
        # t_min_subset: minimum threshold time in between 2 transactions of this client
        # - based on the max distance between any pair of atms of the subset list
        # Therefore we set the t_min_subset approx to be the time needed to traverse that max_distance at REGULAR_SPEED km/h
        # if |ATM_subset| = 0 -> max_distance_subset = 0 -> t_min_subset = 0
        # t_min_subset = int(((max_distance_subset * 2) / REGULAR_SPEED) * 60 * 60)  # in seconds
        t_min_subset = int(
            (max_distance_subset / REGULAR_SPEED) * 60 * 60
        )  # in seconds
        # print(max_distance_subset)
        # print(f"tx-generation ----- t_min_subset = {t_min_subset}")

        # Generation of transactions
        withdrawal_day = card["withdrawal_day"]
        deposit_day = card["deposit_day"]
        inquiry_day = card["inquiry_day"]
        transfer_day = card["transfer_day"]

        ops_day = withdrawal_day + deposit_day + inquiry_day + transfer_day
        num_tx = np.random.poisson(ops_day * NUM_DAYS)

        op_type_probabilities = [
            withdrawal_day / ops_day,
            deposit_day / ops_day,
            inquiry_day / ops_day,
            transfer_day / ops_day,
        ]

        if num_tx > 0:
            # distributed transaction start moments (in seconds)
            tx_times = distribute_tx(num_tx, t_min_subset)
            for tx_time in tx_times:
                # 0. ATM id
                # randomly among the subset of ATMs -> all of them satisfy the constraints
                # of the min threshold time TMIN etc...
                rand_index = np.random.choice(atm_df_regular.index)
                ATM_id = atm_df_regular.loc[rand_index]["ATM_id"]
                # 1. transaction_start
                start_time_delta = datetime.timedelta(seconds=tx_time[0])
                transaction_start = start_datetime + start_time_delta
                # 2. transaction_end
                end_time_delta = datetime.timedelta(seconds=tx_time[1])
                transaction_end = start_datetime + end_time_delta
                # 3. transaction_type
                transaction_type = random.choices(
                    OP_TYPES, weights=op_type_probabilities
                )[0]

                # transaction_amount - depending on the type of tx
                if transaction_type == 0:  # withdrawal
                    transaction_amount = np.random.normal(
                        card["amount_avg_withdrawal"], card["amount_std_withdrawal"]
                    )
                    # If negative amount, draw from a uniform distribution
                    if transaction_amount < 0:
                        transaction_amount = np.random.uniform(
                            0, card["amount_avg_withdrawal"] * 2
                        )
                elif transaction_type == 1:  # deposit
                    transaction_amount = np.random.normal(
                        card["amount_avg_deposit"], card["amount_std_deposit"]
                    )
                    if transaction_amount < 0:
                        transaction_amount = np.random.uniform(
                            0, card["amount_avg_deposit"] * 2
                        )
                elif transaction_type == 2:  # balance inquiry
                    transaction_amount = 0.0
                elif transaction_type == 3:  # transfer
                    transaction_amount = np.random.normal(
                        card["amount_avg_transfer"], card["amount_std_transfer"]
                    )
                    if transaction_amount < 0:
                        transaction_amount = np.random.uniform(
                            0, card["amount_avg_transfer"] * 2
                        )

                transaction_amount = np.round(transaction_amount, decimals=2)

                new_tx = {
                    "transaction_id": tx_id,
                    "number_id": card["number_id"],  # card id
                    "ATM_id": ATM_id,
                    "transaction_type": transaction_type,
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
def introduce_anomalous_fp_1(regular_tx_card, atm_regular, atm_non_regular, tx_id):

    num_regular = len(regular_tx_card)
    num_anomalous = round(num_regular * ANOMALOUS_RATIO_1)
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
        "transaction_type",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    anomalous_df = pd.DataFrame(columns=cols)
    while anomalous < num_anomalous:
        # random hole selection in [0, num_regular-1]
        hole_index = np.random.randint(0, num_regular)
        if holes[hole_index] == 0:
            # not occupied, mark as occupied
            holes[hole_index] = 1
            tx_prev = regular_tx_card.iloc[hole_index]

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
            t_min = int((distance_km / ANOMALOUS_SPEED) * 60 * 60)  # in seconds

            # Generate (start, end) times - avoiding overlapping with prev and next tx
            fit_time = False
            while not fit_time:
                # Make t_diff(tx_prev, tx_new) < t_min(ATM_prev, ATM_new)
                # - tx_new.start = tx_prev.end + s_time s.t. s_time < t_min (in seconds) & s_time > 0
                # (so that tx_new.start > tx_prev.end)
                # take a random number of seconds s_time in [1, t_min)
                s_time = np.random.randint(1, t_min)
                tx_new_start = tx_prev["transaction_end"] + datetime.timedelta(
                    seconds=s_time
                )
                # - tx_new.end < tx_next.start
                # transaction_end:
                # tx_end = tx_start + ANOMALOUS_TX_DURATION segs, for all the anomalous tx
                tx_new_end = tx_new_start + datetime.timedelta(
                    seconds=ANOMALOUS_TX_DURATION
                )

                if hole_index + 1 < num_regular:
                    tx_next = regular_tx_card.iloc[hole_index + 1]
                    # Check tx_new.end < tx_next.start
                    if tx_new_end < tx_next["transaction_start"]:
                        fit_time = True

                else:
                    # no next tx
                    fit_time = True

            # transaction_type: randomly assign a type: [0,3]
            transaction_type = np.random.randint(0, 4)

            # create the tx and insert it in the dataframe
            tx_new = {
                "transaction_id": tx_id,
                "number_id": tx_prev["number_id"],  # card id
                "ATM_id": ATM_new["ATM_id"],
                "transaction_type": transaction_type,
                "transaction_start": tx_new_start,
                "transaction_end": tx_new_end,
                "transaction_amount": tx_prev["transaction_amount"] * 2,
            }

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

    if len(sys.argv) < 2:
        print("Usage: python txGenerator.py <outputFileName>")
        sys.exit(1)

    output_file_name = sys.argv[1]

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
        "transaction_type",
        "transaction_start",
        "transaction_end",
        "transaction_amount",
    ]
    transaction_df = pd.DataFrame(columns=cols)
    anomalous_df = pd.DataFrame(columns=cols)
    tx_id = 0

    # added progress bar with tqdm
    for card_index in tqdm(
        card_df.index,
        desc="Generating synthetic transaction stream for each of the cards",
    ):
        # atm_non_regular: is the set of atms not selected for the generated tx of the card since distance <= max_distance
        tx_card, tx_id, atm_regular, atm_non_regular = transaction_generator(
            card_df.iloc[card_index], atm_df, tx_id
        )

        # Introduction of anomalous
        if len(tx_card) > 0:
            # Generation of anomalous tx for this card
            card_anomalous_df, tx_id = introduce_anomalous_fp_1(
                tx_card, atm_regular, atm_non_regular, tx_id
            )

            # Ensure the df is not empty and does not contain only NaN values, to avoid warnings
            if not card_anomalous_df.dropna(how="all").empty:
                anomalous_df = (
                    card_anomalous_df.copy()
                    if anomalous_df.empty
                    else pd.concat([anomalous_df, card_anomalous_df], ignore_index=True)
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

    # Split the tx in two: tx_start and tx_end
    # Custom sorting logic:
    # - If tx_end is None use tx_start.
    # - Otherwise, use tx_end.

    if total_regular > 0:

        # Split the tx in 2: tx_start and tx_end
        transaction_df_ext = split_tx(transaction_df)
        transaction_df_ext["sort_key"] = transaction_df_ext.apply(
            lambda row: (
                row["transaction_end"]
                if pd.notna(
                    row["transaction_end"]
                )  # if tx_end is not missing (not NaN) -> use tx_end to sort, otherwise use tx_start
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
            # Create the output dir if it does not exist
            os.makedirs("tx", exist_ok=True)
            anomalous_df_ext.to_csv(
                "tx/" + output_file_name + "-anomalous.csv", index=False
            )
        else:
            all_tx_ext = transaction_df_ext

        all_tx_ext = all_tx_ext.drop(columns=["sort_key"])
        all_tx_ext.to_csv("tx/" + output_file_name + "-all.csv", index=False)

        transaction_df_ext = transaction_df_ext.drop(columns=["sort_key"])
        transaction_df_ext.to_csv(
            "tx/" + output_file_name + "-regular.csv", index=False
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
