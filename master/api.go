package master

import (
	"distributed_cache/client"
	"distributed_cache/consistenthash"
	"log"
	"sync"
)

var iLog = true

type Master struct {
	sync.RWMutex
	register *consistenthash.Map
	peers    map[string]*client.Client
}

// replias: virtual peer num
// hash: hash function
func NewMaster(replias int, hash consistenthash.HashFunc) *Master {
	return &Master{
		register: consistenthash.NewMap(replias, hash),
		peers:    make(map[string]*client.Client),
	}
}

func (m *Master) log(format string, v ...any) {
	if iLog {
		log.Printf(format, v...)
	}
}

func (m *Master) direct(key string) (*client.Client, error) {
	var (
		addr string
		err  error
	)
	addr, err = m.register.Search(key)
	if err != nil {
		return nil, err
	}
	return m.peers[addr], nil
}

func (m *Master) Register(prefix string, suffix string, addrs ...string) error {
	var err error
	m.Lock()
	defer m.Unlock()
	err = m.register.Add(addrs...)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		m.peers[addr] = client.NewClient(prefix + addr + suffix)
	}
	return nil
}

func (m *Master) Delete(addrs ...string) error {
	var err error
	m.Lock()
	defer m.Unlock()
	err = m.register.Delete(addrs...)
	if err != nil {
		return err
	}
	for _, addr := range addrs {
		delete(m.peers, addr)
	}
	return nil
}

func (m *Master) Get(serviceName string, key string) ([]byte, error) {
	m.RLock()
	defer m.RUnlock()
	peer, err := m.direct(key)
	if err != nil {
		return nil, err
	}
	m.log("direct to %s", peer.ServerAddr())
	return peer.Get(serviceName, key)
}
