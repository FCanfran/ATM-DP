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
# - Hard and more accurate -> obtain the exact bounding box of the city - open street maps (like in CSN project)
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



def generate_random_geolocation_city(city, country, atm_dictionary):

    # obtain the bbox from the atm_dictionary
    print(atm_dictionary[city])
    min_latitude, max_latitude, min_longitude, max_longitude = 0,0,0,0
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
def atms_generator(n):

    # create the ATM dataframe
    cols = ['ATM_id', 'loc_latitude', 'loc_longitude', 'city', 'country']
    atm_df = pd.DataFrame(columns=cols)

    # read the csv of wisabi atms 
    atm_file = 'wisabi/atm_location lookup.csv'
    atm_df_wisabi = pd.read_csv(atm_file)
    print(atm_df_wisabi.head())
    
    num_atms_wisabi = len(atm_df_wisabi)

    # Create a dictionary of the geolocation of the ATMs in the wisabi df
    # - city, country, bbox (bounding box)
    # so that the bbox of each city does not need to be retrieved for each
    # of the new ATMs generated
    atm_dictionary = create_atm_dictionary(atm_df_wisabi)
    """
    for i in range(n):
        print("_____________________________________")
        ATM_id = i
        # Select random wisabi ATM to assign the location - 
        # discrete value drawn from uniform distribution in range (0,num_atms_wisabi) 
        rand_index = random.randint(0, num_atms_wisabi-1) # randint [a,b]
        rand_atm = atm_df_wisabi.iloc[rand_index]
        city = rand_atm['City']
        country = rand_atm['Country']

        loc_latitude, loc_longitude = generate_random_geolocation_city(city, country, atm_dictionary)

        new_atm = {
            'ATM_id': ATM_id, 
            'loc_latitude': loc_latitude, 
            'loc_longitude': loc_longitude, 
            'city': city, 
            'country': country
        }
    
        print(new_atm)
    """

def main():
    atms_generator(1)

if __name__ == "__main__":
    main()