package orderedmap

import "testing"

func Test(t *testing.T) {
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

	m := New(3)
	defer m.Destroy()

	// Test Get
	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			m.Set(tc.key, tc.value)

			expected := tc.value
			got := m.Get(tc.key)

			if got != expected {
				t.Errorf("Expected %v, got %v", expected, got)
			}
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

		if got != expected {
			t.Errorf("Expected %v, got %v", expected, got)
		}
	}
}
