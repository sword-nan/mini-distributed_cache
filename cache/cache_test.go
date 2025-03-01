package cache

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
)

func transformKey(i int) string {
	return fmt.Sprintf("%d", i)
}

func transformValue(i int) String {
	return String(fmt.Sprintf("%d", i))
}

func transformKeyAndValue(i, j int) (string, String) {
	return transformKey(i), transformValue(j)
}

func TestLruSize(t *testing.T) {
	_, err := NewLRU(0)
	if err == nil {
		t.Fail()
	}

	_, err = NewLRU(10)
	if err != nil {
		t.Fail()
	}
}

func TestLruPutGet(t *testing.T) {
	lru, _ := NewLRU(10)

	for i := 0; i < 5; i++ {
		key, value := transformKeyAndValue(i, i+1)
		lru.Put(key, value)
	}
	if lru.nbytes != 5*2 {
		t.Fail()
	}

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("%d", i)
		value, err := lru.Get(key)
		if err != nil || value.(String) != String(fmt.Sprintf("%d", i+1)) {
			t.Fail()
		}
	}
}

func TestLruPutExisted(t *testing.T) {
	lru, _ := NewLRU(10)

	lru.Put("1", String("1"))
	value, err := lru.Get("1")
	if err != nil || value.(String) != String("1") {
		t.Fail()
	}
	lru.Put("1", String("2"))
	value, err = lru.Get("1")
	if err != nil || value.(String) != String("2") || lru.GetCurrentBytes() != 2 {
		t.Fail()
	}
}

func TestLruFull(t *testing.T) {
	var (
		size      int64 = 20
		insertNum       = 50
	)
	lru, _ := NewLRU(size)

	for i := 0; i < insertNum; i++ {
		key, value := transformKeyAndValue(i+10, i+11)
		lru.Put(key, value)
	}
	if lru.GetCurrentBytes() != size {
		t.Fail()
	}
}

func TestLruGetVictim(t *testing.T) {
	var (
		size      int64 = 100
		partion   int64 = 4
		insertNum       = 50
	)
	lru, _ := NewLRU(size)

	var key string
	var value Value
	for i := 0; i < insertNum; i++ {
		key, value = transformKeyAndValue(i+10, i+11)
		if int64(i+1)*partion > size {
			node := lru.getVictim()
			key = transformKey(i + 10 - int(size/partion))
			if node.key != key {
				fmt.Printf("the victim key[%s] != target key[%s]", node.key, key)
				t.Fail()
				break
			}
		}
		lru.Put(key, value)
	}
}

func TestLruSwitchGetVictim(t *testing.T) {
	var (
		size int64 = 10
	)

	lru, _ := NewLRU(size)
	var (
		key   string
		value String
	)
	key, value = transformKeyAndValue(1, 2)
	lru.Put(key, value)
	key, value = transformKeyAndValue(2, 10)
	lru.Put(key, value)
	lru.Get("1")
	node := lru.getVictim()
	if node.key != "2" {
		t.Fail()
	}
}

func TestLrukSize(t *testing.T) {
	_, err := NewLRUK(-1, 10)
	if err == nil {
		t.Fail()
	}

	_, err = NewLRUK(10, -1)
	if err == nil {
		t.Fail()
	}

	lruk, err := NewLRUK(10, 10)
	if err != nil {
		t.Fail()
	}

	if lruk.GetCurrentBytes() != 0 || lruk.GetK() != 10 {
		t.Fail()
	}
}

func TestLrukPut(t *testing.T) {
	lruk, _ := NewLRUK(10, 2)
	var (
		key   string
		value String
	)
	key, value = transformKeyAndValue(1, 2)
	lruk.Put(key, value)
	key, value = transformKeyAndValue(2, 3)
	lruk.Put(key, value)
	fmt.Println(lruk)
	key, value = transformKeyAndValue(1, 3)
	lruk.Put(key, value)
	lruk.Get(key)
	key, value = transformKeyAndValue(3, 4)
	lruk.Put(key, value)
	fmt.Println(lruk)
}

func checklrukSize(lruk *LRUK, l1Size int64, l2Size int64, n int, t *testing.T) (flag bool) {
	l1size := lruk.Getl1CurrentBytes()
	l2size := lruk.Getl2CurrentBytes()
	size := lruk.GetCurrentBytes()
	if l1size != l1Size || l2size != l2Size {
		fmt.Printf("lru1 size must be %d but get %d, lru2 size must be %d but get %d\n", l1Size, l1size, l2Size, l2size)
		t.Fail()
		return
	}
	if l1size+l2size != size {
		fmt.Printf("l1cache size[%d] + l2cache size[%d] != lruk cache size[%d]\n", l1size, l2size, size)
		t.Fail()
		return
	}

	if len(lruk.historyCounter) != n {
		fmt.Printf("the historycount length must be %d\n", size)
		t.Fail()
		return
	}
	flag = true
	return
}

func TestLrukPutGet(t *testing.T) {
	lruk, _ := NewLRUK(10, 2)
	var (
		key   string
		value String
	)
	key, value = transformKeyAndValue(1, 2)
	lruk.Put(key, value)
	// fmt.Println(lruk.GetCurrentBytes())
	key, value = transformKeyAndValue(2, 3)
	lruk.Put(key, value)
	// fmt.Println(lruk.GetCurrentBytes())
	checklrukSize(lruk, 4, 0, 2, t)
	key, value = transformKeyAndValue(1, 3)
	lruk.Put(key, value)
	lruk.Get(key)
	// fmt.Println(lruk.GetCurrentBytes())
	checklrukSize(lruk, 2, 2, 2, t)
	key, value = transformKeyAndValue(3, 4)
	lruk.Put(key, value)
	lruk.Get(key)
	checklrukSize(lruk, 4, 2, 3, t)
	fmt.Println(lruk)
}

func TestLrukGetNotExisted(t *testing.T) {
	lruk, _ := NewLRUK(10, 2)
	var (
		key   string
		value String
	)
	key, value = transformKeyAndValue(1, 2)
	lruk.Put(key, value)
	key, value = transformKeyAndValue(2, 3)
	lruk.Put(key, value)
	key = transformKey(3)
	_, err := lruk.Get(key)
	if err == nil {
		t.Fail()
	}
	fmt.Println(lruk)
}

func TestLrukPutManyTimes(t *testing.T) {
	lruk, _ := NewLRUK(10, 2)
	var (
		key   string
		value String
	)
	for i := 0; i < 999; i++ {
		key, value = transformKeyAndValue(1, i+1)
		lruk.Put(key, value)
		cacheValue, err := lruk.Get(key)
		lruk.Get(key)
		if err != nil {
			t.Fail()
		}
		if cacheValue.(String) != value {
			fmt.Printf("\n")
			t.Fail()
			break
		}
		var l2Size int64
		if i < 9 {
			l2Size = 2
		} else if i < 99 {
			l2Size = 3
		} else {
			l2Size = 4
		}
		if !checklrukSize(lruk, 0, l2Size, 1, t) {
			break
		}
	}
	fmt.Println(lruk)
}

func TestLrukGetVictim(t *testing.T) {
	var (
		size int64 = 40
		k          = 2
	)
	lruk, _ := NewLRUK(size, k)
	var (
		key   string
		value String
	)
	for i := 0; i < 10; i++ {
		key, value = transformKeyAndValue(i+10, i+11)
		lruk.Put(key, value)
	}
	var victim *linkedNode
	for i := 0; i < 10; i++ {
		victim = lruk.getVictim()
		key = transformKey(i + 10)
		if victim.key != key {
			fmt.Printf("the victiom must be key[%s], but get key[%s]\n", key, victim.key)
			t.Fail()
			break
		}
		lruk.Get(key)
	}
	for i := 0; i < 10; i++ {
		victim = lruk.getVictim()
		key = transformKey(i + 10)
		if victim.key != key {
			fmt.Printf("the victiom must be key[%s], but get key[%s]\n", key, victim.key)
			t.Fail()
			break
		}
		lruk.Get(key)
	}
	fmt.Println(lruk)
}

func BenchmarkLrukConcurrent(b *testing.B) {
	var (
		size int64 = 100
		k          = 2
	)
	lruk, _ := NewLRUK(size, k)
	var mu sync.Mutex
	hit := 0
	wg := sync.WaitGroup{}
	putGoroutines := 100
	putLoop := 10
	getGoroutines := 10
	getLoop := 1000
	for i := 0; i < putGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var (
				key   string
				value String
			)
			for i := 0; i < putLoop; i++ {
				key, value = transformKeyAndValue(rand.Intn(1000)+1, rand.Intn(100))
				lruk.Put(key, value)
			}
		}()
	}
	for i := 0; i < getGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var key string
			for i := 0; i < getLoop; i++ {
				key = transformKey(rand.Intn(1000) + 1)
				_, err := lruk.Get(key)
				if err == nil {
					mu.Lock()
					hit += 1
					mu.Unlock()
				}
			}
		}()
	}
	wg.Wait()
}
