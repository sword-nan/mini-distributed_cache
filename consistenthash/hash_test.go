package consistenthash

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

func TestConsistenthashAdd(t *testing.T) {
	peerMap := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})
	var err error
	// 11, 12, 13, 21, 22, 23, 31, 32, 33
	err = peerMap.Add("1", "2", "3")
	if err != nil {
		fmt.Println("add peers error")
		t.Fail()
	}

	if peerMap.PeerCount() != 3 {
		fmt.Println("peer count is wrong")
		t.Fail()
	}

	err = peerMap.Add("1", "4", "5")
	if err == nil {
		fmt.Println("add replicate peer error")
		t.Fail()
	}

	peerMap.Add("4", "5", "6")
	if peerMap.PeerCount() != 6 {
		t.Fail()
	}
	// fmt.Printf("%+v\n", peerMap)
}

func TestConsistenthashDelete(t *testing.T) {
	peerMap := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})
	var err error
	// 11, 12, 13, 21, 22, 23, 31, 32, 33
	peerMap.Add("1", "2", "3")
	err = peerMap.Delete("4", "5")
	if err == nil {
		fmt.Println("delete the nonexisted peer error")
		t.Fail()
	}

	err = peerMap.Delete("1", "2")
	if err != nil || peerMap.PeerCount() != 1 {
		t.Fail()
	}
	// fmt.Printf("%+v\n", peerMap)
}

func TestConsistenthashSearch(t *testing.T) {
	peerMap := NewMap(3, func(data []byte) uint32 {
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})
	peerMap.Add("1", "2", "3")
	testCases := map[string]string{
		"20": "2",
		"10": "1",
		"34": "1",
		"30": "3",
	}
	for key, ans := range testCases {
		res, err := peerMap.Search(key)
		if err != nil || res != ans {
			fmt.Printf("the key [%s] get the wrong answer [%s], should be %s\n", key, res, ans)
			t.Fail()
			break
		}
	}
	// fmt.Printf("%+v\n", peerMap)
}

func TestConsistenthashConcurrent(t *testing.T) {
	peerMap := NewMap(3, nil)
	peerMap.Add("1", "2", "3")
	var wg sync.WaitGroup
	var rmu sync.RWMutex
	var nAdd, nRead, nDelete = 100, 1000, 20
	for i := 0; i < nAdd; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rmu.Lock()
			defer rmu.Unlock()
			err := peerMap.Add(strconv.Itoa(rand.Intn(1000)))
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	for i := 0; i < nRead; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rmu.RLock()
			defer rmu.RUnlock()
			_, err := peerMap.Search(strconv.Itoa(rand.Intn(1000)))
			if err != nil {
				fmt.Println(err)
			}
		}()
	}

	for i := 0; i < nDelete; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rmu.Lock()
			defer rmu.Unlock()
			err := peerMap.Delete(strconv.Itoa(rand.Intn(1000)))
			if err != nil {
				fmt.Println(err)
			}
		}()
	}
	wg.Wait()
	// fmt.Printf("%+v\n", peerMap)
}
