import datetime
import numpy as np
import bisect


start_date = "2018-04-01"  # start date, from which the first transaction is generated
num_days = 5  # num of days for which transactions are generated (init start_date)

# Uniform distribution on the full interval time (considering all the days together as a single interval)


def main():
    # fix a constant seed so that experiments are reproducible
    key = 37
    np.random.seed(int(key))

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

    #############################################################################
    num_holes = upper_bound - lower_bound
    needed_holes = t_min * n
    print(f"num_holes: {num_holes}, needed_holes: {needed_holes}")

    # TODO: Change -> This is not correct like this
    # - Consider as well here the max duration of the tx
    """
    if (upper_bound - lower_bound) < (n - 1) * t_min:
        raise ValueError(
            f"Impossible to distribute {n} transactions over a day with tmin = {t_min}"
        )
    """
    # Instead
    if (needed_holes) > (num_holes):
        print("It can't be fitted")
    #############################################################################

    tx_ordered_times = [(10, 11), (12, 18), (24, 30)]
    start_time = 19
    diff_end = int(np.random.normal(3, 1))
    if diff_end < 0:
        diff_end = 3  # if negative -> then it is = to the mean
    if diff_end > 10:
        diff_end = 10  # if above 10 mins -> then 10 min

    end_time = start_time + diff_end

    candidate_tx = (start_time, end_time)

    print(candidate_tx)

    def get_start(element):
        return element[0]

    def get_end(element):
        return element[1]

    # Check with previous and end
    # Find the insertion index
    index = bisect.bisect_left(tx_ordered_times, get_start(candidate_tx), key=get_start)

    # Access the previous element if it exists
    prev = tx_ordered_times[index - 1] if index > 0 else None

    # Access the next element if it exists
    next = tx_ordered_times[index] if index < len(tx_ordered_times) else None

    print(f"Previous element: {prev}")
    print(f"Next element: {next}")

    # Check if insertion is possible with prev and next
    # prev: prev.end < tx.start
    if prev != None:
        if get_end(prev) < get_start(candidate_tx)
    # next: tx.end < next.start

    """
    while len(tx_times) < n:
        candidate = int(np.random.uniform(lower_bound, upper_bound))

    """
    """
    start_datetime = datetime.datetime.strptime(start_date, "%Y-%m-%d")
    start_time_tx = 86399

    candidate = int(np.random.uniform(interval_low, interval_upper))
    start_time_delta = datetime.timedelta(seconds=start_time_tx)
    # Add the timedelta to the start date
    transaction_start = start_datetime + start_time_delta

    print(transaction_start)
    """


if __name__ == "__main__":
    main()
