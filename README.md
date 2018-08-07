graceful-shutdown



Commands:
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
