package test

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"testing"
)

func BenchmarkOps(b *testing.B) {
	var nClient = 10000
	var numbers = 100
	var serviceName = "api"
	addr := "http://localhost:9999"
	b.ResetTimer()
	var wg sync.WaitGroup
	for i := 0; i < nClient; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			key := strconv.Itoa(rand.Intn(numbers))
			url := fmt.Sprintf("%s/%s?name=%s&key=%s", addr, serviceName, "test", key)
			http.Get(url)
		}()
	}
	wg.Wait()
	key := strconv.Itoa(1000)
	url := fmt.Sprintf("%s/%s?name=%s&key=%s", addr, serviceName, "test", key)
	http.Get(url)
}
