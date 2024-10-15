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
