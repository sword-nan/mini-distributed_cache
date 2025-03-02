package main

import (
	"distributed_cache/cache"
	"distributed_cache/common"
	"distributed_cache/master"
	"distributed_cache/server"
	"distributed_cache/service"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

// simulate the database
var db = make(map[string]string)
var numbers = 100

func NewCacheService(addr string, serviceName string) {
	fmt.Printf("cache service [%s] is running at [%s]\n", serviceName, addr)
	service.NewService(
		serviceName,
		service.GetterFunc(func(key string) ([]byte, error) {
			// simulate the long time waiting
			time.Sleep(10 * time.Millisecond)
			value, ok := db[key]
			if !ok {
				// msg := fmt.Sprintf("the key %s not in db", key)
				// errors.New(msg)
				return nil, common.ErrKeyNotInDB
			}
			return []byte(value), nil
		}),
		service.PutterFunc(func(key string, value []byte) error {
			return nil
		}),
		cache.NewValueFunc(func(b []byte) cache.Value {
			return cache.NewByteView(b)
		}),
		int64(common.CacheCapacity),
		2,
	)
	server := server.NewHTTPPool(addr)
	log.Fatal(http.ListenAndServe(addr, server))
}

func genDataInDB() {
	for i := 0; i < numbers; i++ {
		db[strconv.Itoa(i)] = strconv.Itoa(i + 1)
	}
}

func main() {
	var (
		port    string
		isCache bool
	)
	flag.StringVar(&port, "port", "8001", "service port")
	flag.BoolVar(&isCache, "cache", true, "cache or master?")
	flag.Parse()
	genDataInDB()

	port2addr := map[string]string{
		"8001": "localhost:8001",
		"8002": "localhost:8002",
		"8003": "localhost:8003",
		"8004": "localhost:8004",
	}
	if isCache {
		NewCacheService(port2addr[port], "test")
	} else {
		addr := "localhost:" + port
		master := master.NewMaster(3, nil)
		var addrs []string
		for _, addr := range port2addr {
			addrs = append(addrs, addr)
		}
		master.Register("http://", "/_Cache/", addrs...)
		http.Handle("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			serviceName := r.URL.Query().Get("name")
			key := r.URL.Query().Get("key")
			value, err := master.Get(serviceName, key)
			if err == nil {
				w.Write(value)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}))
		fmt.Printf("api service [master] is running at [%s]\n", addr)
		http.ListenAndServe(addr, nil)
	}
}
