package wallet

import (
	"reflect"
	"testing"
)

// Checkboxes
const (
	succeed = "\u2713"
	failed  = "\u2717"
)

func TestNewWallet(t *testing.T) {
	expected := &Wallet{
		Name:         "name",
		Type:         "type",
		ScriptType:   "script type",
		KeystoreType: "keystore type",
		SeedPhrase:   "seed phrase",
		PublicKey:    "public key",
		PrivateKey:   "private key",
	}

	got := New("name", "type", "script type", "keystore type", "seed phrase", "public key", "private key")

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("new wallet failed \nexpected: %v \ngot: %v", expected, got)
	}
}
