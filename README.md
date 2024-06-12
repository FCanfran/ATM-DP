
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


Initially:
- 1 card per client --> TODO: Later we can modify

#### Approach 1: 1 client at-a-time



