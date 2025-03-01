package service

import (
	"distributed_cache/cache"
	"fmt"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

// 接口型函数
func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

type Service struct {
	name         string
	cache        cache.Cache
	getter       Getter         // call when data not in cache
	newValueItem cache.NewValue // create the Value interface
}

var (
	mu sync.RWMutex
	// store many service
	groups = make(map[string]*Service)
)

// create the Service instance
func NewService(name string, getter Getter, newValueItem cache.NewValue, maxBytes int64, k int) *Service {
	mu.RLock()
	if _, ok := groups[name]; ok {
		panic("service is already existed")
	}
	mu.RUnlock()
	lruk, err := cache.NewLRUK(maxBytes, k)
	if err != nil {
		panic(err)
	}
	service := &Service{
		name:         name,
		cache:        lruk,
		getter:       getter,
		newValueItem: newValueItem,
	}
	mu.Lock()
	groups[name] = service
	mu.Unlock()
	return service
}

// Get service
func (s *Service) Get(key string) ([]byte, error) {
	var (
		value []byte
		err   error
	)
	cacheEntry, err := s.cache.Get(key)
	if err != nil { // cache not hit
		return s.load(key)
	}
	// cache hit
	value = cacheEntry.Bytes()
	log.Printf("[Cache hit] get the value %s of the key %s", value, key)
	return value, nil
}

// load data from local
func (s *Service) load(key string) ([]byte, error) {
	return s.getlocally(key)
}

// call Get method in getter interface
func (s *Service) getlocally(key string) ([]byte, error) {
	value, err := s.getter.Get(key)
	if err != nil {
		log.Printf("[DB not hit], err: %v", err)
		return nil, err
	}
	log.Printf("[DB hit] get the value %s of the key %s", value, key)
	s.populateCache(key, value)
	return value, nil
}

// update the cache
func (s *Service) populateCache(key string, value []byte) {
	s.cache.Put(key, s.newValueItem.New(value))
}

func (s *Service) ViewCache() {
	s.cache.View()
}

func GetService(name string) (*Service, error) {
	mu.RLock()
	defer mu.RUnlock()
	Service, ok := groups[name]
	if !ok {
		err := fmt.Errorf("the name [%s] not in the Service group", name)
		return nil, err
	}
	return Service, nil
}
