package pb

import (
	"strings"
	"unsafe"

	"github.com/awnumar/memguard"
)

// SecureCard returns a Card along with the locked buffer where it is allocated.
func SecureCard() (*memguard.LockedBuffer, *Card) {
	s := new(Card)
	// Allocate a LockedBuffer of the correct size
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	// Return the LockedBuffer along with the initialised struct
	return b, (*Card)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureCardSlice returns a locked buffer and the slice of Cards allocated in it.
func SecureCardSlice() (*memguard.LockedBuffer, []*Card) {
	// Initialise an instance of the struct type
	s := new(Card)
	// Allocate the enough memory to store the struct values
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	// Construct the slice from its parameters
	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Bytes()[0])), 0, 0}

	// Return the LockedBuffer along with the constructed slice
	return b, *(*[]*Card)(unsafe.Pointer(&sl))
}

// SecureEntry returns an Entry along with the locked buffer where it is allocated.
func SecureEntry() (*memguard.LockedBuffer, *Entry) {
	s := new(Entry)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*Entry)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureEntrySlice returns a locked buffer and the slice of Entries allocated in it.
func SecureEntrySlice() (*memguard.LockedBuffer, []*Entry) {
	s := new(Entry)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Bytes()[0])), 0, 0}

	return b, *(*[]*Entry)(unsafe.Pointer(&sl))
}

// SecureEntryList returns an EntryList along with the locked buffer where it is allocated.
func SecureEntryList() (*memguard.LockedBuffer, *EntryList) {
	s := new(EntryList)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*EntryList)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureFile returns a File along with the locked buffer where it is allocated.
func SecureFile() (*memguard.LockedBuffer, *File) {
	s := new(File)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*File)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureFileCheap returns a File along with the locked buffer where it is allocated.
func SecureFileCheap() (*memguard.LockedBuffer, *FileCheap) {
	s := new(FileCheap)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*FileCheap)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureFileSlice returns a locked buffer and the slice of Files allocated in it.
func SecureFileSlice() (*memguard.LockedBuffer, []*File) {
	s := new(File)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Bytes()[0])), 0, 0}

	return b, *(*[]*File)(unsafe.Pointer(&sl))
}

// SecureNote returns a Note along with the locked buffer where it is allocated.
func SecureNote() (*memguard.LockedBuffer, *Note) {
	s := new(Note)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*Note)(unsafe.Pointer(&b.Bytes()[0]))
}

// SecureNoteSlice returns a locked buffer and the slice of Notes allocated in it.
func SecureNoteSlice() (*memguard.LockedBuffer, []*Note) {
	s := new(Note)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{uintptr(unsafe.Pointer(&b.Bytes()[0])), 0, 0}

	return b, *(*[]*Note)(unsafe.Pointer(&sl))
}

// SecureStringBuilder returns a strings.Builder along with the locked buffer where it is allocated.
func SecureStringBuilder() (*memguard.LockedBuffer, *strings.Builder) {
	s := new(strings.Builder)
	b := memguard.NewBuffer(int(unsafe.Sizeof(*s)))

	return b, (*strings.Builder)(unsafe.Pointer(&b.Bytes()[0]))
}
