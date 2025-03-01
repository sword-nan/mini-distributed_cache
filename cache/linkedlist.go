package cache

import (
	"bytes"
	"fmt"
)

type entry struct {
	key   string
	value Value
}

func entrySize(key string, value Value) int64 {
	return int64(len(key)) + int64(value.NBytes())
}

// double linked node
type linkedNode struct {
	entry
	prev *linkedNode
	next *linkedNode
}

func (l *linkedNode) String() string {
	return fmt.Sprintf("Node[%s, %v]", l.key, l.value)
}

func (l *linkedNode) Value() Value {
	return l.value
}

func (l *linkedNode) setValue(value Value) {
	l.value = value
}

func newLinkedNode(key string, value Value) *linkedNode {
	return &linkedNode{
		entry: entry{
			key:   key,
			value: value,
		},
	}
}

// 链表
type linkedList struct {
	head *linkedNode
}

func newLinkedList() *linkedList {
	head := newLinkedNode("head", nil)
	head.next = head
	head.prev = head
	return &linkedList{
		head: head,
	}
}

func (l *linkedList) String() string {
	if l.head.next == l.head {
		return ""
	}
	var p = l.head.next
	var buf bytes.Buffer
	buf.WriteString(p.String())
	p = p.next
	for p != l.head {
		buf.WriteString("<-->")
		buf.WriteString(p.String())
		p = p.next
	}
	return buf.String()
}

func (l *linkedList) remove(n *linkedNode) {
	n.prev.next = n.next
	n.next.prev = n.prev
}

func (l *linkedList) addToHead(n *linkedNode) {
	n.next = l.head.next
	n.prev = l.head
	l.head.next.prev = n
	l.head.next = n
}

func (l *linkedList) moveToHead(n *linkedNode) {
	l.remove(n)
	l.addToHead(n)
}

func (l *linkedList) insert(key string, value Value) *linkedNode {
	new := newLinkedNode(key, value)
	l.addToHead(new)
	return new
}
