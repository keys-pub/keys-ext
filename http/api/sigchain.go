package api

import "github.com/keys-pub/keys"

// SigchainResponse is the response format for a Sigchain request.
type SigchainResponse struct {
	KID        keys.ID             `json:"kid"`
	Metadata   map[string]Metadata `json:"md,omitempty"`
	Statements []*keys.Statement   `json:"statements"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URL()]
	if !ok {
		return Metadata{}
	}
	return md
}

// Sigchain from response.
func (r SigchainResponse) Sigchain() (*keys.Sigchain, error) {
	spk, err := keys.StatementPublicKeyFromID(r.KID)
	if err != nil {
		return nil, err
	}
	sc := keys.NewSigchain(spk)
	for _, st := range r.Statements {
		// md := r.MetadataFor(st)
		// if md.CreatedAt.IsZero() {
		// 	return nil, errors.Errorf("missing metadata for statement in response")
		// }
		if err := sc.Add(st); err != nil {
			return nil, err
		}
	}
	return sc, nil
}

// SigchainsResponse is the response format for a listing all sigchain
// statements.
type SigchainsResponse struct {
	Metadata   map[string]Metadata `json:"md,omitempty"`
	Statements []*keys.Statement   `json:"statements"`
	Version    string              `json:"version"`
}

// MetadataFor returns metadata for Signed.
func (r SigchainsResponse) MetadataFor(st *keys.Statement) Metadata {
	md, ok := r.Metadata[st.URL()]
	if !ok {
		return Metadata{}
	}
	return md
}
