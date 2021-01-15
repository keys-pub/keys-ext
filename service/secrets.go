package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault/secrets"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

func (s *service) Secret(ctx context.Context, req *SecretRequest) (*SecretResponse, error) {
	secrets := secrets.New(s.vault)
	secret, err := secrets.Get(req.ID)
	if err != nil {
		return nil, err
	}
	if secret == nil {
		return nil, keys.NewErrNotFound(req.ID)
	}
	return &SecretResponse{
		Secret: secretToRPC(secret),
	}, nil
}

func (s *service) SecretSave(ctx context.Context, req *SecretSaveRequest) (*SecretSaveResponse, error) {
	secret := secretFromRPC(req.Secret)
	if secret.Type == secrets.UnknownType {
		return nil, errors.Errorf("unknown secret type")
	}

	name := strings.TrimSpace(secret.Name)
	if name == "" {
		return nil, errors.Errorf("no name specified")
	}

	secrets := secrets.New(s.vault)
	out, _, err := secrets.Save(secret)
	if err != nil {
		return nil, err
	}

	return &SecretSaveResponse{
		Secret: secretToRPC(out),
	}, nil
}

func (s *service) SecretRemove(ctx context.Context, req *SecretRemoveRequest) (*SecretRemoveResponse, error) {
	if req.ID == "" {
		return nil, errors.Errorf("id not specified")
	}
	// TODO: What if ID isn't for secret?
	ok, err := s.vault.Delete(req.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(req.ID)
	}
	return &SecretRemoveResponse{}, nil
}

func (s *service) Secrets(ctx context.Context, req *SecretsRequest) (*SecretsResponse, error) {
	if req.SortField == "" {
		req.SortField = "name"
	}

	sv := secrets.New(s.vault)
	results, err := sv.List(
		secrets.WithQuery(req.Query),
		secrets.WithSort(req.SortField),
		secrets.WithSortDirection(sortDirectionToVault(req.SortDirection)),
	)
	if err != nil {
		return nil, err
	}
	out := secretsToRPC(results)

	return &SecretsResponse{
		Secrets:       out,
		SortField:     req.SortField,
		SortDirection: req.SortDirection,
	}, nil
}

func sortDirectionToVault(d SortDirection) secrets.SortDirection {
	switch d {
	case SortAsc:
		return secrets.Ascending
	case SortDesc:
		return secrets.Descending
	default:
		return secrets.Ascending
	}
}

func secretsToRPC(ss []*secrets.Secret) []*Secret {
	out := make([]*Secret, 0, len(ss))
	for _, s := range ss {
		out = append(out, secretToRPC(s))
	}
	return out
}

func secretToRPC(s *secrets.Secret) *Secret {
	return &Secret{
		ID:        s.ID,
		Name:      s.Name,
		Type:      secretTypeToRPC(s.Type),
		Username:  s.Username,
		Password:  s.Password,
		URL:       s.URL,
		Notes:     s.Notes,
		CreatedAt: tsutil.Millis(s.CreatedAt),
		UpdatedAt: tsutil.Millis(s.UpdatedAt),
	}
}

func secretFromRPC(s *Secret) *secrets.Secret {
	return &secrets.Secret{
		ID:        s.ID,
		Name:      s.Name,
		Type:      secretTypeFromRPC(s.Type),
		Username:  s.Username,
		Password:  s.Password,
		URL:       s.URL,
		Notes:     s.Notes,
		CreatedAt: tsutil.ParseMillis(s.CreatedAt),
		UpdatedAt: tsutil.ParseMillis(s.UpdatedAt),
	}
}

func secretTypeToRPC(t secrets.Type) SecretType {
	switch t {
	case secrets.PasswordType:
		return PasswordSecret
	case secrets.NoteType:
		return NoteSecret
	default:
		return UnknownSecret
	}
}

func secretTypeFromRPC(s SecretType) secrets.Type {
	switch s {
	case PasswordSecret:
		return secrets.PasswordType
	case NoteSecret:
		return secrets.NoteType
	default:
		return secrets.UnknownType
	}
}
