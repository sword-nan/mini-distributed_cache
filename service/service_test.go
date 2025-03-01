package service

import (
	"distributed_cache/cache"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
)

type Mapper struct {
	sync.RWMutex
	db map[string][]byte
}

func (m *Mapper) Get(key string) ([]byte, error) {
	m.RLock()
	v, ok := m.db[key]
	if !ok {
		msg := fmt.Sprintf("the key %s not in db", key)
		m.RUnlock()
		return nil, errors.New(msg)
	}
	m.RUnlock()
	res := make([]byte, len(v))
	copy(res, v)
	return res, nil
}

func (m *Mapper) Put(key string, value []byte) error {
	m.Lock()
	defer m.Unlock()
	m.db[key] = value
	return nil
}

func TestServerGetter(t *testing.T) {
	lruk, _ := cache.NewLRUK(10, 2)
	for i := 0; i < 3; i++ {
		byteView := cache.NewByteView([]byte(strconv.Itoa(i + 1)))
		lruk.Put(strconv.Itoa(i), byteView)
		// lruk.Get(strconv.Itoa(i))
	}
	fmt.Println(lruk)

	mapper := make(map[string][]byte)
	for i := 0; i < 10; i++ {
		mapper[strconv.Itoa(i)] = []byte(strconv.Itoa(i + 1))
	}

	m := &Mapper{
		db: mapper,
	}

	var f = cache.NewValueFunc(func(b []byte) cache.Value {
		return cache.NewByteView(b)
	})
	server := Service{
		name:         "1",
		getter:       m,
		cache:        lruk,
		newValueItem: f,
	}

	// fmt.Println(m)
	// fmt.Println(lruk)

	for i := 0; i < 3; i++ {
		for j := 0; j < 11; j++ {
			server.Get(strconv.Itoa(j))
			// fmt.Println(lruk)
			if rand.Intn(5) > 3 {
				server.Get(strconv.Itoa(j))
				server.Get(strconv.Itoa(j))
			}
		}
	}
	fmt.Println(lruk)
}

func TestServiceConcurrent(t *testing.T) {
	var f = cache.NewValueFunc(func(b []byte) cache.Value {
		return cache.NewByteView(b)
	})
	var wg sync.WaitGroup
	var nRead, nWrite = 100, 1000
	for i := 0; i < nWrite; i++ {
		wg.Add(1)
		ii := i
		go func(i int) {
			defer wg.Done()
			NewService(
				strconv.Itoa(i),
				GetterFunc(func(key string) ([]byte, error) {
					return []byte(key), nil
				}),
				PutterFunc(func(key string, value []byte) error {
					return nil
				}),
				f,
				2<<10,
				2,
			)
		}(ii)
	}

	for i := 0; i < nRead; i++ {
		wg.Add(1)
		ii := i
		go func(i int) {
			defer wg.Done()
			service, err := GetService(strconv.Itoa(i))
			if err != nil {
				fmt.Println(err)
				return
			}
			for j := 0; j < nRead*100; j++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					service.Get(strconv.Itoa(j))
				}()
			}
		}(ii)
	}
	wg.Wait()
	fmt.Println(groups)
}

func TestServiceCacheConcurrent(t *testing.T) {
	var f = cache.NewValueFunc(func(b []byte) cache.Value {
		return cache.NewByteView(b)
	})
	var wg sync.WaitGroup
	var nRead, nWrite = 10000, 10000
	var hitCount, count int32
	mapper := &Mapper{
		db: make(map[string][]byte),
	}
	service := NewService(
		"test",
		mapper,
		mapper,
		f,
		2<<5,
		2,
	)
	for i := 0; i < nWrite; i++ {
		ii := i
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			service.Put(strconv.Itoa(i), []byte(strconv.Itoa(rand.Intn(i+1))))
		}(ii)
	}

	for i := 0; i < nRead; i++ {
		ii := i
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			nReplicas := rand.Intn(2) + 1
			for j := 0; j < nReplicas; j++ {
				atomic.AddInt32(&count, 1)
				_, err := service.Get(strconv.Itoa(i))
				if err == nil {
					atomic.AddInt32(&hitCount, 1)
				}
			}
		}(ii)
	}
	wg.Wait()
	// fmt.Println(mapper)
	service.ViewCache()
	fmt.Printf("hit rate is %.2f\n", float32(hitCount)/float32(count))
	fmt.Println(count, hitCount)
}
