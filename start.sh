#!/bin/bash
trap "rm distributed_cache;kill 0" EXIT
go build -o distributed_cache
./distributed_cache -port=8001 &
./distributed_cache -port=8002 &
./distributed_cache -port=8003 &
./distributed_cache -port=8004 &
./distributed_cache -port=9999 -cache=false

# sleep 2
# echo ">>> start test"
# curl "http://localhost:9999/api?name=test&key=Tom" &
# curl "http://localhost:9999/api?name=test&key=Tom" &
# curl "http://localhost:9999/api?name=test&key=Tom" &
# sleep 1
