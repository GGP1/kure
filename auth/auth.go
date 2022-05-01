package auth

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/auth"
	authDB "github.com/GGP1/kure/db/auth"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	bolt "go.etcd.io/bbolt"
)

// Key file path configuration key
const keyfilePath string = "keyfile.path"

// Login verifies that the human/machine that is trying to execute
// a command is effectively the owner of the information.
//
// If it's the first record the user is registered.
func Login(db *bolt.DB) cmdutil.RunEFunc {
	return func(cmd *cobra.Command, args []string) error {
		// If auth is not nil it means the user is already logged in (session)
		if auth := config.Get("auth"); auth != nil {
			return nil
		}

		params, err := authDB.GetParameters(db)
		if err != nil {
			return err
		}
		// The auth key will be nil only on the user's first (successful) command
		if params.AuthKey == nil {
			return Register(db, os.Stdin)
		}

		password, err := AskPassword("Enter master password", false)
		if err != nil {
			return err
		}

		if params.UseKeyfile {
			password, err = combineKeys(os.Stdin, password)
			if err != nil {
				return err
			}
		}

		setAuthToConfig(password, params)

		// Try to decrypt the authentication key
		if _, err := crypt.Decrypt(params.AuthKey); err != nil {
			return errors.New("invalid master password")
		}

		return nil
	}
}

// Register registers the user when there aren't any records yet.
func Register(db *bolt.DB, r io.Reader) error {
	password, err := AskPassword("New master password", true)
	if err != nil {
		return err
	}

	iterations, memory, threads, err := askArgon2Params(r)
	if err != nil {
		return err
	}

	useKeyfile, err := askKeyfile(r)
	if err != nil {
		return err
	}

	if useKeyfile {
		password, err = combineKeys(r, password)
		if err != nil {
			return err
		}
	}

	params := authDB.Parameters{
		Iterations: iterations,
		Memory:     memory,
		Threads:    threads,
		UseKeyfile: useKeyfile,
	}

	setAuthToConfig(password, params)
	return authDB.Register(db, params)
}

func askArgon2Params(r io.Reader) (iterations, memory, threads uint32, err error) {
	fmt.Println("Set argon2 parameters, leave blank to use the default value")
	fmt.Println("For more information visit https://github.com/GGP1/kure/wiki/Authentication")

	reader := bufio.NewReader(r)

	iterations, err = scanParameter(reader, "Iterations", 1)
	if err != nil {
		return 0, 0, 0, err
	}

	// memory is measured in kibibytes, 1 kibibyte = 1024 bytes. 1048576 kibibytes -> 1GB
	memory, err = scanParameter(reader, "Memory", 1<<20)
	if err != nil {
		return 0, 0, 0, err
	}

	threads, err = scanParameter(reader, "Threads", uint32(runtime.NumCPU()))
	if err != nil {
		return 0, 0, 0, err
	}

	return iterations, memory, threads, nil
}

// askKeyfile asks the user if he wants to use a key file or not.
func askKeyfile(r io.Reader) (bool, error) {
	if !cmdutil.Confirm(r, "Would you like to use a key file?") {
		return false, nil
	}

	if config.GetString(keyfilePath) != "" {
		if !cmdutil.Confirm(r, "Would you like to use the path specified in the configuration file?") {
			config.Set(keyfilePath, "")
		}
	}

	return true, nil
}

func combineKeys(r io.Reader, password *memguard.Enclave) (*memguard.Enclave, error) {
	path := config.GetString(keyfilePath)
	if path == "" {
		path = cmdutil.Scanln(bufio.NewReader(r), "Enter key file path")
		path = strings.Trim(path, "\"")
		if path == "" || path == "." {
			return nil, errors.New("invalid key file path")
		}
	}

	key, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading key file")
	}
	defer memguard.WipeBytes(key)

	pwdBuf, err := password.Open()
	if err != nil {
		return nil, errors.Wrap(err, "decrypting password")
	}

	// If the content is not 32 bytes, hash it and use the hash as the key
	if len(key) != 32 {
		keyHash := sha256.Sum256(key)
		key = keyHash[:]
	}

	key = append(key, pwdBuf.Bytes()...)
	pwdBuf.Destroy()

	return memguard.NewEnclave(key), nil
}

func scanParameter(r *bufio.Reader, field string, defaultValue uint32) (uint32, error) {
	valueStr := cmdutil.Scanln(r, " "+field)
	if valueStr == "" {
		return defaultValue, nil
	}

	v, err := strconv.Atoi(valueStr)
	if err != nil || v < 1 {
		return 0, errors.Wrapf(err, "invalid %s number", strings.ToLower(field))
	}

	return uint32(v), nil
}

// Auth values must be set to the configuration before any encryption/decryption occurs.
// Probable not the best way of handling the parameters but it's flexible.
func setAuthToConfig(password *memguard.Enclave, params auth.Parameters) {
	auth := map[string]interface{}{
		"password":   password,
		"iterations": params.Iterations,
		"memory":     params.Memory,
		"threads":    params.Threads,
	}
	config.Set("auth", auth)
}
