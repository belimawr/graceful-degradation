Graceful Degradation and Timeout
================================
This code was used to present a talk on Berlin Golang Meetup on 07/08/2018 about
**Graceful Degradation**. The example problem is an API that needs to fetch data
from another upstream API, process, cache and return processed data.

This repository contains a full-working example to the problem described below,
however it still in active development/improvemet.

The Problem
-----------
* Given a list of team ids
* Fetch name and country in a fixed set of languages
* Return the list for each language
* 404s must be ignored
* Other errors must be informed/fail the request
* Hard deadline/request timeout after a few seconds

There are two solutions implemented:
* **Naive:**
  * Parse the ids
  * Fetch one after another
  * Build and return the response
* **Cached:**
  * Parse the ids
  * Concurrently fetch from cache (Redis)
      * If cache miss or error, go to upstream API
  * Concurrently fetch remaining data from upstream API
  * Save on cache
  * Build and return the response

Those solutions are implemented on those endpoints:
There are two endpoints on the application:
* `/naive?ids=<comma,separated,IDs,list>`: All data is sequentially fetched
  * Example request: `http://localhost:3000/naive?ids=113,79,1797,100,200`
* `/cached?ids=<comma,separated,IDs,list>`: All data returned by the upstrem API
is cached on redis.
  * Example request: `http://localhost:3000/cached?ids=113,79,1797,100,200`

The cenario for the graceful degradation is to have some network issues in the
communication with the redis, on those cases go to the upstream API and return
a successful response for the user. 

The next sections explain how to run the code and how to use
[toxiporxy](#toxiproxy) to tamper with the network.
The ["Example Cenario"](#examplecenario) section contains a step-by-step of
how to run the cenario

Running the code
----------------
There are two dependencies to run this code:
* **Redis:** the easiest way is to run a redis on Docker:
`docker run --rm -p 6379:6379 -d redis`
* **The upstream API:** the default configuration already points to a working API 
you can also overide it with the flag `-scoresURL`

```shell
// Install dependencies
$ dep ensure

// Run application
$ go run main.go

// Help - listing all configurable parameters
$ go run main.go --help
```

Toxiproxy
---------
Toxiporxy is the tool we will use to tamper with the network conditions. According
ot it's [GitHub](https://github.com/Shopify/toxiproxy) toxiporxy is: 

> Toxiproxy is a framework for simulating network conditions. It's madespecifically to work in testing, CI and development environments, supporting deterministic tampering with connections, but with support for randomized chaos and customization.

Refer to its [documentation](https://github.com/Shopify/toxiproxy) for
instalation and usage details.

### Quick usage of toxiproxy on CLI
```
// Start the toxiproxy server. This will block the terminal
$ toxiproxy-server

// Create a proxy called redis listening on port 26379 and proxing to port 6379
$ toxiproxy-cli create redis -l localhost:26379 -u localhost:6379

// List all proxies
$ toxiproxy-cli list

// Deswcribe a proxy and its toxics
$ toxiproxy-cli inspect redis

// Add a toxic called latency
$ toxiproxy-cli toxic add redis --toxicName="latency" -t latency -a latency=310

// Update the latency toxic
$ toxiproxy-cli toxic update --toxicName latency  -a latency=290 -a jitter=50 redis
```

Example Cenario
---------------

```
// Start a Redis server using docker
$ docker run --rm -p 6379:6379 -d redis

// Start the toxiproxy server (do it on a separated terminal)
$ toxiproxy-server

// Create a proxy
$ toxiproxy-cli create redis -l localhost:26379 -u localhost:6379

// Start the application (do it on another terminal)
$ go run main.go -redisURL="localhost:26379"

// Make a test request to naive endpoint to ensure everything is working
$ curl -s -w "\nTotal Time: %{time_total}\n"  http://localhost:3000/naive\?ids\=113,79,1797,100,200

// Make a request to the cached endpoint (do it at least twice and observe the
difference on response time and logs)
$ curl -s -w "\nTotal Time: %{time_total}\n"  http://localhost:3000/cached\?ids\=113,79,1797,100,200

// Tamper with the network by adding a toxic on the proxy. Now all calls to Redis will fail
$ toxiproxy-cli toxic add redis --toxicName="latency" -t latency -a latency=310

// Make a few requests again and look at the logs and the difference on response time
$ curl -s -w "\nTotal Time: %{time_total}\n"  http://localhost:3000/cached\?ids\=113,79,1797,100,200

// Add some varition on the toxic so it will fail some requests, but not all
$ toxiproxy-cli toxic update --toxicName latency  -a latency=290 -a jitter=50 redis

// Make a few requests again and look at the logs to see how it affected the communication with the cache
$ curl -s -w "\nTotal Time: %{time_total}\n"  http://localhost:3000/cached\?ids\=113,79,1797,100,200
```
