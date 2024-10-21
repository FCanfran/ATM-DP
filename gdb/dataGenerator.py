import pandas as pd
import random
import datetime
from geopy.geocoders import (
    Nominatim,
)  # https://nominatim.org/ -> open-source geocoding with OpenStreetMap data -> search API
from geopy.exc import GeocoderTimedOut, GeocoderUnavailable
from tqdm import tqdm

# Parameters
# --------------------------------------------------------------------------
# num_ATMs = 10  # number of ATMs
# .............
# num_cards = 10  # number of cards
"""
NOTE: If we decide to implement 1 client : N cards
# mean number of cards per client - to decide the number of cards per each of the clients
# - coming from Poisson distribution of lambda = mean_cards
mean_cards = 1
"""
# - true -> random location is taken from the random selected customer from wisabi from which the card
# info is going to be generated (the location of his usual ATM)
# - false -> random location is taken from one of the locations of the ATMs of the newly generated ATMs dataset
loc_from_wisabi = False
# --------------------------------------------------------------------------


# Function to obtain the geographical bounding box of a given city and country
# Approaches:
# - Easy and less accurate -> obtain a rectangular bounding box of the city (*)
# - Hard and more accurate -> obtain the exact bounding box (exact polygon box) of the city -
#   open street maps (like in CSN project)
def get_city_bbox(city, country):
    # Rectangular bbox
    geolocator = Nominatim(user_agent="canfranero")

    try:
        location = geolocator.geocode(f"{city}, {country}", timeout=10)
        if not location:
            raise ValueError(
                f"Could not geocode the city: {city} in country: {country}"
            )

        bbox = location.raw["boundingbox"]
        min_latitude, max_latitude, min_longitude, max_longitude = map(float, bbox)
        return min_latitude, max_latitude, min_longitude, max_longitude

    except GeocoderTimedOut:
        print(f"Geocoding timed out for city: {city} in country: {country}")
        raise

    except GeocoderUnavailable:
        print(
            f"Geocoding service is unavailable for city: {city} in country: {country}"
        )
        raise

    except Exception as e:
        print(f"An error occurred: {e}")
        raise


# Create a dictionary of the geolocation of the ATMs in the wisabi df
# - city, country, bbox (bounding box)
# so that the bbox of each city does not need to be retrieved for each
# of the new ATMs generated
def create_atm_dictionary(atm_df_wisabi):
    print("... creation of the ATM geolocation dictionary")
    atm_dict = {}
    for i in tqdm(
        range(0, len(atm_df_wisabi)),
        desc="Creating ATM geolocation dictionary of the wisabi dataset",
    ):
        atm = atm_df_wisabi.iloc[i]
        city = atm["City"]
        country = atm["Country"]

        if city not in atm_dict:
            min_latitude, max_latitude, min_longitude, max_longitude = get_city_bbox(
                city, country
            )
            atm_dict[city] = {
                "city": city,
                "country": country,
                "bbox": {
                    "min_latitude": min_latitude,
                    "max_latitude": max_latitude,
                    "min_longitude": min_longitude,
                    "max_longitude": max_longitude,
                },
            }

    return atm_dict


# Generate a random geolocation in decimal degrees format inside the bbox of the given
# city,country
# - option: using the atm_dictionary of the atms of the cities in the wisabi dataset
def generate_random_geolocation_city(city, country, atm_dictionary):

    if atm_dictionary == None:
        # obtain the bbox of the city
        min_latitude, max_latitude, min_longitude, max_longitude = get_city_bbox(
            city, country
        )
    else:
        # obtain the bbox from the atm_dictionary
        min_latitude = atm_dictionary[city]["bbox"]["min_latitude"]
        max_latitude = atm_dictionary[city]["bbox"]["max_latitude"]
        min_longitude = atm_dictionary[city]["bbox"]["min_longitude"]
        max_longitude = atm_dictionary[city]["bbox"]["max_longitude"]

    random_latitude = random.uniform(min_latitude, max_latitude)
    random_longitude = random.uniform(min_longitude, max_longitude)
    # limit the values to have only 6 decimals - enough
    return round(random_latitude, 6), round(random_longitude, 6)


# ------------------------------------------------------
# ATM:
# - ATM_id
# - location (loc_latitude, loc_longitude)
# - city
# - country
# ------------------------------------------------------
# For each, we take a random ATM of the wisabi dataset and keep the same city and country, so
# that we keep the same location distribution of the ATMs as in the wisabi dataset.
# - Location is taken as random coordinates belonging to the specific (city, country).
# - ATM_id is assigned sequentially
# TODO: Take into account that for each ATM location we have x number of atms... have this into account for the density distribution
# from which we drawn the city location of the new generated ATM??
# FOR THE MOMENT IT IS NOT TAKEN INTO ACCOUNT!
def atm_generator(
    atm_df_wisabi, n_atms_internal, n_atms_external, bank_code, atm_dictionary
):

    # create the ATM dataframe
    cols = ["ATM_id", "loc_latitude", "loc_longitude", "city", "country"]
    atm_df = pd.DataFrame(columns=cols)
    # create the relationship ATM-bank dataframe
    cols = ["code", "ATM_id", "ownership"]
    atm_bank_df = pd.DataFrame(columns=cols)

    num_atms_wisabi = len(atm_df_wisabi)

    # Generate the n_atms_internal + n_atms_external synthetic ATMs
    for i in range(n_atms_internal + n_atms_external):
        ATM_id = (
            bank_code + "-" + str(i)
            if i < n_atms_internal
            else "EXT-" + str(i - n_atms_internal)
        )
        # Select random wisabi ATM to assign the location -
        # discrete value drawn from uniform distribution in range (0,num_atms_wisabi)
        rand_index = random.randint(0, num_atms_wisabi - 1)  # randint [a,b]
        rand_atm = atm_df_wisabi.iloc[rand_index]
        city = rand_atm["City"]
        country = rand_atm["Country"]

        loc_latitude, loc_longitude = generate_random_geolocation_city(
            city, country, atm_dictionary
        )

        new_atm = {
            "ATM_id": ATM_id,
            "loc_latitude": loc_latitude,
            "loc_longitude": loc_longitude,
            "city": city,
            "country": country,
        }

        atm_df.loc[i] = new_atm

        # Relationship
        # internal -> the first [0, n_atms_internal)
        # external -> [n_atms_internal, n_atms_internal + n_atms_external]
        ownership = True if i < n_atms_internal else False

        new_atm_bank = {
            "code": bank_code,
            "ATM_id": ATM_id,
            "ownership": ownership,
        }
        atm_bank_df.loc[i] = new_atm_bank

    return atm_df, atm_bank_df


# Different types of transactions:
# - 1: Withdrawal       (Retirada de dinero)
# - 2: Deposit          (Ingreso)
# - 3: Balance Inquiry  (Consulta de saldo/balance)
# - 4: Transfer         (Transferencia)
# NOTE: For the moment we only consider the withdrawal (1) type of transaction in the behavior
def get_client_behavior_wisabi(customer):
    # CardholderID to locate the transactions of the customer in the wisabi dataset
    # for a customer, all the transactions take place in the same atm (in the wisabi dataset)

    # CardholderID
    # -> to gather the transactions of this client
    # -> also indicates in which transaction csv we have to look into
    behavior = {}
    cardholderid = customer["CardholderID"]
    csv_code = cardholderid.split("-")[
        0
    ]  # to read the transactions from the corresponding CSV

    if csv_code == "EN":
        csv_file = "enugu_transactions.csv"
    elif csv_code == "FC":
        csv_file = "fct_transactions.csv"
    elif csv_code == "KN":
        csv_file = "kano_transactions.csv"
    elif csv_code == "LA":
        csv_file = "lagos_transactions.csv"
    elif csv_code == "RI":
        csv_file = "rivers_transactions.csv"
    else:
        print("No matching transaction file, csv code was:", csv_code)
        return

    all_transactions_df = pd.read_csv("wisabi/" + csv_file)

    # obtain all the transactions of the customer by the cardholderid
    # & that are of the type withdrawal (1)
    transactions = all_transactions_df[
        (all_transactions_df["CardholderID"] == cardholderid)
    ]
    # print(f"# of transactions: {len(transactions)}")
    # withdrawals only
    transactions = transactions[(transactions["TransactionTypeID"] == 1)]
    # print(f"# of withdrawals: {len(transactions)}")

    if not transactions.empty:
        amount_avg = round(transactions["TransactionAmount"].mean(), 2)
        amount_std = round(transactions["TransactionAmount"].std(), 2)
        # Number of transactions per day - we have transactions of the year 2022 - 365 days
        num_transacc_per_day = round(len(transactions) / 365, 4)
        behavior["amount_avg"] = amount_avg
        behavior["amount_std"] = amount_std
        behavior["transacc_day"] = num_transacc_per_day
    else:
        print("No matching transactions with CardholderID found in transactions table")

    return behavior


def card_generator(
    customers_df_wisabi, atm_df_wisabi, atm_df, loc_from_wisabi, n, bank_code
):

    # create the card dataframe
    cols = [
        "number_id",
        "client_id",
        "expiration",
        "CVC",
        "loc_latitude",
        "loc_longitude",
        "extract_limit",
        "amount_avg",
        "amount_std",
        "withdrawal_day",
    ]
    card_df = pd.DataFrame(columns=cols)
    # create the relationship card-bank dataframe
    cols = ["code", "number_id"]
    card_bank_df = pd.DataFrame(columns=cols)

    num_customers_wisabi = len(customers_df_wisabi)

    # Generate the n synthetic cards
    for i in tqdm(range(n), desc="Generating cards"):

        # Select random wisabi customer
        # discrete value drawn from uniform distribution in range (0,num_customers_wisabi)
        rand_index = random.randint(0, num_customers_wisabi - 1)  # randint [a,b]
        rand_customer = customers_df_wisabi.iloc[rand_index]

        # 1. Behavior
        # Get behaviour of this customer so that we also assign them to the new card/client generated
        # --> behavior obtained from its transactions in the wisabi dataset
        # NOTE: For the moment we only consider the withdrawal (1) type of transaction in the behavior
        behavior = get_client_behavior_wisabi(rand_customer)

        # 2. Location

        # Option 1: assign a random location of the usual ATM of the selected wisabi customer
        if loc_from_wisabi:
            atmid = rand_customer["ATMID"]
            # find the city of this atm
            atm = atm_df_wisabi[atm_df_wisabi["LocationID"] == atmid]
            if not atm.empty:
                city = atm["City"].iloc[0]
                country = atm["Country"].iloc[0]
            else:
                print("No matching ATM with LocationID found in ATMs table")

        # Option 2: assign a random location of the city of a random ATM of the newly generated ATMs
        else:
            # Get a random ATM from the new ATM table
            rand_atm_index = random.randint(0, len(atm_df) - 1)
            city = atm_df.iloc[rand_atm_index]["city"]
            country = atm_df.iloc[rand_atm_index]["country"]

        # optional -> use the previously constructed bbox atm_dictionary of wisabi, so that it is faster
        loc_latitude, loc_longitude = generate_random_geolocation_city(
            city, country, atm_dictionary=None
        )

        # id
        number_id = "c-" + bank_code + "-" + str(i)

        new_card = {
            "number_id": number_id,
            "client_id": i,
            "expiration": datetime.date.today(),
            "CVC": 999,
            "loc_latitude": loc_latitude,
            "loc_longitude": loc_longitude,
            # NOTE: Temporary approach
            "extract_limit": round(behavior["amount_avg"] * 5, 2),
            # optional -> for the generation of the transactions based on the behavior
            # of the clients of the wisabi dataset
            "amount_avg": behavior["amount_avg"],
            "amount_std": behavior["amount_std"],
            "withdrawal_day": behavior["transacc_day"],
        }

        card_df.loc[i] = new_card

        # Relationship
        new_card_bank = {
            "code": bank_code,
            "number_id": number_id,
        }

        card_bank_df.loc[i] = new_card_bank

    return card_df, card_bank_df


def main():
    # Pre: read wisabi dataset info

    # read the csv of wisabi atms
    atm_file = "wisabi/atm_location lookup.csv"
    atm_df_wisabi = pd.read_csv(atm_file)
    # read the csv of wisabi customers
    customers_file = "wisabi/customers_lookup.csv"
    customers_df_wisabi = pd.read_csv(customers_file)

    # Create a dictionary of the geolocation bbox of the ATMs in the wisabi df
    atm_dictionary = create_atm_dictionary(atm_df_wisabi)

    # Read the bank csv - to do the relationships between
    # bank - ATM
    # bank - Card
    banks_file = "csv/bank.csv"
    bank_df = pd.read_csv(banks_file)
    # For the bank, introduce the number of ATMs and Cards that we want to generate
    print(f"Introduce the desired number of ATMs and Cards to be generated")

    while True:
        n_atms_internal = input(f"\t-> Number of internal ATMs: ")
        try:
            n_atms_internal = int(n_atms_internal)
            break  # exit True loop if inputs are valid
        except ValueError:
            print("Input has to be an integer!")

    while True:
        n_atms_external = input(f"\t-> Number of external ATMs: ")
        try:
            n_atms_external = int(n_atms_external)
            break  # exit True loop if inputs are valid
        except ValueError:
            print("Input has to be an integer!")

    while True:
        n_cards = input(f"\t-> Number of Cards: ")
        try:
            num_cards = int(n_cards)
            break  # exit True loop if inputs are valid
        except ValueError:
            print("Input has to be an integer!")

    # For the bank we generate the desired number of ATMs and Cards, and create
    # the corresponding relationships Bank-ATM, Bank-Card
    bank_code = bank_df.iloc[0]["code"]
    # ATM generator: returns the ATM df and the ATM-bank relationship dataframe
    atm_df, atm_bank_df = atm_generator(
        atm_df_wisabi, n_atms_internal, n_atms_external, bank_code, atm_dictionary
    )

    print(atm_df)
    print(atm_bank_df)

    atm_df.to_csv("csv/atm.csv", index=False)
    atm_bank_df.to_csv("csv/atm-bank.csv", index=False)

    print(f"ATMs correctly generated.")
    print()

    print(f"Card generation process...")
    # Card generation
    bank_code = bank_df.iloc[0]["code"]
    # Card generator: returns the card df and the card-bank relationship dataframe
    card_df, card_bank_df = card_generator(
        customers_df_wisabi,
        atm_df_wisabi,
        atm_df,
        loc_from_wisabi,
        num_cards,
        bank_code,
    )

    print(card_df)
    print(card_bank_df)

    card_df.to_csv("csv/card.csv", index=False)
    card_bank_df.to_csv("csv/card-bank.csv", index=False)
    print(f"Cards correctly generated.")


if __name__ == "__main__":
    main()
