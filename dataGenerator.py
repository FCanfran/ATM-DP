import pandas as pd
import random

# https://nominatim.org/ -> open-source geocoding with OpenStreetMap data -> search API
from geopy.geocoders import Nominatim
from geopy.exc import GeocoderTimedOut, GeocoderUnavailable
import datetime

# Parameters
# --------------------------------------------------------------------------
num_ATMs = 10  # number of ATMs
# .............
num_cards = 10  # number of cards
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
    atm_dict = {}
    for i in range(0, len(atm_df_wisabi)):
        atm = atm_df_wisabi.iloc[i]
        city = atm["City"]
        country = atm["Country"]

        if city not in atm_dict:
            print("~~~~~~~~~~~~~~~~~", city, "~~~~~~~~~~~~~~~~~")
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
        print(i)

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

    print(
        "(%f, %f, %f, %f)" % (min_longitude, min_latitude, max_longitude, max_latitude)
    )

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
def atm_generator(atm_df_wisabi, n):

    # create the ATM dataframe
    cols = ["ATM_id", "loc_latitude", "loc_longitude", "city", "country"]
    atm_df = pd.DataFrame(columns=cols)

    num_atms_wisabi = len(atm_df_wisabi)

    # Create a dictionary of the geolocation bbox of the ATMs in the wisabi df
    atm_dictionary = create_atm_dictionary(atm_df_wisabi)

    # Generate the n synthetic ATMs
    for i in range(n):
        print("_____________________________________")
        ATM_id = i
        # Select random wisabi ATM to assign the location -
        # discrete value drawn from uniform distribution in range (0,num_atms_wisabi)
        rand_index = random.randint(0, num_atms_wisabi - 1)  # randint [a,b]
        rand_atm = atm_df_wisabi.iloc[rand_index]
        city = rand_atm["City"]
        country = rand_atm["Country"]

        print(city)

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

        print(new_atm)
        atm_df.loc[i] = new_atm

    return atm_df


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
    print(f"# of transactions: {len(transactions)}")
    # withdrawals only
    transactions = transactions[(transactions["TransactionTypeID"] == 1)]
    print(f"# of withdrawals: {len(transactions)}")

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


def card_generator(customers_df_wisabi, atm_df_wisabi, atm_df, loc_from_wisabi, n):

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
    num_customers_wisabi = len(customers_df_wisabi)

    # Generate the n synthetic cards
    for i in range(n):
        # Select random wisabi customer
        # discrete value drawn from uniform distribution in range (0,num_customers_wisabi)
        rand_index = random.randint(0, num_customers_wisabi - 1)  # randint [a,b]
        rand_customer = customers_df_wisabi.iloc[rand_index]
        print(rand_customer)

        # 1. Behavior
        # Get behaviour of this customer so that we also assign them to the new card/client generated
        # --> behavior obtained from its transactions in the wisabi dataset
        # NOTE: For the moment we only consider the withdrawal (1) type of transaction in the behavior
        behavior = get_client_behavior_wisabi(rand_customer)

        # 2. Location

        # Option 1: assign a random location of the usual ATM of the selected wisabi customer
        if loc_from_wisabi:
            atmid = rand_customer["ATMID"]
            print(atmid)
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

        print(city)
        # optional -> use the previously constructed bbox atm_dictionary of wisabi, so that it is faster
        loc_latitude, loc_longitude = generate_random_geolocation_city(
            city, country, atm_dictionary=None
        )

        new_card = {
            "number_id": i,
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

    return card_df


def main():
    # Pre: read wisabi dataset info
    # read the csv of wisabi atms
    atm_file = "wisabi/atm_location lookup.csv"
    atm_df_wisabi = pd.read_csv(atm_file)
    # read the csv of wisabi customers
    customers_file = "wisabi/customers_lookup.csv"
    customers_df_wisabi = pd.read_csv(customers_file)

    # ATM generator
    atm_df = atm_generator(atm_df_wisabi, num_ATMs)
    print(atm_df)
    atm_df.to_csv("csv/atm.csv", index=False)

    # Card generator
    card_df = card_generator(
        customers_df_wisabi, atm_df_wisabi, atm_df, loc_from_wisabi, num_cards
    )
    print(card_df)
    card_df.to_csv("csv/card.csv", index=False)


if __name__ == "__main__":
    main()
