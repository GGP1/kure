package add

import (
	"bytes"
	"net/url"
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/totp"
)

func TestAdd(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	cases := []struct {
		desc   string
		name   string
		input  string
		digits string
		url    string
	}{
		{
			desc:   "Key with 6 digits",
			name:   "6_digits",
			input:  "IFGEWRKSIFJUMR2R",
			digits: "6",
		},
		{
			desc:   "Key with 7 digits",
			name:   "7_digits",
			input:  "IFGEWRKSIFJUMR2R",
			digits: "7",
		},
		{
			desc:   "Key with 8 digits",
			name:   "8_digits",
			input:  "IFGEWRKSIFJUMR2R",
			digits: "8",
		},
		{
			desc:  "URL",
			name:  "url",
			input: "otpauth://totp/Test?secret=IFGEWRKSIFJUMR2R",
			url:   "true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			f := cmd.Flags()
			f.Set("digits", tc.digits)
			f.Set("url", tc.url)

			if err := cmd.Execute(); err != nil {
				t.Errorf("Failed adding TOTP: %v", err)
			}
		})
	}
}

func TestAddErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	name := "test"
	if err := createTOTP(db, name, "", 0); err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		desc   string
		input  string
		name   string
		digits string
		url    string
	}{
		{
			desc: "Invalid name",
			name: "",
		},
		{
			desc: "Key already exists",
			name: name,
		},
		{
			desc:   "Invalid digits",
			name:   "fail",
			digits: "10",
		},
		{
			desc:   "Invalid key",
			name:   "fail",
			digits: "7",
			input:  "invalid/%",
		},
		{
			desc:  "URL already exists",
			input: "otpauth://totp/Test?secret=IFGEWRKSIFJUMR2R",
			url:   "true",
		},
		{
			desc:  "Invalid url secret",
			input: "otpauth://totp/Testing?secret=not-base32",
			url:   "true",
		},
		{
			desc:  "Invalid url format",
			input: "otpauth://hotp/Tests?secret=IFGEWRKSIFJUMR2R",
			url:   "true",
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			buf := bytes.NewBufferString(tc.input)
			cmd := NewCmd(db, buf)
			cmd.SetArgs([]string{tc.name})

			f := cmd.Flags()
			f.Set("digits", tc.digits)
			f.Set("url", tc.url)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestCreateTOTP(t *testing.T) {
	db := cmdutil.SetContext(t, "../../../db/testdata/database")

	t.Run("Success", func(t *testing.T) {
		name := "test"
		if err := createTOTP(db, name, "secret", 6); err != nil {
			t.Fatalf("Failed creating TOTP: %v", err)
		}

		if _, err := totp.Get(db, name); err != nil {
			t.Errorf("%q TOTP not found: %v", name, err)
		}
	})

	t.Run("Fail", func(t *testing.T) {
		if err := createTOTP(db, "", "", 0); err == nil {
			t.Error("Expected an error and got nil")
		}
	})
}

func TestGetName(t *testing.T) {
	expected := "test"
	path := "/Test:mail@enterprise.com"
	got := getName(path)

	if got != expected {
		t.Errorf("Expected %s, got %s", expected, got)
	}
}

func TestStringDigits(t *testing.T) {
	cases := []struct {
		desc     string
		digits   string
		expected int32
	}{
		{
			desc:     "Empty",
			digits:   "",
			expected: 6,
		},
		{
			desc:     "Six digits",
			digits:   "6",
			expected: 6,
		},
		{
			desc:     "Seven digits",
			digits:   "7",
			expected: 7,
		},
		{
			desc:     "Eight digits",
			digits:   "8",
			expected: 8,
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			got := stringDigits(tc.digits)

			if got != tc.expected {
				t.Errorf("Expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		u := &url.URL{
			Scheme: "otpauth",
			Host:   "totp",
		}
		query := u.Query()
		query.Add("algorithm", "SHA1")
		query.Add("period", "30")
		if err := validateURL(u, query); err != nil {
			t.Errorf("Expected a valid URL but got an error: %v", err)
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		cases := []struct {
			desc      string
			scheme    string
			host      string
			period    string
			algorithm string
		}{
			{
				desc:   "Invalid scheme",
				scheme: "otpauth-migration",
			},
			{
				desc:   "Invalid host",
				scheme: "otpauth",
				host:   "hotp",
			},
			{
				desc:      "Invalid algorithm",
				scheme:    "otpauth",
				host:      "totp",
				algorithm: "SHA256",
			},
			{
				desc:   "Invalid period",
				scheme: "otpauth",
				host:   "totp",
				period: "60",
			},
		}

		for _, tc := range cases {
			t.Run(tc.desc, func(t *testing.T) {
				u := &url.URL{
					Scheme: tc.scheme,
					Host:   tc.host,
				}
				query := u.Query()
				query.Add("algorithm", tc.algorithm)
				query.Add("period", tc.period)
				if err := validateURL(u, query); err == nil {
					t.Error("Expected an error and got nil")
				}
			})
		}
	})
}

func TestPostRun(t *testing.T) {
	NewCmd(nil, nil).PostRun(nil, nil)
}
