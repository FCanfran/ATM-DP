import os
import csv

NUM_CARDS = 500000

# List of directories containing the files
directories = [
    "./1c",
    "./2c",
    "./4c",
    "./8c",
    "./16c",
]

header = ["txFile", "test", "approach", "maxFilterSize"]

# Define the new value for the txFile column
new_txfile_name = "../input/medium/15-0.03.csv"
new_test = "15-0.03"
cores = [1, 2, 4, 8, 16]
# filters = [1, 5, 10, 100, 250, 500, 1000, 2000, 5000, 10000, 50000]
filters = [20, 40, 200]

# Iterate over each directory
for core in cores:
    # 1. create directory
    directory = str(core) + "c"
    os.makedirs(directory, exist_ok=True)
    print(f"Processing directory: {directory}")
    # For each core, create a file for each filter
    for filter in filters:
        filename = directory + "-" + str(filter) + "f"
        file_path = os.path.join(directory, filename)

        row = [
            new_txfile_name,  # tx_file
            f"{new_test}-{core}c",  # test
            f"{core}c-{filter}f",  # approach
            f"{NUM_CARDS // filter}",  # maxFilterSize
        ]
        # Write the CSV file
        with open(file_path + ".csv", mode="w", newline="") as file:
            writer = csv.writer(file)
            # Write the header
            writer.writerow(header)
            # Write the row
            writer.writerow(row)
