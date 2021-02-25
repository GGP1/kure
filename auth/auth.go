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
	"github.com/GGP1/kure/crypt"
	"github.com/GGP1/kure/db/auth"
	authDB "github.com/GGP1/kure/db/auth"

	"github.com/awnumar/memguard"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		if auth := viper.Get("auth"); auth != nil {
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
	// Default values
	iterations = 1
	// memory is measured in kibibytes, 1 kibibyte = 1024 bytes.
	memory = 1 << 20 // 1048576 kibibytes -> 1GB
	threads = uint32(runtime.NumCPU())

	fmt.Println("Set argon2 parameters, leave blank to use the default value")
	fmt.Println("For more information visit https://github.com/GGP1/kure/wiki/Authentication")
	scanner := bufio.NewScanner(r)

	if iter := cmdutil.Scanln(scanner, " Iterations"); iter != "" {
		i, err := strconv.Atoi(iter)
		if err != nil || i < 1 {
			return 0, 0, 0, errors.New("invalid iterations number")
		}
		iterations = uint32(i)
	}

	if mem := cmdutil.Scanln(scanner, " Memory"); mem != "" {
		m, err := strconv.Atoi(mem)
		if err != nil || m < 1 {
			return 0, 0, 0, errors.New("invalid memory number")
		}
		memory = uint32(m)
	}

	if th := cmdutil.Scanln(scanner, " Threads"); th != "" {
		t, err := strconv.Atoi(th)
		if err != nil || t < 1 {
			return 0, 0, 0, errors.New("invalid threads number")
		}
		threads = uint32(t)
	}

	return iterations, memory, threads, nil
}

// askKeyfile asks the user if he wants to use a key file or not.
func askKeyfile(r io.Reader) (bool, error) {
	use := false

	if cmdutil.Confirm(r, "Would you like to use a key file?") {
		use = true

		if viper.GetString(keyfilePath) != "" {
			if !cmdutil.Confirm(r,
				"Would you like to use the path specified in the configuration file?") {

				// Overwrite the path value in the configuration file
				viper.Set(keyfilePath, "")
				if err := viper.WriteConfigAs(viper.ConfigFileUsed()); err != nil {
					return false, errors.Wrap(err, "writing the configuration file")
				}
			}
		}
	}

	return use, nil
}

func combineKeys(r io.Reader, password *memguard.Enclave) (*memguard.Enclave, error) {
	path := viper.GetString(keyfilePath)
	if path == "" {
		fmt.Print("Enter key file path: ")
		fmt.Fscanln(r, &path)

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

	// If the content is not 32 bytes, hash it an use the hash as the key
	if len(key) != 32 {
		h := sha256.New()
		h.Write(key) // Never fails
		key = h.Sum(nil)
	}

	key = append(key, pwdBuf.Bytes()...)
	pwdBuf.Destroy()

	return memguard.NewEnclave(key), nil
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
	viper.Set("auth", auth)
}
