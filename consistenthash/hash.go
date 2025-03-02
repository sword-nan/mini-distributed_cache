package consistenthash

import (
	"distributed_cache/common"
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFunc func(data []byte) uint32

type Map struct {
	hash       HashFunc
	replicas   int                 // virtual peer num
	hashValues []int               // peers hash value in hash cycle
	hash2peer  map[int]string      // virtual peer to peer
	peers      map[string]struct{} // check if the peer has been in the map
}

func NewMap(replicas int, hash HashFunc) *Map {
	if hash == nil {
		hash = crc32.ChecksumIEEE
	}
	return &Map{
		hash:      hash,
		replicas:  replicas,
		hash2peer: make(map[int]string),
		peers:     make(map[string]struct{}),
	}
}

func (m *Map) genVirtualPeer(key string, i int) []byte {
	virtualKey := key + strconv.Itoa(i)
	return []byte(virtualKey)
}

func (m *Map) searchIdx(value int) int {
	// no result, idx will be len(m.hashValues)
	idx := sort.Search(len(m.hashValues), func(i int) bool {
		return m.hashValues[i] >= value
	})
	return idx % len(m.hashValues)
}

func (m *Map) deleteIdx(idx int) {
	var new []int
	new = append(new, m.hashValues[:idx]...)
	new = append(new, m.hashValues[idx+1:]...)
	m.hashValues = new
}

func (m *Map) Add(peers ...string) error {
	for _, peer := range peers {
		if _, ok := m.peers[peer]; ok {
			return common.ErrPeerRegistered
		}
	}
	for _, peer := range peers {
		for i := 1; i <= m.replicas; i++ {
			// virtual peer key
			virtualPeer := m.genVirtualPeer(peer, i)
			hashValue := int(m.hash(virtualPeer))
			// virtual peer to peer
			m.hash2peer[hashValue] = peer
			m.hashValues = append(m.hashValues, hashValue)
		}
		m.peers[peer] = struct{}{}
	}
	sort.Ints(m.hashValues)
	return nil
}

func (m *Map) Delete(peers ...string) error {
	// is all virtual peer existed
	for _, peer := range peers {
		if _, ok := m.peers[peer]; !ok {
			return common.ErrPeerNotRegistered
		}
	}
	// if all existed
	for _, peer := range peers {
		for i := 1; i <= m.replicas; i++ {
			virtualPeer := m.genVirtualPeer(peer, i)
			hashValue := int(m.hash(virtualPeer))
			delete(m.hash2peer, hashValue)
			idx := m.searchIdx(hashValue)
			m.deleteIdx(idx)
		}
		delete(m.peers, peer)
	}
	return nil
}

func (m *Map) Search(key string) (string, error) {
	if m.Empty() {
		return "", common.ErrNoPeerRegistered
	}
	value := m.hash([]byte(key))
	idx := m.searchIdx(int(value))
	virtualHashValue := m.hashValues[idx]
	return m.hash2peer[virtualHashValue], nil
}

func (m *Map) PeerCount() int {
	return len(m.peers)
}

func (m *Map) Empty() bool {
	return m.PeerCount() == 0
}
