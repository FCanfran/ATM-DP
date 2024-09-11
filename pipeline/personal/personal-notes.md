

# 1. Adapting transaction flux

- So far: 1 filter per card. Later it could be done so that we allow to have multiple different cards per filter.


# 2. Volatile subgraph storage (in each filter)

Options:

- [Golang graph library](https://pkg.go.dev/github.com/dominikbraun/graph#section-readme)
- Easy and cheap approach: linked list. To be able to append at the end and also delete from the begining of the list.
- [] *slice* approach: (**Not efficient**) a slice is implemented as a dynamically-sized array. When you delete an element from the beginning of a slice, it can be inefficient because it requires shifting all the remaining elements to the left. This shifting operation has a time complexity of O(n), where n is the number of elements in the slice.

- (*) [Linked list](https://pkg.go.dev/container/list): it is implemented in golang as a doubly linked list. It is the preferred option. *See the notebook description*


# 3. Connection with Neo4j static GDB

# 4. Fraud pattern detection

## 4.1. Pattern 1: Two or more withdrawals with the same card at ATMs from different locations (at the same time)

### Technical issues

Implementation of the easiest described fraud pattern.
- Implement 
- Introduce anomalous transactions - toy example cases to test
- Explain in the report memory

More details: 
- Do with connection to cluster Neo4j gdb, but maybe initially to the local gdb?

**Pending/to investigate:**

- [ ] Way to do Connection to the gdb on each thread - DriverWithContext VS Sessions:
 - DriverWithContext: Immutable, thread-safe, and fairly expensive to create, so your application should only create one instance
 - Sessions: created with the method DriverWithContext.NewSession(). Session creation is a lightweight operation, so sessions can be created and destroyed without significant cost. Always close sessions when you are done with them. However, they are **not thread safe**: you can share the main DriverWithContext object
across threads, but make sure each routine creates its own sessions.

Ideally we want:
- DriverWithContext object: only 1 instance (in the connection module) and shared across all the threads.
- Sessions: 1 per filter

**Therefore, investigate to do it with sessions instead on each of the golang threads (on each filter)**

~~So far, creating 1 session per filter.~~ Only create a session when querying/connecting to the database (in CheckFraud() function)

**Considerations:**

- Session Timeout: Neo4j sessions may time out after a period of inactivity, but it's best to explicitly close them when no longer needed.
- **Concurrency**: If multiple filters are running in parallel, each will create its own session. Be careful with the **number of concurrent sessions to avoid overwhelming the database**. -> Therefore, revise this. An option will be to avoid having one session per filter and create a session **only when doing a connection (a query or whatever) to the 
database.**

### Context usage & understanding

```
func SafeConnect() {
	// root context: it has no deadline and it can not be cancelled.
	// used as the base context for connecting to the Neo4j database
	ctx := context.Background()
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}
	dbUri := os.Getenv("NEO4J_URI")
	dbUser := os.Getenv("NEO4J_USERNAME")
	dbPassword := os.Getenv("NEO4J_PASSWORD")
	driver, err = neo4j.NewDriverWithContext(
		dbUri,
		neo4j.BasicAuth(dbUser, dbPassword, ""))
	if err != nil {
		panic(err)
	}

	err = driver.VerifyConnectivity(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Connection established.")
}
```
*From Chatgpt*

In this case it does not have much complexity. In this case, since you’re using context.Background(), there is no timeout or cancellation mechanism associated with it, and the Neo4j operations will continue until they succeed or fail. We are using it as the base context for connecting to the Neo4j database.

However it could be used to allow our program to be more flexible and responsive in more advanced use cases. For example:

- Timeouts: You might want to ensure that a connection attempt doesn’t take more than 5 seconds. You could use context.WithTimeout to enforce that.
- Cancellation: If you're running many concurrent database operations, you could use context.WithCancel to cancel ongoing operations if some condition is met.

Apparently, we could directly use the root context inside the goroutines, specially if we do not need to establish any specific timeout or cancelation policy for the requests, although for better management it is not recommended. Instead creating *derived* contexts is the recommended practice, allowing a potential usage of a cancelation or timeout policy.

_____________________ 

#### More information (blogs)


