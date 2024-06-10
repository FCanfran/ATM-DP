import pandas as pd

def bank_generator():
    print("Bank generator")
    # 1. generate random coordinates


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
def atms_generator(n):

    # create the ATM dataframe
    cols = ['ATM_id', 'loc_latitude', 'loc_longitude', 'city', 'country']
    atm_df = pd.DataFrame(columns=cols)

    # read the csv of wisabi atms 
    atm_file = 'wisabi/atm_location lookup.csv'
    atm_df_wisabi = pd.read_csv(atm_file)
    print(atm_df_wisabi.head())
    print(len(atm_df_wisabi))

    """
    for i in range(n):
        print(i)
        ATM_id = i
        # Select random wisabi ATM to assign the location
        city = 
        loc_latitude = 0
        loc_longitude = 0

        new_atm = {
            'ATM_id':, 
            'loc_latitude':, 
            'loc_longitude':, 
            'city':, 
            'country':}
        """


def main():
    atms_generator(2)

if __name__ == "__main__":
    main()