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

### COLAB
* Worked and Discussed with TINA OBEROI and RAJAT GUPTA