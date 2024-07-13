package restore

import (
	"fmt"
	"os"

	"github.com/GGP1/kure/auth"
	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/bucket"
	"github.com/GGP1/kure/sig"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

// NewCmd returns a new command.
func NewCmd(db *bolt.DB) *cobra.Command {
	return &cobra.Command{
		Use:   "restore",
		Short: "Restore the database using new credentials",
		Long: `Restore the database using new credentials.

Overwrite the registered credentials and re-encrypt every record with the new ones.

WARNING: this command is computationally expensive, it may cause memory (OOM) and CPU errors.`,
		PreRunE: auth.Login(db),
		RunE:    runRestore(db),
	}
}

func runRestore(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		buckets := bucket.GetNames()
		logs := make([]*log, 0, len(buckets))

		for _, bucket := range buckets {
			log, err := newLog(bucket)
			if err != nil {
				return err
			}
			defer log.Close()
			sig.Signal.AddCleanup(func() error { return log.Close() })
			logs = append(logs, log)
		}

		if err := writeLogs(db, logs); err != nil {
			return errors.Wrap(err, "writing logs")
		}

		// Initialize registration and re-encrypt the records with the new credentials
		if err := auth.Register(db, os.Stdin); err != nil {
			return err
		}
		fmt.Println("Re-encrypting records...")

		if err := readLogs(db, logs); err != nil {
			return errors.Wrap(err, "recreating records")
		}

		return nil
	}
}

func readLogs(db *bolt.DB, logs []*log) error {
	tx, err := db.Begin(true)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	for _, log := range logs {
		b := tx.Bucket(log.BucketName())
		data, err := log.Read()
		if err != nil {
			return err
		}

		for i := 0; i < len(data)-1; i += 2 {
			key := data[i]
			value := data[i+1]
			encValue, err := crypt.Encrypt(value)
			if err != nil {
				return errors.Wrap(err, "encrypt record value")
			}
			if err := b.Put(key, encValue); err != nil {
				return errors.Wrap(err, "saving record")
			}
		}
	}

	return tx.Commit()
}

func writeLogs(db *bolt.DB, logs []*log) error {
	tx, err := db.Begin(false)
	if err != nil {
		return errors.Wrap(err, "starting transaction")
	}
	defer tx.Rollback()

	for _, l := range logs {
		b := tx.Bucket(l.BucketName())

		err = b.ForEach(func(k, v []byte) error {
			decValue, err := crypt.Decrypt(v)
			if err != nil {
				return errors.Wrap(err, "decrypt record value")
			}

			if err := l.Write(k); err != nil {
				return err
			}
			return l.Write(decValue)
		})
		if err != nil {
			return err
		}

		// Flush data to stable storage
		if err := l.Sync(); err != nil {
			return err
		}
	}

	return nil
}
