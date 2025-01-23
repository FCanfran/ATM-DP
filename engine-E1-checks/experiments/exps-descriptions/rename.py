import os
import pandas as pd

# List of directories containing the files
directories = [
    "./1c",
    "./2c",
    "./4c",
    "./8c",
    "./16c",
]

# Define the new value for the txFile column
new_txfile_name = "../input/medium/15-0.03.csv"
new_test = "15-0.03"
cores = [1, 2, 4, 8, 16]
filters = [1, 5, 10, 100, 250, 500, 1000, 2000, 5000, 10000, 50000]

# Iterate over each directory
for directory in directories:
    print(f"Processing directory: {directory}")
    # Iterate over all files in the current directory
    for filename in os.listdir(directory):
        if filename.endswith(".csv"):  # Process only CSV files
            file_path = os.path.join(directory, filename)

            # Load the CSV file
            try:
                df = pd.read_csv(file_path)
            except Exception as e:
                print(f"Error reading {filename} in {directory}: {e}")

            new_test_name = new_test + "-" + directory.replace("./", "")

            # txFile column
            df["txFile"] = new_txfile_name
            # test column
            df["test"] = new_test_name

            # Save the updated DataFrame back to the file
            try:
                df.to_csv(file_path, index=False)
                print(f"Updated txFile in {filename} (Directory: {directory})")
            except Exception as e:
                print(f"Error saving {filename} in {directory}: {e}")
