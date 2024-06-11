import pandas as pd
import random
# https://nominatim.org/ -> open-source geocoding with OpenStreetMap data
# - search API 
from geopy.geocoders import Nominatim 

def bank_generator():
    print("Bank generator")
    # 1. generate random coordinates

# Function to generate a random geolocation (latitude, longitude) of a given
# city and country
#
# Approaches:
# - Easy and less accurate -> obtain a rectangular bounding box of the city  
# - Hard and more accurate -> obtain the exact bounding box (exact polygon box) of the city - 
#   open street maps (like in CSN project)
# 
# then, in any case, drawn a random geolocation from the bounding box


def get_city_bbox(city, country):
    # Rectangular bbox 
    geolocator = Nominatim(user_agent="canfranero")
    location = geolocator.geocode(f"{city}, {country}")
    if not location:
        raise ValueError(f"Could not geocode the city: {city} in country: {country}")
    
    bbox = location.raw['boundingbox']
    min_latitude, max_latitude, min_longitude, max_longitude = map(float, bbox)
    return min_latitude, max_latitude, min_longitude, max_longitude

def create_atm_dictionary(atm_df_wisabi): 
    atm_dict = {}
    for i in range(0,len(atm_df_wisabi)):
        atm = atm_df_wisabi.iloc[i]
        city = atm['City']
        country = atm['Country']

        if city not in atm_dict:
            print("~~~~~~~~~~~~~~~~~",city, "~~~~~~~~~~~~~~~~~")
            min_latitude, max_latitude, min_longitude, max_longitude = get_city_bbox(city, country)
            atm_dict[city] = {
                'city': city,
                'country': country,
                'bbox': {
                    'min_latitude': min_latitude, 
                    'max_latitude': max_latitude, 
                    'min_longitude': min_longitude, 
                    'max_longitude': max_longitude
                }
            }
        print(i)
    
    return atm_dict

# Generate a random geolocation inside the bbox of the given city,country
# - option: using the atm_dictionary of the atms of the cities in the wisabi dataset
def generate_random_geolocation_city(city, country, atm_dictionary):

    if atm_dictionary == None:
        # obtain the bbox of the city
        min_latitude, max_latitude, min_longitude, max_longitude = get_city_bbox(city, country)
    else:
        # obtain the bbox from the atm_dictionary
        print(atm_dictionary[city])
        min_latitude = atm_dictionary[city]['bbox']['min_latitude']
        max_latitude = atm_dictionary[city]['bbox']['max_latitude']
        min_longitude = atm_dictionary[city]['bbox']['min_longitude']
        max_longitude = atm_dictionary[city]['bbox']['max_longitude']

    print(min_latitude, max_latitude, min_longitude, max_longitude)

    print("(%f, %f, %f, %f)" % (min_longitude, min_latitude, max_longitude, max_latitude))


    random_latitude = random.uniform(min_latitude, max_latitude)
    random_longitude = random.uniform(min_longitude, max_longitude)

    print(random_latitude, random_longitude)
    return random_latitude, random_longitude

"""
ATM: 
- ATM_id
- location (loc_latitude, loc_longitude)
- city 
- country

------------------------------------------------------
For each, we take a random ATM of the wisabi dataset and keep the same 
city and country, so that we keep the same location distribution of the
ATMs as in the wisabi dataset. 
- Location is taken as random coordinates belonging to the specific (city, country).
- ATM_id is assigned sequentially 
"""

"""
TODO: Take into account that for each ATM location we have x number of atms... have this into account for the density distribution 
from which we drawn the city location of the new generated ATM??  
FOR THE MOMENT IT IS NOT TAKEN INTO ACCOUNT!
"""
def atm_generator(atm_df_wisabi, n):

    # create the ATM dataframe
    cols = ['ATM_id', 'loc_latitude', 'loc_longitude', 'city', 'country']
    atm_df = pd.DataFrame(columns=cols)
    
    num_atms_wisabi = len(atm_df_wisabi)

    # Create a dictionary of the geolocation of the ATMs in the wisabi df
    # - city, country, bbox (bounding box)
    # so that the bbox of each city does not need to be retrieved for each
    # of the new ATMs generated
    atm_dictionary = create_atm_dictionary(atm_df_wisabi)

    # Generate the n synthetic ATMs
    for i in range(n):
        print("_____________________________________")
        ATM_id = i
        # Select random wisabi ATM to assign the location - 
        # discrete value drawn from uniform distribution in range (0,num_atms_wisabi) 
        rand_index = random.randint(0, num_atms_wisabi-1) # randint [a,b]
        rand_atm = atm_df_wisabi.iloc[rand_index]
        city = rand_atm['City']
        country = rand_atm['Country']

        print(city)

        loc_latitude, loc_longitude = generate_random_geolocation_city(city, country, atm_dictionary)

        new_atm = {
            'ATM_id': ATM_id, 
            'loc_latitude': loc_latitude, 
            'loc_longitude': loc_longitude, 
            'city': city, 
            'country': country
        }
    
        print(new_atm)
        atm_df.loc[i] = new_atm
    
    print(atm_df)
    atm_df.to_csv('atms.csv', index=False)

def card_generator(customers_df_wisabi, atm_df_wisabi, n):
    
    # create the card dataframe
    cols = ['number_id', 'client_id', 'expiration', 'CVC', 'extract_limit', 'loc_latitude', 'loc_longitude']
    card_df = pd.DataFrame(columns=cols)

    num_customers_wisabi = len(customers_df_wisabi)

    # Generate the n synthetic cards
    for i in range(n):
        print("_____________________________________")
        
        # Select random wisabi customer 
        # discrete value drawn from uniform distribution in range (0,num_customers_wisabi) 
        rand_index = random.randint(0, num_customers_wisabi-1) # randint [a,b]
        rand_customer = customers_df_wisabi.iloc[rand_index]
        print(rand_customer)

        # get its usual ATM to assign the location address to this card/client
        atmid = rand_customer['ATMID']
        print(atmid)
        # find the city of this atm 
        atm = atm_df_wisabi[atm_df_wisabi['LocationID'] == atmid]

        if not atm.empty:
            city = atm['City'].iloc[0]
            country = atm['Country'].iloc[0]
        else: 
            print("No matching ATM with LocationID found in ATMs table")

        print(city, country)
        # optional -> use the previously constructed atm_dictionary of wisabi, so that it is faster
        loc_latitude, loc_longitude = generate_random_geolocation_city(city, country, atm_dictionary=None)
        print(loc_latitude, loc_longitude)



def main():


    # read the csv of wisabi atms 
    atm_file = 'wisabi/atm_location lookup.csv'
    atm_df_wisabi = pd.read_csv(atm_file)
    print(atm_df_wisabi.head())
    # read the csv of wisabi customers 
    customers_file = 'wisabi/customers_lookup.csv'
    customers_df_wisabi = pd.read_csv(customers_file)
    print(customers_df_wisabi.head())



    #atm_generator(atm_df_wisabi, 2)
    card_generator(customers_df_wisabi, atm_df_wisabi, 1)

if __name__ == "__main__":
    main()