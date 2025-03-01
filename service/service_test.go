package service

import (
	"distributed_cache/cache"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
)

type Mapper struct {
	db map[string][]byte
}

func (m *Mapper) Get(key string) ([]byte, error) {
	v, ok := m.db[key]
	if !ok {
		msg := fmt.Sprintf("the key %s not in db", key)
		return nil, errors.New(msg)
	}
	res := make([]byte, len(v))
	copy(res, v)
	return res, nil
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
