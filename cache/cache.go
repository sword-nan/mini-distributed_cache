package cache

import (
	"bytes"
	"errors"
	"fmt"
	"sync"
)

type Cache interface {
	Get(key string) (Value, error)
	Put(key string, value Value) error
	View()
}

type LRU struct {
	nbytes     int64
	maxBytes   int64                  // lru max size
	key2node   map[string]*linkedNode // hash map
	linkedList *linkedList            // double linkedList
	mu         sync.Mutex
}

func (lru *LRU) getVictim() *linkedNode {
	return lru.linkedList.head.prev
}

func (lru *LRU) remove(key string) {
	node := lru.key2node[key]
	lru.linkedList.remove(node)
	delete(lru.key2node, key)
	lru.nbytes -= entrySize(key, node.value)
}

func (lru *LRU) Get(key string) (Value, error) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	node, ok := lru.key2node[key]
	if !ok {
		msg := fmt.Sprintf("key [%s] not found in lru cache", key)
		return nil, errors.New(msg)
	}
	// move the node to head
	lru.linkedList.moveToHead(node)
	val := node.value
	return val, nil
}

func (lru *LRU) Put(key string, value Value) (err error) {
	lru.mu.Lock()
	defer lru.mu.Unlock()
	if entrySize(key, value) > lru.maxBytes {
		msg := "the entry size is bigger than the cache max bytes"
		err = errors.New(msg)
		return
	}
	// key in cache，just update the value
	if node, ok := lru.key2node[key]; ok {
		nbytes := entrySize(key, value) - entrySize(key, node.value)
		for lru.nbytes+nbytes > lru.maxBytes {
			victim := lru.getVictim()
			lru.remove(victim.key)
		}
		lru.nbytes += nbytes
		node.setValue(value)
		lru.linkedList.moveToHead(node)
		return
	}
	nbytes := entrySize(key, value)
	for lru.nbytes+nbytes > lru.maxBytes {
		victim := lru.getVictim()
		lru.remove(victim.key)
	}
	node := lru.linkedList.insert(key, value)
	lru.key2node[key] = node
	lru.nbytes += nbytes
	return
}

func (lru *LRU) GetCurrentBytes() int64 {
	return lru.nbytes
}

func (lru *LRU) IsEmpty() bool {
	return len(lru.key2node) == 0
}

func (lru *LRU) View() {
	fmt.Println(lru.String())
}

func NewLRU(maxBytes int64) (*LRU, error) {
	if maxBytes <= 0 {
		return &LRU{}, errors.New("lru maxBytes is must be a positive number")
	}
	return &LRU{
		maxBytes:   maxBytes,
		key2node:   make(map[string]*linkedNode),
		linkedList: newLinkedList(),
	}, nil
}

func (lru *LRU) String() string {
	var buf bytes.Buffer
	buf.WriteString("LRUCache(")
	buf.WriteString(lru.linkedList.String())
	buf.WriteString(")")
	return buf.String()
}

type LRUK struct {
	nbytes         int64
	maxBytes       int64
	k              int
	historyCounter map[string]int // record the count of the node access
	lru1           *LRU
	lru2           *LRU
	sync.Mutex
}

func NewLRUK(maxBytes int64, k int) (*LRUK, error) {
	if maxBytes <= 0 || k <= 0 {
		msg := fmt.Sprintf(
			`cache maxBytes, threshold-k must be positive, but you give the maxBytes[%d] k[%d]`,
			maxBytes,
			k,
		)
		return &LRUK{}, errors.New(msg)
	}

	cache := LRUK{}
	cache.maxBytes = maxBytes
	cache.k = k + 1
	cache.historyCounter = make(map[string]int)
	cache.lru1, _ = NewLRU(maxBytes)
	cache.lru2, _ = NewLRU(maxBytes)
	return &cache, nil
}

func (l *LRUK) getVictim() *linkedNode {
	var victim *linkedNode
	if !l.lru1.IsEmpty() {
		victim = l.lru1.getVictim()
	} else {
		victim = l.lru2.getVictim()
	}
	return victim
}

func (l *LRUK) remove(key string, value Value) {
	if l.historyCounter[key] < l.k {
		l.lru1.remove(key)
	} else {
		l.lru2.remove(key)
	}
	delete(l.historyCounter, key)
	l.nbytes -= entrySize(key, value)
}

func (l *LRUK) switchTo(key string, value Value) {
	l.lru1.remove(key)
	l.lru2.Put(key, value)
}

func (l *LRUK) incrementCount(key string, value Value) {
	l.historyCounter[key]++
	if l.historyCounter[key] == l.k {
		l.switchTo(key, value)
	}
}

func (l *LRUK) Get(key string) (Value, error) {
	l.Lock()
	defer l.Unlock()
	count, ok := l.historyCounter[key]
	if !ok {
		msg := fmt.Sprintf("the key[%s] not in the cache", key)
		return nil, errors.New(msg)
	}
	var value Value
	if count < l.k {
		value, _ = l.lru1.Get(key)
	} else {
		value, _ = l.lru2.Get(key)
	}
	l.incrementCount(key, value)
	return value, nil
}

func (l *LRUK) Put(key string, value Value) (err error) {
	l.Lock()
	defer l.Unlock()
	if entrySize(key, value) > l.maxBytes {
		err = errors.New("the entry size is bigger than the cache max bytes")
		return
	}
	if count, ok := l.historyCounter[key]; ok {
		var nodeValue Value
		var flag = false
		if count < l.k {
			nodeValue, _ = l.lru1.Get(key)
		} else {
			nodeValue, _ = l.lru2.Get(key)
			flag = true
		}
		nbytes := entrySize(key, value) - entrySize(key, nodeValue)
		for l.nbytes+nbytes > l.maxBytes {
			victim := l.getVictim()
			l.remove(victim.key, victim.value)
		}
		if !flag {
			l.lru1.Put(key, value)
		} else {
			l.lru2.Put(key, value)
		}
		l.incrementCount(key, value)
		l.nbytes += nbytes
		return
	}

	nbytes := entrySize(key, value)
	for l.nbytes+nbytes > l.maxBytes {
		victim := l.getVictim()
		l.remove(victim.key, victim.value)
	}
	l.lru1.Put(key, value)
	l.incrementCount(key, value)
	l.nbytes += nbytes
	return
}

func (l *LRUK) GetCurrentBytes() int64 {
	return l.nbytes
}

func (l *LRUK) Getl1CurrentBytes() int64 {
	return l.lru1.GetCurrentBytes()
}

func (l *LRUK) Getl2CurrentBytes() int64 {
	return l.lru2.GetCurrentBytes()
}

func (l *LRUK) GetK() int {
	return l.k - 1
}

func (l *LRUK) String() string {
	var buf bytes.Buffer
	buf.WriteString("LRU-K(")
	buf.WriteString("lru1: ")
	buf.WriteString(l.lru1.String())
	buf.WriteString(" lru2: ")
	buf.WriteString(l.lru2.String())
	buf.WriteString(")")
	return buf.String()
}

func (l *LRUK) View() {
	fmt.Println(l.String())
}
