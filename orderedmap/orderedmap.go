// Package orderedmap offers an ordered group of key/value pairs stored in a secure way.
package orderedmap

import (
	"container/list"
	"unsafe"

	"github.com/awnumar/memguard"
)

// Map is a combination of a map and a double linked list.
type Map struct {
	mp         map[string]*list.Element
	linkedList *list.List

	// List of each pair's locked buffer
	buffers []*memguard.LockedBuffer

	// Count is used to store locked buffers inside "buffers" fixed slice
	count uint8
}

type pair struct {
	key   string
	value string
}

// New initializes map's fields with the size provided and returns it.
func New(size int) *Map {
	return &Map{
		mp:         make(map[string]*list.Element, size),
		linkedList: list.New(),
		buffers:    make([]*memguard.LockedBuffer, size),
		count:      0,
	}
}

// Destroy destroys all pairs' underlying buffers.
func (m *Map) Destroy() {
	for _, b := range m.buffers {
		b.Destroy()
	}
}

// Get returns key's correspondent value.
func (m *Map) Get(key string) string {
	return m.mp[key].Value.(*pair).value
}

// Keys returns a slice with all the keys of the map.
func (m *Map) Keys() []string {
	keys := make([]string, len(m.mp))
	item := m.linkedList.Front()

	for i := 0; item != nil; i++ {
		keys[i] = item.Value.(*pair).key
		item = item.Next()
	}

	return keys
}

// Set inserts a new pair at the end of the map.
func (m *Map) Set(key, value string) {
	lockedBuf, p := securePair()
	p.key = key
	p.value = value
	memguard.WipeBytes([]byte(value))

	m.mp[key] = m.linkedList.PushBack(p)
	m.buffers[m.count] = lockedBuf
	m.count++
}

// securePair returns a pair along with the locked buffer where it is allocated.
func securePair() (*memguard.LockedBuffer, *pair) {
	s := new(pair)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*pair)(unsafe.Pointer(&b.Bytes()[0]))
}
