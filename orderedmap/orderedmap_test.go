package orderedmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderedMap(t *testing.T) {
	cases := []struct {
		desc  string
		key   string
		value string
	}{
		{
			desc:  "1",
			key:   "Rohan",
			value: "Cavalry",
		},
		{
			desc:  "2",
			key:   "The shire",
			value: "Hobbits",
		},
		{
			desc:  "3",
			key:   "Mordor",
			value: "Orcs",
		},
	}

	m := New()

	// Test Get
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			m.Set(tc.key, tc.value)

			expected := tc.value
			got := m.Get(tc.key)
			assert.Equal(t, expected, got)
		})
	}

	// Test Keys
	for i, key := range m.Keys() {
		var expected string
		got := m.Get(key)

		switch i {
		case 0:
			expected = "Cavalry"
		case 1:
			expected = "Hobbits"
		case 2:
			expected = "Orcs"
		}

		assert.Equal(t, expected, got)
	}
}
