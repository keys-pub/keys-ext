package vault

import httpclient "github.com/keys-pub/keys-ext/http/client"

// ConvertID for testing.
var ConvertID = convertID

func (v *Vault) CheckNonce(n string) error {
	return v.checkNonce(n)
}

func (v *Vault) CommitNonces(ns []string) error {
	return v.commitNonces(ns)
}

func (v *Vault) CheckEventNonces(events []*httpclient.Event) ([]string, error) {
	return v.checkEventNonces(events)
}
