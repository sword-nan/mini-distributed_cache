package cache

import (
	"fmt"
	"testing"
)

type String string

func (s String) NBytes() int {
	return len(s)
}

func (s String) Bytes() []byte {
	return []byte(s)
}

func (s String) New(b []byte) Value {
	return String(b)
}

func TestLinkedListInsert(t *testing.T) {
	l := newLinkedList()
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("%d", i)
		value := String(fmt.Sprintf("%d", i+1))
		l.insert(key, value)
	}
	// fmt.Println(l)
}

func TestLinkedListRemove(t *testing.T) {
	l := newLinkedList()
	var nodes []*linkedNode
	for i := 5; i > 0; i-- {
		key := fmt.Sprintf("%d", i)
		value := String(fmt.Sprintf("%d", i+1))
		node := l.insert(key, value)
		nodes = append(nodes, node)
	}
	l.remove(nodes[0])
	fmt.Println(l)
}

func TestLinkedListMoveToHead(t *testing.T) {
	l := newLinkedList()
	var nodes []*linkedNode
	for i := 5; i > 0; i-- {
		key := fmt.Sprintf("%d", i)
		value := String(fmt.Sprintf("%d", i+1))
		node := l.insert(key, value)
		nodes = append(nodes, node)
	}
	l.moveToHead(nodes[0])
	fmt.Println(l)
}
