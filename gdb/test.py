import datetime
import numpy as np
import bisect
import random
from bitarray import bitarray

max_duration = 600  # max duration - 600s (10min)
mean_duration = 300  # mean duration - 300s (5min)
std_duration = 120  # std duration - 120s  (2min)
start_date = "2018-04-01"  # start date, from which the first transaction is generated
num_days = 5  # num of days for which transactions are generated (init start_date)

# Uniform distribution on the full interval time (considering all the days together as a single interval)


def main():

    anomalous = 0
    num_anomalous = 2
    num_regular = 4
    holes = bitarray(num_regular)
    holes.setall(0)

    while anomalous < num_anomalous:
        print("................... ANOMALOUS: ", anomalous, "...................")
        # random hole selection in [0, num_regular-1]
        hole_index = np.random.randint(0, num_regular)
        if holes[hole_index] == 0:
            # not occupied, mark as occupied
            holes[hole_index] = 1
            tx_prev = hole_index
            print(f"tx_prev: {tx_prev}")

            # Generate (start_time,end_time)
            fit_time = False
            while not fit_time:
                # Generate s_time
                # start and end time
                # check end time - only if there is a next tx
                if hole_index + 1 < num_regular:
                    tx_next = hole_index + 1
                    print(f"tx_next: {tx_next}")
                    # check that tx_end < next.start
                    if true:
                        fit_time = True
                    else: 
                        # try to assign another time
                else:
                    fit_time = True  # no next tx

    exit(1)

    # fix a constant seed so that experiments are reproducible
    key = 37
    np.random.seed(int(key))

    # TODO: start_datetime by input parameter
    start_datetime = datetime.datetime.strptime(start_date, "%Y-%m-%d")

    # TODO: Adjust better
    SPEED = 50  # km/h
    t_min = int(((30 * 2) / SPEED) * 60 * 60)  # in seconds

    print(f"t_min: {t_min}")

    # random number of transactions on this day:
    # poisson of lambda = withdrawal_day (= avg number of withdrawals per day)
    # num_tx = np.random.poisson(card["withdrawal_day"])
    # n = np.random.poisson(lam=2.5 * num_days)
    # TODO: Provisional
    n = int(input("Enter n: "))
    num_days = int(input("Enter num_days: "))

    # print(f"num_tx_total: {n}, num_tx_total")

    # full interval: [0, 86400*num_days)
    lower_bound = 0
    upper_bound = (86400 * num_days) - 1

    num_holes = upper_bound - lower_bound
    needed_holes = (max_duration + t_min) * (n - 1) + max_duration
    print(f"num_holes: {num_holes}, needed_holes: {needed_holes}")

    if num_holes < needed_holes:
        raise ValueError(
            f"Impossible to distribute {n} transactions over the given interval time with tmin = {t_min}"
        )

    tx_ordered_times = []
    n = 5
    while len(tx_ordered_times) < n:
        start_time = int(np.random.uniform(lower_bound, upper_bound))
        diff_end = int(np.random.normal(mean_duration, std_duration))
        if diff_end < 0:
            diff_end = mean_duration  # if negative -> then it is = to the mean
        if diff_end > max_duration:
            diff_end = max_duration  # if above max_duration -> then max_duration

        end_time = start_time + diff_end
        candidate_tx = (start_time, end_time)
        print(candidate_tx)

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

        print(f"Previous element: {prev}")
        print(f"Next element: {next}")

        # Check if insertion is possible with prev and next
        if (prev == None or get_end(prev) + t_min < get_start(candidate_tx)) and (
            next == None or get_end(candidate_tx) + t_min < get_start(next)
        ):
            # Insert in this position
            bisect.insort(tx_ordered_times, candidate_tx)

        print(tx_ordered_times)

    print(tx_ordered_times)

    for tx in tx_ordered_times:
        print(tx)
        start_time_delta = datetime.timedelta(seconds=tx[0])
        # Add the timedelta to the start date
        transaction_start = start_datetime + start_time_delta
        end_time_delta = datetime.timedelta(seconds=tx[1])
        transaction_end = start_datetime + end_time_delta
        print(transaction_start, transaction_end)


if __name__ == "__main__":
    main()
