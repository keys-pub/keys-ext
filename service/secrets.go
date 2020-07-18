package service

import (
	"context"
	"strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/secret"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

func (s *service) Secret(ctx context.Context, req *SecretRequest) (*SecretResponse, error) {
	secret, err := s.vault.Secret(req.ID)
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
	sec := secretFromRPC(req.Secret)
	if sec.ID == "" {
		sec.ID = secret.RandID()
	}

	if sec.Type == secret.UnknownType {
		return nil, errors.Errorf("unknown secret type")
	}

	name := strings.TrimSpace(sec.Name)
	if name == "" {
		return nil, errors.Errorf("name not specified")
	}

	out, _, err := s.vault.SaveSecret(sec)
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

	secrets, err := s.vault.Secrets(
		vault.Secrets.Query(req.Query),
		vault.Secrets.Sort(req.SortField),
		vault.Secrets.SortDirection(sortDirectionToVault(req.SortDirection)),
	)
	if err != nil {
		return nil, err
	}
	out := secretsToRPC(secrets)

	return &SecretsResponse{
		Secrets:       out,
		SortField:     req.SortField,
		SortDirection: req.SortDirection,
	}, nil
}

func sortDirectionToVault(d SortDirection) vault.SortDirection {
	switch d {
	case SortAsc:
		return vault.Ascending
	case SortDesc:
		return vault.Descending
	default:
		return vault.Ascending
	}
}

func secretsToRPC(ss []*secret.Secret) []*Secret {
	out := make([]*Secret, 0, len(ss))
	for _, s := range ss {
		out = append(out, secretToRPC(s))
	}
	return out
}

func secretToRPC(s *secret.Secret) *Secret {
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

func secretFromRPC(s *Secret) *secret.Secret {
	return &secret.Secret{
		ID:        s.ID,
		Name:      s.Name,
		Type:      secretTypeFromRPC(s.Type),
		Username:  s.Username,
		Password:  s.Password,
		URL:       s.URL,
		Notes:     s.Notes,
		CreatedAt: tsutil.ConvertMillis(s.CreatedAt),
		UpdatedAt: tsutil.ConvertMillis(s.UpdatedAt),
	}
}

func secretTypeToRPC(t secret.Type) SecretType {
	switch t {
	case secret.PasswordType:
		return PasswordSecret
	case secret.ContactType:
		return ContactSecret
	case secret.CardType:
		return CardSecret
	case secret.NoteType:
		return NoteSecret
	default:
		return UnknownSecret
	}
}

func secretTypeFromRPC(s SecretType) secret.Type {
	switch s {
	case PasswordSecret:
		return secret.PasswordType
	case ContactSecret:
		return secret.ContactType
	case CardSecret:
		return secret.CardType
	case NoteSecret:
		return secret.NoteType
	default:
		return secret.UnknownType
	}
}
