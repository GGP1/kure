package auth

import (
	"crypto/rand"
	"encoding/binary"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/bucket"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

var (
	// authKey is the key we are trying to decrypt on every Login
	authKey = []byte("key")
	// keyfileKey will exist only if the user uses a keyfile
	keyfileKey = []byte("keyfile")
	iterKey    = []byte("iterations")
	memKey     = []byte("memory")
	thKey      = []byte("threads")
)

// Params contains all the information needed for logging in.
type Params struct {
	AuthKey    []byte
	Argon2     Argon2
	UseKeyfile bool
}

// Argon2 execution parameters.
type Argon2 struct {
	Iterations uint32
	Memory     uint32
	Threads    uint32
}

// GetParams returns the authentication parameters.
func GetParams(db *bolt.DB) (Params, error) {
	tx, err := db.Begin(false)
	if err != nil {
		return Params{}, err
	}
	defer tx.Rollback()

	b := tx.Bucket(bucket.Auth.GetName())
	if b == nil {
		return Params{}, nil
	}

	params := make(map[string][]byte, 5)
	_ = b.ForEach(func(k, v []byte) error {
		params[string(k)] = v
		return nil
	})
	_, useKeyfile := params[string(keyfileKey)]

	return Params{
		AuthKey: params[string(authKey)],
		Argon2: Argon2{
			Iterations: binary.BigEndian.Uint32(params[string(iterKey)]),
			Memory:     binary.BigEndian.Uint32(params[string(memKey)]),
			Threads:    binary.BigEndian.Uint32(params[string(thKey)]),
		},
		UseKeyfile: useKeyfile,
	}, nil
}

// Register creates all the buckets, saves the authentication key and the argon2 parameters used.
func Register(db *bolt.DB, params Params) error {
	return db.Update(func(tx *bolt.Tx) error {
		// Create all the buckets except auth, it will be created in setParameters()
		buckets := bucket.GetNames()
		for _, bucket := range buckets {
			if _, err := tx.CreateBucketIfNotExists([]byte(bucket)); err != nil {
				return errors.Wrapf(err, "creating %q bucket", bucket)
			}
		}

		return storeParams(tx, params)
	})
}

// storeParams creates the auth bucket and sets the authentication parameters.
//
// The transaction shouldn't be closed as it's already handled by Register().
func storeParams(tx *bolt.Tx, params Params) error {
	b, err := tx.CreateBucketIfNotExists(bucket.Auth.GetName())
	if err != nil {
		return errors.Wrap(err, "creating auth bucket")
	}

	if err := storeArgon2Params(b, params); err != nil {
		return err
	}

	if err := storeKeyfileFlag(b, params); err != nil {
		return err
	}

	return storeAuthKey(b, params)
}

func storeArgon2Params(b *bolt.Bucket, params Params) error {
	i := make([]byte, 4)
	m := make([]byte, 4)
	t := make([]byte, 4)
	binary.BigEndian.PutUint32(i, params.Argon2.Iterations)
	binary.BigEndian.PutUint32(m, params.Argon2.Memory)
	binary.BigEndian.PutUint32(t, params.Argon2.Threads)

	if err := b.Put(iterKey, i); err != nil {
		return errors.Wrap(err, "saving iterations")
	}

	if err := b.Put(memKey, m); err != nil {
		return errors.Wrap(err, "saving memory")
	}

	if err := b.Put(thKey, t); err != nil {
		return errors.Wrap(err, "saving threads")
	}

	return nil
}

func storeKeyfileFlag(b *bolt.Bucket, params Params) error {
	if params.UseKeyfile {
		if err := b.Put(keyfileKey, []byte("1")); err != nil {
			return errors.Wrap(err, "saving key file value")
		}
		return nil
	}

	// Does not fail if the key doesn't exist
	if err := b.Delete(keyfileKey); err != nil {
		return errors.Wrap(err, "deleting key file value")
	}
	return nil
}

func storeAuthKey(b *bolt.Bucket, params Params) error {
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
