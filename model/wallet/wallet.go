package wallet

// New creates a new wallet.
func New(name, wType, scriptType, keystoreType, seedPhrase, publicKey, privateKey string) *Wallet {
	return &Wallet{
		Name:         name,
		Type:         wType,
		ScriptType:   scriptType,
		KeystoreType: keystoreType,
		SeedPhrase:   seedPhrase,
		PublicKey:    publicKey,
		PrivateKey:   privateKey,
	}
}
