package ls

import (
	"testing"

	cmdutil "github.com/GGP1/kure/commands"
	"github.com/GGP1/kure/db/entry"
	"github.com/GGP1/kure/pb"

	bolt "go.etcd.io/bbolt"
)

func TestLs(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db, "test", "testing")

	cases := []struct {
		desc   string
		name   string
		filter string
		hide   string
		qr     string
	}{
		{
			desc: "List one",
			name: "test",
		},
		{
			desc: "List one and show qr",
			name: "test",
			qr:   "true",
		},
		{
			desc:   "Filter by name",
			name:   "te*",
			filter: "true",
		},
		{
			desc: "List all",
			name: "",
		},
		{
			desc: "List one and hide",
			name: "test",
			hide: "true",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("filter", tc.filter)
			f.Set("hide", tc.hide)
			f.Set("qr", tc.qr)

			if err := cmd.Execute(); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestLsErrors(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db, "test", "")

	cases := []struct {
		desc   string
		name   string
		filter string
		qr     string
	}{
		{
			desc: "Entry does not exist",
			name: "non-existent",
		},
		{
			desc:   "No entries found",
			name:   "non-existent",
			filter: "true",
		},
		{
			desc:   "Filter syntax error",
			name:   "[error",
			filter: "true",
		},
		{
			desc: "No data to encode",
			name: "test",
			qr:   "true",
		},
	}

	cmd := NewCmd(db)

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			cmd.SetArgs([]string{tc.name})
			f := cmd.Flags()
			f.Set("filter", tc.filter)
			f.Set("qr", tc.qr)

			if err := cmd.Execute(); err == nil {
				t.Error("Expected an error and got nil")
			}
		})
	}
}

func TestQRCodeError(t *testing.T) {
	db := cmdutil.SetContext(t, "../../db/testdata/database")
	createEntry(t, db, "test", longSecret)

	cmd := NewCmd(db)
	cmd.SetArgs([]string{"test"})
	cmd.Flags().Set("qr", "true")

	if err := cmd.Execute(); err == nil {
		t.Error("Expected secret too long to encode to QR and got nil")
	}
}

func TestCheckExpiration(t *testing.T) {
	cases := []struct {
		expiration string
		expected   bool
	}{
		{
			expiration: "Mon, 02 Jan 2020 15:04:05 -0700",
			expected:   true,
		},
		{
			expiration: "Never",
			expected:   false,
		},
		{
			expiration: "Mon, 02 Jan 2040 15:04:05 -0700",
			expected:   false,
		},
	}

	for _, tc := range cases {
		t.Run("Check expiration", func(t *testing.T) {
			got := expired(tc.expiration)
			if tc.expected != got {
				t.Errorf("Expected %v, got %v", tc.expected, got)
			}
		})
	}
}

func TestPostRun(t *testing.T) {
	NewCmd(nil).PostRun(nil, nil)
}

func createEntry(t *testing.T, db *bolt.DB, name, password string) {
	t.Helper()

	e := &pb.Entry{
		Name:     name,
		Password: password,
		Expires:  "Mon, 01 Jan 2021 15:04:05 -0700",
	}
	if err := entry.Create(db, e); err != nil {
		t.Fatal(err)
	}
}

var longSecret string = `GLRSEp1Pn0f8LYr5OBcFDguVdR2qeElxSuculJxn7o3aecr YMmOZ3lyXxviLQGtT0eQTOtMeXvo0LZxSev 8U7zigJAb9CqYPLngsbMR44Uq3WHTiLP3f9kzK8iESNr7 hKCi82xz3CH5EWpFkP4zGAtUqU39XMgxe7A5D1TgM Y6nCyMN7amn4bhEE6fyPlvbiuYD8rQUMSRjL8zuFrgyRK3WKO1EvDNi X5oHmuJpeqJCmxZKSaPAN3axCNAvuCltbdZgByhIEzIOD1g1b8M4QVHhzZrIWPiUUyA26frQQ0Ri3RnMKjbIOlFmcULS69758PeYxbCbA4XVJTBNEf gpkNRaMZkx2mSF02DLM4qZpij6Riu3cHLaSImwUoEjsM3x0aVWFfLUk11 IyX9Aq9gXZMZzOK2b8BBmY KAnvIwIg8yGwZKLB6b WL0PVUNe55IoAZaCISvvZt0qqydFBZ6W8pdMRYjjJ4ZfDf41LDS1y G2FnzcEWfUxK00GIkZUV3Y3uWsMfjkC9Eac1LzXbCL10p1 m1IxSA2jSY7XqTzruyOLCfw0suzDL2yIKir7OEK0gugVeIaaWopCqQ5QmhktOGMDoIf8u0hJkUjc5Ej5Idppc5 3lmzABWQ7Km4xg 6HcNn0yddZgsnFrDSgkRxTVIIUBiXoe5oI6U1 1bPTlzqrgfyoT2EyHvqtH1h9FVlcqx3tmDibzOmgNW rEm6IpI5gD96MlF4ej3MsZg7 wnI 81ekz1Wh6KDGT9aTuBMKUd44vDANZeP9QUD7m9Vvn86LRHLtVj9qby0dCgJVRFjxi2o9MmGQNURyXxcLdEMCC uo1B3Pfaygex SRfWztVUfhp17hONHy2rhSBK5SLza vzuiSYlZXlA96Xc4gXPCKBKvkyHJJo96u8FyXawQPcP4z5TnTp0nJUmzUzXgBPCskIFxJNCteNNcvyV0ra1STr R3pp2qWI4UZIfL32p766p3JklYfPpAn45MtdApjgYUNKWunQ39dGVMym4U9zUpgUDE5jftyVELNvVo7YUy5xgAVYuXvpvDMhzL72VzXKD74p P5IjKwPO5xEp0Civwfs5MH9FQwOc1cPAUnb3d0WeovONapBMXZ 34EIbBvranZTc3ASUUDgVqjzZ8ivuwks8Cgd93F3wIZLSfUMzhrzW BjA3hKynuZWqh4q04Mnq30PSFzevkLF2RTPVYw44xyBG cnYeq3wCdo9ShUab jOVUDoAD1atzImIMK6Q5nFQjFIs2Su9u42tnQrF8BIL7LdDy3WT1`
