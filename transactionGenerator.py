import pandas as pd
import random

# months: number of months for which we generate transactions
def transaction_generator(card, atm_df, months):

    num_days = 2
    for day in num_days:

        # random number of transactions on this day:
        # poisson of lambda = 


def main():
    # Read the card and atm datasets
    atm_df = pd.read_csv('atm.csv')
    card_df = pd.read_csv('card.csv')

    print(atm_df)
    print(card_df)

    



if __name__ == "__main__":
    main()