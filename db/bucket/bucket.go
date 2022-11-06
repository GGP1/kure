// Package bucket contains the database buckets where records are stored.
//
// They are in a separate bucket and inside an unexported struct so they can't be
// modified by other modules.
package bucket

// Database bucket
var (
	Auth  = bucket{[]byte("kure_auth")}
	Card  = bucket{[]byte("kure_card")}
	Entry = bucket{[]byte("kure_entry")}
	File  = bucket{[]byte("kure_file")}
	TOTP  = bucket{[]byte("kure_totp")}
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
	return [][]byte{
		Card.GetName(),
		Entry.GetName(),
		File.GetName(),
		TOTP.GetName(),
	}
}
