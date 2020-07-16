package vault

// ConvertID for testing.
var ConvertID = convertID

func (v *Vault) CheckNonce(n []byte) error {
	return v.checkNonce(n)
}

func (v *Vault) CommitNonce(n []byte) error {
	return v.commitNonce(n)
}
