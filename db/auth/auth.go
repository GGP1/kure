package auth

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/GGP1/kure/crypt"
	dbutil "github.com/GGP1/kure/db"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	authBucket = []byte("kure_auth")
	// authKey is the key we are trying to decrypt on every Login
	authKey = []byte("key")
	// keyfileKey will exist only if the user uses a keyfile
	keyfileKey = []byte("keyfile")
	iterKey    = []byte("iterations")
	memKey     = []byte("memory")
	thKey      = []byte("threads")
)

// Parameters contains all the information needed for logging in.
type Parameters struct {
	AuthKey    []byte
	Iterations uint32
	Memory     uint32
	Threads    uint32
	UseKeyfile bool
}

// GetParameters returns the authentication parameters.
func GetParameters(db *bolt.DB) (Parameters, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return Parameters{}, err
	}
	defer tx.Rollback()

	b := tx.Bucket(authBucket)
	if b == nil {
		return Parameters{}, nil
	}

	params := make(map[string][]byte, 5)
	_ = b.ForEach(func(k, v []byte) error {
		params[string(k)] = v
		return nil
	})

	// Key file will be used only if it isn't nil
	useKeyfile := false
	if _, ok := params[string(keyfileKey)]; ok {
		useKeyfile = true
	}

	return Parameters{
		AuthKey:    params[string(authKey)],
		Iterations: binary.BigEndian.Uint32(params[string(iterKey)]),
		Memory:     binary.BigEndian.Uint32(params[string(memKey)]),
		Threads:    binary.BigEndian.Uint32(params[string(thKey)]),
		UseKeyfile: useKeyfile,
	}, nil
}

// Register creates all the buckets, saves the authentication key and the argon2 parameters used.
func Register(db *bolt.DB, params Parameters) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Create all the buckets except auth, it will be created in setParameters()
		buckets := [][]byte{dbutil.CardBucket, dbutil.EntryBucket, dbutil.FileBucket, dbutil.TOTPBucket}
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return errors.Wrapf(err, "creating %q bucket", bucket)
			}
		}

		return setParameters(tx, params)
	})
}

// setParameters creates the auth bucket and sets parameters.
//
// The transaction shouldn't be closed as it's already handled by Register().
func setParameters(tx *bolt.Tx, params Parameters) error {
	b, err := tx.CreateBucketIfNotExists(authBucket)
	if err != nil {
		return errors.Wrap(err, "creating auth bucket")
	}

	// Argon2
	i := make([]byte, 4)
	m := make([]byte, 4)
	t := make([]byte, 4)
	binary.BigEndian.PutUint32(i, params.Iterations)
	binary.BigEndian.PutUint32(m, params.Memory)
	binary.BigEndian.PutUint32(t, params.Threads)

	if err := b.Put(iterKey, i); err != nil {
		return errors.Wrap(err, "saving iterations")
	}

	if err := b.Put(memKey, m); err != nil {
		return errors.Wrap(err, "saving memory")
	}

	if err := b.Put(thKey, t); err != nil {
		return errors.Wrap(err, "saving threads")
	}

	// Keyfile
	if params.UseKeyfile {
		if err := b.Put(keyfileKey, []byte("1")); err != nil {
			return errors.Wrap(err, "saving key file value")
		}
	} else {
		// Does not fail if the key doesn't exist
		if err := b.Delete(keyfileKey); err != nil {
			return errors.Wrap(err, "deleting key file value")
		}
	}

	// Auth key
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return errors.Wrap(err, "generating key")
	}

	encKey, err := crypt.Encrypt(key)
	if err != nil {
		return err
	}

	if err := b.Put(authKey, encKey); err != nil {
		return errors.Wrap(err, "saving auth key")
	}

	return nil
}
