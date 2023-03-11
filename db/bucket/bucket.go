// Package bucket contains the database buckets where records are stored.
//
// They are in a separate bucket and inside an unexported struct so they can't be
// modified by other modules.
package bucket

// Database bucket
var (
	Auth       = bucket{[]byte("kure_auth")}
	Card       = bucket{[]byte("kure_card")}
	CardNames  = bucket{[]byte("card_names")}
	Entry      = bucket{[]byte("kure_entry")}
	EntryNames = bucket{[]byte("entry_names")}
	File       = bucket{[]byte("kure_file")}
	FileNames  = bucket{[]byte("file_names")}
	TOTP       = bucket{[]byte("kure_totp")}
	TOTPNames  = bucket{[]byte("totp_names")}
	names      = [][]byte{
		Card.GetName(), CardNames.GetName(),
		Entry.GetName(), EntryNames.GetName(),
		File.GetName(), FileNames.GetName(),
		TOTP.GetName(), TOTPNames.GetName(),
	}
)

type bucket struct {
	name []byte
}

func (b *bucket) GetName() []byte {
	return b.name
}

// GetNames returns a slice with the names of the buckets where records are stored.
// The auth bucket is not included.
func GetNames() [][]byte {
	return names
}
