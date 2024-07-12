
# Description

This was done for the *Detection of triangles in bipartite graphs using Dynamic Pipelines*. See BipartiteDP.

- Graph: a list of Edges

- Edge: u:Card ___weight: Time___v:ATM. (It has as attribute/weight the Time in which the transaction was performed)


- in_comm: internal communication channel structure:
    --> formed by:
            - Edge channel: to pass the edges between the different components of the pipeline  
            - Front channel: to pass in_comm structures between the different components of the pipeline (it is going to be used whenever a filter is decided to be deleted, so that the pipeline flow can be reconnected properly)


## Generator

Input channels:

    - edges: (<-chan cmn.Edge) edge input channel. To receive edges. The channel is given as parameter of the generator creation function.

    - alerts: (chan cmn.Graph) to receive / pass a graph pattern that has been detected. Creates an alert.

    - front_channels: (<-chan in_comm) to receive in_comm channels, needed for the case that a filter is eliminated, to properly reconnect the pipeline.


# 1. Adapting transaction flux

- So far: 1 filter per card. Later it could be done so that we allow to have multiple different cards per filter.




# 2. Volatile subgraph storage (in each filter)

# 3. Connection with Neo4j static GDB

# 4. Filter lifetime management

