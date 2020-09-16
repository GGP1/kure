package db

import (
	"fmt"
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/wallet"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateWallet creates a new cryptocurrency wallet.
func CreateWallet(wallet *wallet.Wallet) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)

		exists := b.Get([]byte(wallet.Name))
		if exists != nil {
			return errors.New("already exists a wallet with this name")
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

// DeleteWallet removes a wallet from the database.
func DeleteWallet(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		n := strings.ToLower(name)

		if err := b.Delete([]byte(n)); err != nil {
			return errors.Wrap(err, "delete wallet")
		}

		return nil
	})
}

// GetWallet retrieves a wallet with the name specified,.
func GetWallet(name string) (*wallet.Wallet, error) {
	c := &wallet.Wallet{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		n := strings.ToLower(name)

		result := b.Get([]byte(n))

		decWallet, err := crypt.Decrypt(result)
		if err != nil {
			return errors.Wrap(err, "decrypt wallet")
		}

		if err := proto.Unmarshal(decWallet, c); err != nil {
			return errors.Wrap(err, "unmarshal wallet")
		}

		if c.Name == "" {
			return fmt.Errorf("%s does not exist", name)
		}

		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "view wallet")
	}

	return c, nil
}

// ListWallets returns a list with all the wallets.
func ListWallets() ([]*wallet.Wallet, error) {
	var wallets []*wallet.Wallet

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wallet := &wallet.Wallet{}

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
		return nil, errors.Wrap(err, "list wallets")
	}

	return wallets, nil
}
