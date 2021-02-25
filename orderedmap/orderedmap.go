// Package orderedmap offers an ordered group of key/value pairs.
package orderedmap

import "container/list"

// Map is a combination of a map and a double linked list.
type Map struct {
	mp         map[string]*list.Element
	linkedList *list.List
}

type pair struct {
	key   string
	value string
}

// New initializes an ordered map and returns it.
func New() *Map {
	return &Map{
		mp:         make(map[string]*list.Element),
		linkedList: list.New(),
	}
}

// Get returns key's value.
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
	p := &pair{
		key:   key,
		value: value,
	}

	m.mp[key] = m.linkedList.PushBack(p)
}
