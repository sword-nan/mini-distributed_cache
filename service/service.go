package service

import (
	"context"
	"distributed_cache/cache"
	"distributed_cache/common"
	"fmt"
	"log"
	"sync"

	"golang.org/x/sync/singleflight"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (g GetterFunc) Get(key string) ([]byte, error) {
	return g(key)
}

type Putter interface {
	Put(key string, value []byte) error
}

type PutterFunc func(key string, value []byte) error

func (p PutterFunc) Put(key string, value []byte) error {
	return p(key, value)
}

type Service struct {
	name         string
	cache        cache.Cache
	getter       Getter // call when data not in cache
	putter       Putter
	newValueItem cache.NewValue // create the Value interface
	group        *singleflight.Group
}

var (
	mu sync.RWMutex
	// store many service
	groups = make(map[string]*Service)
)

func ViewServiceGroup() {
	var services []string
	for serviceName := range groups {
		services = append(services, serviceName)
	}
	fmt.Println(services)
}

// create the Service instance
func NewService(name string, getter Getter, putter Putter, newValueItem cache.NewValue, maxBytes int64, k int) *Service {
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
		putter:       putter,
		newValueItem: newValueItem,
		group:        &singleflight.Group{},
	}
	mu.Lock()
	groups[name] = service
	mu.Unlock()
	return service
}

func (h *Service) log(format string, v ...any) {
	if common.DEBUG {
		log.Printf(format, v...)
	}
}

// load data from local
func (s *Service) load(key string) ([]byte, error) {
	return s.getlocally(key)
}

// call Get method in getter interface
func (s *Service) getlocally(key string) ([]byte, error) {
	value, err := s.getter.Get(key)
	if err != nil {
		s.log("service-%s: [DB not hit], err: %v", s.name, err)
		return nil, err
	}
	s.log("service-%s: [DB hit] get the value %s of the key %s", s.name, value, key)
	s.populateCache(key, value)
	return value, nil
}

// update the cache
func (s *Service) populateCache(key string, value []byte) {
	err := s.cache.Put(key, s.newValueItem.New(value))
	if err != nil {
		s.log("service-%s: [ERROR] data[key%s] can't store in cache", s.name, key)
	}
}

// Get
func (s *Service) get(key string) ([]byte, error) {
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
	s.log("service-%s: [Cache hit] get the value %s of the key %s", s.name, value, key)
	return value, nil
}

func (s *Service) Get(key string) ([]byte, error) {
	doC := s.group.DoChan(key, func() (interface{}, error) {
		return s.get(key)
	})
	ctx, cancel := context.WithTimeout(context.TODO(), common.TimeoutInterval)
	defer cancel()
	select {
	case val := <-doC:
		return val.Val.([]byte), val.Err
	case <-ctx.Done():
		s.log("service-%s: Get key %s timeout", s.name, key)
		// dead lock!
		go func() {
			<-doC
		}()
		return nil, common.ErrTimeout
	}
}

// Put
func (s *Service) Put(key string, value []byte) error {
	// may be not consistent
	var err error
	err = s.putter.Put(key, value)
	if err != nil {
		return err
	}
	// s.log("service-%s: put [%s, %v] in putter", s.name, key, value)
	err = s.cache.Put(key, s.newValueItem.New(value))
	if err != nil {
		return err
	}
	s.log("service-%s: put [%s, %v] in cache", s.name, key, value)
	return nil
}

func (s *Service) ViewCache() {
	s.cache.View()
}

func GetService(name string) (*Service, error) {
	mu.RLock()
	defer mu.RUnlock()
	Service, ok := groups[name]
	if !ok {
		// err := fmt.Errorf("the name [%s] not in the Service group", name)
		return nil, common.ErrServiceNotExisted
	}
	return Service, nil
}
