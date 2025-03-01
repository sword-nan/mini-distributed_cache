package cache

import "bytes"

type Value interface {
	NBytes() int
	Bytes() []byte
}

// NewValue Interface
// used to create Value instance
type NewValue interface {
	New(b []byte) Value
}

// functional interface
type NewValueFunc func(b []byte) Value

func (f NewValueFunc) New(b []byte) Value {
	return f(b)
}

// default cache Value structure
type ByteView struct {
	b []byte
}

func (v ByteView) NBytes() int {
	return len(v.b)
}

func (v ByteView) Bytes() []byte {
	var bytes = make([]byte, len(v.b))
	copy(bytes, v.b)
	return bytes
}

func (v ByteView) String() string {
	var buffer bytes.Buffer
	buffer.Write(v.b)
	return buffer.String()
}

func (v ByteView) New(b []byte) Value {
	return NewByteView(b)
}

func NewByteView(b []byte) ByteView {
	return ByteView{b: b}
}
