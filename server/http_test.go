package server

import (
	"context"
	"distributed_cache/cache"
	"distributed_cache/service"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	db := map[string]string{
		"Tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}
	service := service.NewService(
		"score",
		service.GetterFunc(
			func(key string) ([]byte, error) {
				value, ok := db[key]
				if !ok {
					msg := fmt.Sprintf("the key %s not in db", key)
					return nil, errors.New(msg)
				}
				return []byte(value), nil
			}),
		service.PutterFunc(func(key string, value []byte) error {
			return nil
		}),
		cache.NewValueFunc(func(b []byte) cache.Value {
			return cache.NewByteView(b)
		}),
		2<<10,
		2,
	)
	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				fmt.Println("out")
				return
			case <-time.NewTimer(time.Second).C:
				service.ViewCache()
			}
		}
	}(ctx)

	addr := "localhost:8888"
	server := NewHTTPPool(addr)
	http.ListenAndServe(addr, server)
}
