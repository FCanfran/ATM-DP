
# Creation process of the Bank (Graph) Database

# 1. Generation of the synthetic data

Note that since for the population of the database, typically is seen that is done through:
- CSV tables imports.
- With scripts to create with cypher commands all the nodes, relations... all the database.

For simplicity and to do it in a more stepwise manner, we are going to first create all the CSV data tables for the nodes and for the relations in the corresponding format and then we will populate the Neo4j GDB with those.

## Generating the synthetic data

**For the generation of random geolocations**:
Bounding box of the cities:
Approaches:
 - Easy and less accurate -> obtain a rectangular bounding box of the city  
 - Hard and more accurate -> obtain the exact bounding box (exact polygon box) of the city  
   open street maps (like in CSN project)
 
then, in any case, drawn a random geolocation from the bounding box.

It was seen that the first approach is not really good/accurate and that it can be improved
possibly by the second one.

Useful links:

- https://nominatim.openstreetmap.org/ui/search.html
- http://bboxfinder.com/ 


### ATM

- ATM_id: string
- loc_latitude: float
- loc_longitude: float
- city: string
- country: string

Note: had to replace Emohua by Emuoha 

Idea: explained in the code...

TODO: Take into account that for each ATM location we have x number of atms... have this into account for the density distribution 
from which we drawn the city location of the new generated ATM??

Generation of `n` ATMs given the geographical distribution of the ATMs in the wisabi dataset. On it there are 50 ATMs distributed along Nigerian cities. The distribution of the ATMs matches the importance of the location since the number of ATMs is larger in the most populated Nigerian cities (30% of the ATM locations are in Lagos, then the 20% in Kano...)

Therefore, for generating a new ATM location first we select uniformly at random an ATM location from the wisabi dataset, and taking its city we produce a new random geolocation belonging to that particular city. 

This is done by:
- First constructing a geolocation dictionary of the wisabi cities, where we get for each city its geographical bounding box. 
- Then, for a particular ATM to be generated, a random ATM is selected from the wisabi dataset.
- A new ATM is produced by generating a random geolocation inside the bounding box of the city location of the selected wisabi ATM.

Note that: 

- We do not take into account for the density distribution of the ATMs of the wisabi dataset that, for each ATM location of the dataset, we have x number of atms.
--> (?) Have this into account for the density distribution from which we drawn the city location of the new generated ATM??

### Bank

- name: string
- code: int
- loc_latitude: float
- loc_longitude: float

* coordinates

### Card

- number_id: string * 
- client_id: string * 
- expiration: date
- CVC: int
- extract_limit: float * 
- loc_latitude: float * 
- loc_longitude: float *

expiration and CVC -> not relevant: could be empty fields indeed or for all the Cards the same values.

#### Some remarks

Initially:
- 1 card per client --> TODO: Later we can modify

- For the moment, for gathering the behavior of the wisabi clients we only consider the withdrawal type of transaction.

- In the generation of each card/client we gather information about their
transactional behavior (amount avg, amount std, number of transactions per day...) based on the clients of the wisabi dataset: this will be useful for the generation of the syntethic trasactions.

Note that:
  - This behavior is gathered from 1 random client at a time of the wisabi dataset, so that we have more variability. But it could be that we assign the same behavior to all the clients, and this behavior be like a summary of all the wisabi dataset clients behavior. 
  Also the behavior could be assigned drawning it from taylored distributions selected by us ("homemade").

### Transaction

- transaction_id: string
+ number_id (card id)
+ ATM_id    (atm id)
- transaction_start: datetime
- transaction_end: datetime
- transaction_amount: float

#### Remarks:

First we are generating not fraud kinds of transactions. Therefore the transactions generated for each client have to be generated in such a way that they wont produce any fraud pattern alert. Later in the process we will poison our system by generating those transactions to produce fraud pattern alerts.

- We generate transactions for a `d` number of days:

- For each day generate `num_tx` transactions, random number drawn from a Poisson distribution
of `lambda` = `withdrawal_day` (= avg number of withdrawals per day).

- transaction_start & transaction_end:

  - transaction_start: draw the time in seconds in that particular day from a normal distribution of mean = 86400/2 and std = 20000
  This choice aims at simulating the fact that most transactions occur during the day, around noon (12h). 24h x 60 min x 60 s = 86400s in a day and half day (noon) = 86400s / 2.

  - transaction_end: increment some diff/delta time based on the normal duration of a transaction.
  TODO: Define this better
  For the moment the difference is drawn from a normal distribution of mean = 5min (300s), std = 2min (120s), whenever the difference is negative then it is assigned the mean (300s).

- transaction_amount: based on card behavior params `amount_avg` and `amount_std`
  it is drawn from a normal distribution of mean = amount_avg, std = amount_std.
  If negative amount, drawn from a uniform distribution(0,amount_avg*2)

- transaction_id: based on an initial id and then increment it whenever a new transaction is created.


TODO: 
- ATM: get an ordered list of all the ATMs ordered by distance to the client and select randomly with more probability (most of it) among the closest ones!.
- Note that: we also need to take into account the time wrt to the previous transaction to allow less or more distance... -> IDEA: do a simple approximate calculation for this that is 
encoded by the value of the THRESHOLD distance at each moment.

Option 1: Not taking into account the previous generated transaction
--> (*) Option 2: Taking into account the previous generated transaction: both for the linked ATM of the new transaction and its transaction time (to avoid transactions that are overlapped or that come one directly after the other --> this may be fraudulent - AVOID!) 

https://stackoverflow.com/questions/51918580/python-random-list-of-numbers-in-a-range-keeping-with-a-minimum-distance

DISTANCE...

Useful links:
- Explanation of geodesic distance: https://michaelminn.net/tutorials/gis-distance/
- Used lirbary for the calculation of the geodesic distance: https://pypi.org/project/geopy/ 

Optional: limit to the ones that lie inside a specific distance threshold
2 approaches for the distance:
- Haversine: (great-circle distance) Earth as a sphere. Less accurate. Less expensive computation.
- Vicenty: Earth as a ellipsoid (oblate spheroid). More accurate. More expensive computation.
NOTE that: Earth is neither perfectly spherical nor ellipse hence calculating the distance on its surface is a challenging task.

https://www.neovasolutions.com/2019/10/04/haversine-vs-vincenty-which-is-the-best/


## Updates: corrections

*02/07*

### 1 client - N cards  

NOTE: For the moment leave it like 1 client:1 card (It does not make much difference at this point!)
----------------------------------------------------------------------------------
Allow for 1 client - N cards. So far for simplicity it was decided to determine the number of cards per each of the clients by drawing this number from a Poisson distribution with lambda equal to 1.

TODO: Change in the overleaf. Also update that now we are not necesarily generating num_cards but cards for num_clients such that for each client we have a certain number of cards coming from a Poisson distribution of lambda = mean_cards.
----------------------------------------------------------------------------------




