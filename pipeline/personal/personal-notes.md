

# 1. Adapting transaction flux

- So far: 1 filter per card. Later it could be done so that we allow to have multiple different cards per filter.


# 2. Volatile subgraph storage (in each filter)

Options:

- [Golang graph library](https://pkg.go.dev/github.com/dominikbraun/graph#section-readme)
- Easy and cheap approach: linked list. To be able to append at the end and also delete from the begining of the list.
- [] *slice* approach: (**Not efficient**) a slice is implemented as a dynamically-sized array. When you delete an element from the beginning of a slice, it can be inefficient because it requires shifting all the remaining elements to the left. This shifting operation has a time complexity of O(n), where n is the number of elements in the slice.

- (*) [Linked list](https://pkg.go.dev/container/list): it is implemented in golang as a doubly linked list. It is the preferred option. *See the notebook description*


# 3. Connection with Neo4j static GDB


