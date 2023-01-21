package auth

import (
	"bytes"
	"crypto/sha256"
	"os"
	"runtime"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/config"
	"github.com/GGP1/kure/db/auth"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	bolt "go.etcd.io/bbolt"
)

func TestLogin(t *testing.T) {
	db := cmdutil.SetContext(t)

	// This mock is used to execute Login as PreRunE
	mock := func(db *bolt.DB) *cobra.Command {
		return &cobra.Command{
			Use:     "mock",
			PreRunE: Login(db),
		}
	}

	cmd := mock(db)
	assert.NoError(t, cmd.PreRunE(cmd, nil))
}

func TestAskArgon2Params(t *testing.T) {
	cases := []struct {
		desc            string
		input           string
		expectedIters   uint32
		expectedMem     uint32
		expectedThreads uint32
	}{
		{
			desc:            "Custom values",
			input:           "3\n2500000\n6",
			expectedIters:   3,
			expectedMem:     2500000,
			expectedThreads: 6,
		},
		{
			desc:            "Default values",
			input:           "\n\n\n",
			expectedIters:   1,
			expectedMem:     1048576,
			expectedThreads: uint32(runtime.NumCPU()),
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			argon2, err := askArgon2Params(buf)
			assert.NoError(t, err, "Failed taking argon2 parameters")

			assert.Equal(t, tc.expectedIters, argon2.Iterations)
			assert.Equal(t, tc.expectedMem, argon2.Memory)
			assert.Equal(t, tc.expectedThreads, argon2.Threads)
		})
	}
}

func TestArgon2ParamsErrors(t *testing.T) {
	cases := []struct {
		desc  string
		input string
	}{
		{
			desc:  "iterations",
			input: "A\n",
		},
		{
			desc:  "memory",
			input: "4\nA\n",
		},
		{
			desc:  "threads",
			input: "4\n500000\nA\n",
		},
	}

	for _, tc := range cases {
		t.Run("Invalid"+tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)

			_, err := askArgon2Params(buf)
			assert.Error(t, err)
		})
	}
}

func TestAskKeyfile(t *testing.T) {
	cases := []struct {
		desc            string
		input           string
		expectedCfgPath string
		expected        bool
	}{
		{
			desc:            "Do not use key file",
			expected:        false,
			input:           "n\n",
			expectedCfgPath: "./testdata/test-32.key",
		},
		{
			desc:            "Use key file with custom path",
			expected:        true,
			input:           "y\nn\n",
			expectedCfgPath: "",
		},
		{
			desc:            "Use key file with the config path",
			expected:        true,
			input:           "y\ny\n",
			expectedCfgPath: "./testdata/test-32.key",
		},
	}

	config.Load("testdata/mock_config.yaml")

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			// Set the key file path to the configuration
			config.Set(keyfilePath, "./testdata/test-32.key")
			buf := bytes.NewBufferString(tc.input)

			got, err := askKeyfile(buf)
			assert.NoError(t, err, "Failed requesting key file")

			assert.Equal(t, tc.expected, got)

			cfgPath := config.Get(keyfilePath)
			assert.Equal(t, tc.expectedCfgPath, cfgPath)
		})
	}
}

func TestCombineKeys(t *testing.T) {
	cases := []struct {
		desc string
		path string
		hash bool
	}{
		{
			desc: "32 bytes file",
			path: "./testdata/test-32.key",
			hash: false,
		},
		{
			desc: "Other file",
			path: "./testdata/test-default.key",
			hash: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			config.Set(keyfilePath, tc.path)

			enclave, err := combineKeys(nil, memguard.NewEnclave([]byte("test")))
			assert.NoError(t, err, "Failed combining keys")

			pwdBuf, err := enclave.Open()
			assert.NoError(t, err, "Failed opening enclave")
			defer pwdBuf.Destroy()

			key, err := os.ReadFile(tc.path)
			assert.NoError(t, err, "Failed reading key file")

			if tc.hash {
				h := sha256.New()
				h.Write(key)
				key = h.Sum(nil)
			}
			key = append(key, []byte("test")...)

			assert.Equal(t, key, pwdBuf.Bytes())
		})
	}
}

func TestCombineKeysRequestPath(t *testing.T) {
	config.Reset()
	path := "./testdata/test-32.key"
	buf := bytes.NewBufferString(path)

	enclave, err := combineKeys(buf, memguard.NewEnclave([]byte("test")))
	assert.NoError(t, err)

	pwdBuf, err := enclave.Open()
	assert.NoError(t, err, "Failed opening enclave")
	defer pwdBuf.Destroy()

	key, err := os.ReadFile(path)
	assert.NoError(t, err, "Failed reading key file")

	key = append(key, []byte("test")...)
	assert.Equal(t, key, pwdBuf.Bytes())
}

func TestCombineKeysErrors(t *testing.T) {
	config.Set("keyfile.path", "non-existent")

	_, err := combineKeys(nil, memguard.NewEnclave([]byte("test")))
	assert.Error(t, err)

	t.Run("Invalid path", func(t *testing.T) {
		config.Reset()
		_, err := combineKeys(bytes.NewBufferString("\n"), nil)
		assert.Error(t, err)
	})
}

func TestSetAuthToConfig(t *testing.T) {
	defer config.Reset()

	expPassword := memguard.NewEnclave([]byte("test"))
	const (
		expMem  uint32 = 150000
		expIter uint32 = 110
		expTh   uint32 = 4
	)

	authParams := auth.Params{
		Argon2: auth.Argon2{
			Iterations: expIter,
			Memory:     expMem,
			Threads:    expTh,
		},
	}

	setAuthToConfig(expPassword, authParams)

	// reflect.DeepEqual does not work
	got := config.Get("auth").(map[string]interface{})
	gotPassword := got["password"]
	gotMem := got["memory"].(uint32)
	gotIter := got["iterations"].(uint32)
	gotTh := got["threads"].(uint32)

	assert.Equal(t, expPassword, gotPassword)
	assert.Equal(t, expMem, gotMem)
	assert.Equal(t, expIter, gotIter)
	assert.Equal(t, expTh, gotTh)
}
