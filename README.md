# Graceful Degradation and Timeout

This code was used to present a talk on Berlin Golang Meetup on 07/08/2018.

This project still a working in progress and there will improvements on it.

## Running
```
// Install dependencies
$ dep ensure

// Run application
$ go run main.go

// Help
$ go run main.go --help

// Building
$ go build
```

## Toxiproxy
Toxiproxy is a framework for simulating network conditions/issues, here is it's GitHub:
https://github.com/Shopify/toxiproxy

**Some example of how to use it:**
```
$ toxiproxy-cli create redis -l localhost:26379 -u localhost:6379

$ toxiproxy-cli list

$ toxiproxy-cli toxic add redis -t latency -a latency=1000

$ toxiproxy-cli toxic update -n latency_downstream  -a latency=110 redis
```


## HTTP
```
GET http://localhost:8474/proxies
GET http://localhost:8474/proxies/redis
GET http://localhost:8474/proxies/redis/toxics/latency_downstream
```
