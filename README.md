# A Simple Twitter Client/Server System


## Part 1: Twitter Feed

A disgruntled Twitter staff member deleted important software from the
servers on their last day at work. You were hired to help restore
services, by re-developing the data structure that represents a user's
feed. Your implementation will redefine it as a singly linked list.

Your task to implement the remaining incomplete methods of a `feed`
(i.e. the `Add`, `Remove`, and `Contains` methods). You must use the
internal representations for `type feed struct` and `type post struct`
in your implementation. **Do not implement a parallel version of the
linked list at this point**. You can only add fields to the struct but
cannot modify the original fields given.

Test your implementation of feed by using the test file called
`feed_test.go`. You should only run the sequential tests first:

  - `TestSimpleSeq`
  - `TestAdd`
  - `TestContains`
  - `TestRemove`

As a reminder from prior assignments, you can run individual tests by
using the `-run` flag when using `go test`. This flag takes in a regular
expression to know which tests to run. Make sure are you in the
directory that has the `*_test.go` file.

Sample run of the `SimpleSeq` test:

    //Run top-level tests matching "SimpleSeq", such as "TestSimpleSeq".
    $ go test -run "SimpleSeq"  
    PASS
    ok    hw4/feed    0.078s

Sample run of the `SimpleSeq` and `TestAdd` tests:

    //Run top-level tests matching "SimpleSeq" such as "TestSimpleSeq" and "TestAdd".
    $ go test -v -run "SimpleSeq|TestAdd"
    === RUN   TestSimpleSeq
    --- PASS: TestSimpleSeq (0.00s)
    === RUN   TestAdd
    --- PASS: TestAdd (0.00s)
    PASS
    ok    hw4/feed    0.078s

You can also run specific tests by using anchors. For example to run
only `TestAdd` then execute the following command:

    $ go test -v -run ^TestAdd$
    === RUN   TestAdd
    --- PASS: TestAdd (0.00s)
    PASS
    ok      hw4/feed      0.103s

## Part 2: Thread Safety using a Read-Write Lock

A read/write lock mechanism allows multiple readers to access a data
structure concurrently, but only a single writer is allowed to access
the data structures at a time. Implement a read/write lock library that
**only** uses a **single condition variable** and **mutex** for its
synchronization mechanisms. Go provides a Read/Write lock that is
implemented using atomics:

  - [R/W Mutex in Go](https://golang.org/pkg/sync/#RWMutex)

As with the Go implementation, you will need to provide four methods
associated with your lock:

  - `Lock()`
  - `Unlock()`
  - `RLock()`
  - `RUnlock()`

These methods should function exactly like their Go counterparts.
Documentation on each method's functionality is described by the link
provided above. You **must** limit the max number of readers to `32`.

### Coarse Grain Feed

Now, go back to your feed library inside of `feed.go` and make it
thread-safe by using your implementation of a read/write lock. You need
to think about the appropriate places to call the various read/write
locking and unlocking methods in the `feed` methods. **This must be a
coarse grain implementation**. Test that this is working by running the
remaining tests inside the `feed_test.go` file.

Helpful Resource(s):

  - [Condition Variables in
    Go](https://kaviraj.me/understanding-condition-variable-in-go/)

## Part 3: A Twitter Server

Using the `feed` library from Part 1, you will implement a server that
processes **requests**, which perform actions (e.g., add a post, remove
a post, etc.) on a single feed. These requests come from a client
program (e.g., the twitter mobile app) where a user can request to
modify their feed. The server sends **responses** back to the client
with a result (e.g., notification that a post was added and/or removed
successfully, etc.).

The client and server must agree on the format to send the requests and
responses. One common format is to use JSON. For our server, requests
and responses are single strings in the JSON format. If you are
unfamiliar with the JSON standard then you should read up on it before
beginning this portion of the assignment. JSON is a widely used
serialization format that you will mostly encounter at some point in
your career. Thus, it's better that you learn it now so you can be
comfortable with it in the future. How to work with JSON data inside Go
is described here:

  - [JSON in Go](https://blog.golang.org/json)

**Make sure you fully understand the “Streaming Encoders and Decoders”
section. You'll be using encoders and decoders in this part heavily**.

The client and server will need to use decoders and encoders to parse
the JSON string into a type easily usable in Go. You will use the
`json.Decoder`, which acts as a streaming buffer of requests, and
`json.Encoder`, which acts as a streaming buffer of responses. This
model is a simplified version of a real-life client-server model used
heavily in many domains such as web development. At a high-level, you
can think of your program as a simple “server” in the client-server
model illustrated below:

![image](./cs_model.png)

Requests (i.e., tasks in our program) are sent from a “client” (e.g., a
redirected file on the command line, a task generator program piped into
your program, etc.) via os.Stdin. The “server” (i.e., your program) will
process these requests and send their results back to the client via
os.Stdout. This model is a mimicking a real-life client-server model
used heavily in many domains such as web development; however, we are
not actually implementing web client-server system in this assignment.

### Requests and Responses Format

The basic format for the requests coming in from `json.Decoder` will be
of the following format:

``` json
{ 
"command": string, 
"id": integer, 
... data key-value pairings ... 
}
```

A request will always have a `"command"` and `"id"` key. The `"command"`
key holds a string value that represents the type of feed task. The
`"id"` represents an unique identification number for this request.
Requests are processed asynchronously by the server so requests can be
processed out of order from how they are received from `json.Decoder`;
therefore, the `"id"` acts as a way to tell the client that result
coming back from the server is a response to an original request with
this specific `"id"` value. Thus, **it is not your responsibility to
maintain this order and you must not do anything to maintain it in your
program**.

The remaining key-value pairings represent the data for a specific
request. The following subsections will go over the various types of
requests.

### Add Request

An add request adds a new post to the feed data structure. The
`"command"` value will always be the string `"ADD"`. The data fields
include a key-value pairing for the message body (`"body": string`) and
timestamp (`"timestamp": number`). For example,

``` json
{ 
  "command": "ADD", 
  "id": 342, 
  "body": "just setting up my twttr", 
  "timestamp": 43242423
}
```

After completing a `"ADD"` request, the goroutine assigned the request
will send a response back to the client via `json.Encoder` acknowledging
the add was successful. The response is a JSON object that includes a
success key-value pair (`"success": boolean`). For an add request, the
value is always true since you can add an infinite number of posts. The
original identification number should also be included in the response.
For example, using the add request shown above, the response message is

``` json
{ 
  "success": true, 
  "id": 342
}
```

### Remove Request

A remove request removes a post from the feed data structure. The
`"command"` value will always be the string `"REMOVE"`. The data fields
include a key-value pairing for the timestamp (`"timestamp": number`)
that represents the post that should be removed. For example,

``` json
{ 
  "command": "REMOVE", 
  "id": 2361, 
  "timestamp": 43242423
}
```

After completing a `"REMOVE"` task, the goroutine assigned the task will
send a response back to the client via `json.Encoder` acknowledging the
remove was successful or unsuccessful. The response is a JSON object
that includes a success key-value pair (`"success": boolean`). For a
remove request, the value is `true` if the post with the requested
timestamp was removed, otherwise assign the key to `false`. The original
identification number should also be included in the response. For
example, using the remove request shown above, the response message is

``` json
{ 
  "success": true, 
  "id": 2361
}
```

### Contains Request

A contains request checks to see if a feed post is inside the feed data
structure. The `"command"` value will always be the string `"CONTAINS"`.
The data fields include a key-value pairing for the timestamp
(`"timestamp": number`) that represents the post to check. For example,

``` json
{ 
  "command": "CONTAINS", 
  "id": 2362, 
  "timestamp": 43242423
}
```

After completing a `"CONTAINS"` task, the goroutine assigned the task
will send a response back to the client via `json.Encoder` acknowledging
whether the feed contains that post. The response is a JSON object that
includes a success key-value pair (`"success": boolean`). For a contains
request, the value is `true` if the post with the requested timestamp is
inside the feed, otherwise assign the key to `false`. The original
identification number should also be included in the response. For
example, using the contains request shown above, the response message is

``` json
{ 
  "success": false, 
  "id": 2362
}
```

**Note**: Assuming we removed the post previously.

### Feed Request

A feed request returns all the posts within the feed. The `"command"`
value will always be the string `"FEED"`. Their are no data fields for
this request. For example,

``` json
{ 
  "command": "FEED", 
  "id": 2, 
}
```

After completing a `"FEED"` task, the goroutine assigned the task will
send a response back to the client via `json.Encoder` with all the posts
currently in the feed. The response is a JSON object that includes a
success key-value pair (`"feed": [objects]`). For a feed request, the
value is a JSON array that includes a JSON object for each feed post.
Each JSON object will include a `"body"` key (`"body": string`) that
represents a post’s body and a `"timestamp"` key (`"timestamp": number`)
that represents the timestamp for the post. The original identification
number should also be included in the response. For example, assuming we
inserted a few posts into the feed, the response should look like

``` json
{ 
  "id": 2
  "feed":[
        { 
          "body": "This is my second twitter post", 
          "timestamp": 43242423
        },
        {
          "body": "This is my first twitter post", 
          "timestamp": 43242420
        }
        ]
}
```

### Done Request

If client will no longer send requests then it sends a done request. The
`"command"` value will always be the string `"DONE"`. Their are no data
fields for this request. For example,

``` json
{ 
  "command": "DONE" 
}
```

This notifies server it needs to *shutdown* (i.e., close down the
program). A done request signals to the main goroutine that no further
processing is necessary after this request is received. No response is
sent back to the client. Make sure to handle all remaining requests in
the and responses before shutting down the program.

### Implementing the Server

Inside the `server/server.go` file, you will see the following code,

``` go
type Config struct {
  Encoder *json.Encoder // Represents the buffer to encode Responses
  Decoder *json.Decoder // Represents the buffer to decode Requests
  Mode    string        // Represents whether the server should execute
                        // sequentially or in parallel
                        // If Mode == "s"  then run the sequential version
                        // If Mode == "p" then run the parallel version 
                        // These are the only values for Version
  ConsumersCount int    // Represents the number of consumers to spawn
}

//Run starts up the twitter server based on the configuration
//information provided and only returns when the server is fully
// shutdown.
func Run(config Config) {
  panic("TODO")

}
```

When a goroutine calls the `Run` function, it will start the server
based on the configuration passed to the function. Read over the
documentation above to understand the members of the configuration
struct. This function does not return (i.e., it is a blocking function)
until the server is shutdown (i.e., it receives the `"DONE"` request).
You must not modify the `Config` struct or the function header for the
`Run` function because the tests rely on this structure. The following
sections explain the modes of the server.

### Parallel Version: Tasks Queue

Inside the `Run` function, if `config.Mode == "p"` then the server will
run the parallel version. This version is implemented using a *task
queue*. This task queue is another work distribution technique and your
first exposure to the producer-consumer model. In this model, the
producer will be the main goroutine and its job is to collect a series
of tasks and place them in a queue structure to be executed by consumers
(also known as workers). The consumers will be spawned goroutines. You
**must** implement the parallelization as follows:

1.  The main goroutine begins by spawning a specified
    `config.ConsumersCount` goroutines, where each will begin executing
    a function called `func consumer(...)`. It is up to you to decide
    the arguments you pass to this function. Each goroutine will either
    begin doing work or go to sleep in a conditional wait if there is no
    work to begin processing yet. This “work” is explained in Steps 3
    and 4. **Your program cannot spawn additional goroutines after this
    initial spawning by the main goroutine**.
2.  After spawning the consumer goroutines, the main goroutine will call
    a function `func producer(...)`. Again, what you pass to this
    function is for you to decide. Inside the producer function, the
    main goroutine reads in from `json.Decoder` a series of tasks (i.e.,
    requests). For the sake of explicitness, the tasks will be feed
    operations for a single user-feed that the program manages. The main
    goroutine will place the tasks inside of a queue data structure and
    do the following:
      - If there is a consumer goroutine waiting for work then place a
        task inside the queue and wake one consumer up.
      - Otherwise, the main gorountine continues to place tasks into the
        queue. Eventually, the consumers will grab the tasks from the
        queue at later point in time.
3.  Inside the `func consumer(...)` function each consumer goroutine
    will try to grab one task from the queue. The consumer will then
    process the request and send the appropriate response back. When a
    consumer finishes executing its task, it checks the queue to grab
    another task. If there are no tasks in the queue then it will need
    to wait for more tasks to process or exit its function if there are
    no remaining tasks to complete.

### Additional Queue Requirements

You must implement this queue data structure so that both the main and
worker goroutines have access to retrieve and modify it. All work is
placed in this queue so workers can grab tasks when necessary. Along
with the requirements defined in this section, the actual enqueuing and
dequeuing of items must also be done in a unbounded lock-free manner
(i.e., non-blocking). However, the code to make the producer signal to
consumers, and consumers to wait on work must be done using a condition
variable. No busy waiting is allowed in this assignment.

You may want to separate out the queue implementation into its own
package and then have `server.go` import it. This design is up to you.
However, the producer and consumer functions must always remain in
`server.go`. I will also allow you to separate out the producer/consumer
condition variable code from the unbounded lock-free queue code. Again,
this is for you to decide.

### Sequential Version

You will need to write a sequential version of this program where the
main goroutine processes and executes all the requested tasks without
spawning any gorountines.

## Part 4: The Twitter Client

Inside the `twitter/twitter.go`, you must define a simple Twitter client
that has the following usage and required command-line argument:

``` 
Usage: twitter <number of consumers>
    <number of consumers> = the number of goroutines (i.e., consumers) to be part of the parallel version.  
```

The program needs to create a `server.Config` value based off the above
arguments. If `<number of consumers>` is not entered then this means you
need to run your sequential version of the server; otherwise, run the
parallel version. The `json.Decoder` and `json.Encoder` should be
created by using `os.Stdin` and `os.Stdout`. Please refer back to the
“Streaming Encoders and Decoders” section here: [JSON in
Go](https://blog.golang.org/json). The last call in the main function is
to start running the server. Once the `Run` function returns than the
program exits. Don't over think this implementation. It should be simple
and short.

Assumptions: No error checking is needed. All tasks read-in will be in
the correct format with all its specified data. All command line
arguments will be given as specified. You will not need to printout the
usage statement. This is shown for clarity purposes.


## SPEEDUP GRAPH
![SPPED UP GRAPH](SPEEDUP%20GRAPH.png)

* The speedup graph demonstrates that increasing the number of threads yields minimal speedup for xsmall and small test sizes, while producing significant speedup for medium, large, and xlarge test sizes.
* This result is not surprising, as smaller test sizes may not provide sufficient work for all threads, thereby limiting the potential for speedup. Additionally, there is a possibility of decreased speedup due to thread overhead.
---

### FEED

* Add()
    * find the correct order for the new post in the linked list and add it there
    * uses write lock since we are making changes to the linked list

* Remove()
    * finds the post in the linked list and removes it. If not present returns false
    * uses write lock since we are making changes to the linked list

* Contains()
    * finds the post in the linked list. Returns false if not found
    * uses read lock since no changes will be made to the linked list

* GetFeed()
    * returns a list of the posts
    * uses read lock since no changes will be made to the linked list
---

### TWITTER

* main()
    * creates config based on the command line inputs
    * calls the run function in server with a new user feed and the config
---

### SERVER

* struct ActionResponse to encode - ADD, REMOVE, CONTAINS commands
* strcut FeedResponse to encode - FEED command
* Run()
    * creates Locks Free Queue
    * makes calls to producer and consumer if mode == "s"
    * makes call to prodcuer and spawns consumer count number of consumers if mode == "p"
        * After producer exits, as long as more than 1 go routine is alive, wakes the threads up so that they can return.
* Producer()
    * decodes json requests and adds it to the queue
    * keeps looking for requests until "DONE" request is found
* Consumer
    * consumes the requests placed into the queue and generate appropriate responses and encodes into JSON
    * if the queue is empty and done command is not seen yet, the routines wait on the semaphore.
    * once the done command is seen and the queue is empty, routines start exiting consumer.

---

### Running Script to generate graph

* Go to benchmark folder and run
* ```./benchmark.sh```  or   ```sbatch benchmark.sh```

---
### ANALYSIS

* Using Lazy-list should improve performance since it allow nodes to be added or removed from the list without requiring atomic operations or any other synchronization.
* To potentially enhance the performance, instead of utilizing coarse-grained locking, we could have considered using more advanced techniques such as fine-grained locking or optimized locking methods. These techniques are known to reduce contention and enhance concurrency, thus potentially improving overall system performance.
* Yes. Hardware can have an affect on performace. If one consumer routine can finish the work faster, it can pickup more work thereby reducing number of routines needed. With more routines and better hardware, all of the work can be completed faster.

---