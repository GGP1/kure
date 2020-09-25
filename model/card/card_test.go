package card

import (
	"reflect"
	"testing"
)

func TestNewCard(t *testing.T) {
	expected := &Card{
		Name:       "name",
		Type:       "type",
		Number:     12345,
		CVC:        1234,
		ExpireDate: "never",
	}

	got := New("name", "type", "never", 12345, 1234)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("new card failed \nexpected: %v \ngot: %v", expected, got)
	}
}
