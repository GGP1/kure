package wallet

import (
	"reflect"
	"testing"
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
		t.Errorf("New wallet failed \nexpected: %v \ngot: %v", expected, got)
	}
}
