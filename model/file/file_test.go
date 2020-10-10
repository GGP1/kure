package file

import (
	"reflect"
	"testing"
)

func TestNewFile(t *testing.T) {
	expected := &File{
		Name:      "name",
		Content:   []byte("file"),
		Type:      "png",
		CreatedAt: 0,
	}

	got := New("name", "png", []byte("file"), 0)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("New wallet failed \nexpected: %v \ngot: %v", expected, got)
	}
}
