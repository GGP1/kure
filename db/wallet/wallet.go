package wallet

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
	"google.golang.org/protobuf/proto"
)

var (
	walletBucket     = []byte("kure_wallet")
	errInvalidBucket = errors.New("invalid bucket")
)

// Create creates a new cryptocurrency wallet.
func Create(db *bolt.DB, wallet *pb.Wallet) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		if b == nil {
			return errInvalidBucket
		}

		buf, err := proto.Marshal(wallet)
		if err != nil {
			return errors.Wrap(err, "marshal wallet")
		}

		encWallet, err := crypt.Encrypt(buf)
		if err != nil {
			return errors.Wrap(err, "encrypt wallet")
		}

		if err := b.Put([]byte(wallet.Name), encWallet); err != nil {
			return errors.Wrap(err, "save wallet")
		}

		return nil
	})
}

// Get retrieves the wallet with the specified name.
func Get(db *bolt.DB, name string) (*pb.Wallet, error) {
	wallet := &pb.Wallet{}

	err := db.View(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))
		b := tx.Bucket(walletBucket)
		if b == nil {
			return errInvalidBucket
		}

		encWallet := b.Get([]byte(name))
		if encWallet == nil {
			return errors.Errorf("%q does not exist", name)
		}

		decWallet, err := crypt.Decrypt(encWallet)
		if err != nil {
			return errors.Wrap(err, "decrypt wallet")
		}

		if err := proto.Unmarshal(decWallet, wallet); err != nil {
			return errors.Wrap(err, "unmarshal wallet")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// List returns a list with all the wallets.
func List(db *bolt.DB) ([]*pb.Wallet, error) {
	var wallets []*pb.Wallet

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wallet := &pb.Wallet{}

			decWallet, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt wallet")
			}

			if err := proto.Unmarshal(decWallet, wallet); err != nil {
				return errors.Wrap(err, "unmarshal wallet")
			}

			wallets = append(wallets, wallet)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

// ListByName filters all the entries and returns those matching with the name passed.
func ListByName(db *bolt.DB, name string) ([]*pb.Wallet, error) {
	var group []*pb.Wallet
	name = strings.TrimSpace(strings.ToLower(name))

	wallets, err := List(db)
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		if strings.Contains(w.Name, name) {
			group = append(group, w)
		}
	}

	return group, nil
}

// ListNames returns a list with all the wallets names.
func ListNames(db *bolt.DB) ([]*pb.WalletList, error) {
	var wallets []*pb.WalletList

	_, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		if b == nil {
			return errInvalidBucket
		}
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wallet := &pb.WalletList{}

			decWallet, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt wallet")
			}

			if err := proto.Unmarshal(decWallet, wallet); err != nil {
				return errors.Wrap(err, "unmarshal wallet")
			}

			wallets = append(wallets, wallet)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return wallets, nil
}

// Remove removes a wallet from the database.
func Remove(db *bolt.DB, name string) error {
	_, err := Get(db, name)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		name = strings.TrimSpace(strings.ToLower(name))
		b := tx.Bucket(walletBucket)

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove wallet")
		}

		return nil
	})
}
