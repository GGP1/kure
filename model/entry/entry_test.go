package entry

import (
	"reflect"
	"testing"
)

func TestNewEntry(t *testing.T) {
	expected := &Entry{
		Name:     "Github",
		Username: "GGP1",
		Password: "test",
		URL:      "https://www.github.com",
		Notes:    "Commit tests",
		Expires:  "Never",
	}

	got := New("Github", "GGP1", "test", "https://www.github.com", "Commit tests", "Never")

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Test failed, \nexpected: %v \ngot: %v", expected, got)
	}
}
