package db

import (
	"strings"

	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/pb"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	bolt "go.etcd.io/bbolt"
)

// CreateWallet creates a new cryptocurrency wallet.
func CreateWallet(wallet *pb.Wallet) error {
	wallets, err := WalletsByName(wallet.Name)
	if err != nil {
		return err
	}

	var exists bool
	for _, w := range wallets {
		if strings.Split(w.Name, "/")[0] == wallet.Name {
			exists = true
			break
		}
	}

	if exists {
		return errors.New("already exists a wallet or folder with this name")
	}

	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)

		exists := b.Get([]byte(wallet.Name))
		if exists != nil {
			return errors.Errorf("already exists a wallet named %s", wallet.Name)
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

// GetWallet retrieves the wallet with the specified name.
func GetWallet(name string) (*pb.Wallet, error) {
	wallet := &pb.Wallet{}

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		encWallet := b.Get([]byte(name))
		if encWallet == nil {
			return errors.Errorf("\"%s\" does not exist", name)
		}

		decWallet, err := crypt.Decrypt(encWallet)
		if err != nil {
			return errors.Wrap(err, "decrypt wallet")
		}

		if err := proto.Unmarshal(decWallet, wallet); err != nil {
			return errors.Wrap(err, "unmarshal wallet")
		}

		if wallet.Name == "" {
			return errors.Errorf("%s does not exist", name)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// ListWallets returns a list with all the wallets.
func ListWallets() ([]*pb.Wallet, error) {
	var wallets []*pb.Wallet

	password, err := crypt.GetMasterPassword()
	if err != nil {
		return nil, err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		c := b.Cursor()

		// Place cursor in the first line of the bucket and move it to the next one
		for k, v := c.First(); k != nil; k, v = c.Next() {
			wallet := &pb.Wallet{}

			decWallet, err := crypt.DecryptX(v, password)
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

// RemoveWallet removes a wallet from the database.
func RemoveWallet(name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(walletBucket)
		name = strings.TrimSpace(strings.ToLower(name))

		_, err := GetWallet(name)
		if err != nil {
			return err
		}

		if err := b.Delete([]byte(name)); err != nil {
			return errors.Wrap(err, "remove wallet")
		}

		return nil
	})
}

// WalletsByName filters all the entries and returns those matching with the name passed.
func WalletsByName(name string) ([]*pb.Wallet, error) {
	var group []*pb.Wallet
	name = strings.TrimSpace(strings.ToLower(name))

	wallets, err := ListWallets()
	if err != nil {
		return nil, err
	}

	for _, w := range wallets {
		if strings.Contains(w.Name, name) {
			group = append(group, w)
		}
	}

	if len(group) == 0 {
		return nil, errors.New("no wallets were found")
	}

	return group, nil
}
