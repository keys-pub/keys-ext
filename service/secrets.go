package service

import (
	"context"
	"sort"
	strings "strings"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/secret"
	"github.com/keys-pub/keys/util"
	"github.com/pkg/errors"
)

func (s *service) Secret(ctx context.Context, req *SecretRequest) (*SecretResponse, error) {
	secret, err := s.ss.Get(req.ID)
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

	out, _, err := s.ss.Set(sec)
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
	ok, err := s.ss.Delete(req.ID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, keys.NewErrNotFound(req.ID)
	}
	return &SecretRemoveResponse{}, nil
}

func (s *service) Secrets(ctx context.Context, req *SecretsRequest) (*SecretsResponse, error) {
	query := strings.TrimSpace(req.Query)

	sortField := req.SortField
	if sortField == "" {
		sortField = "name"
	}
	sortDirection := req.SortDirection

	secrets, err := s.ss.List(nil)
	if err != nil {
		return nil, err
	}

	out := make([]*Secret, 0, len(secrets))
	for _, s := range secrets {
		sec := secretToRPC(s)
		if query == "" ||
			strings.Contains(sec.Name, query) ||
			strings.Contains(sec.Username, query) ||
			strings.Contains(sec.URL, query) ||
			strings.Contains(sec.Notes, query) {
			out = append(out, sec)
		}
	}

	switch sortField {
	case "id", "name":
	default:
		return nil, errors.Errorf("invalid sort field")
	}

	sort.Slice(out, func(i, j int) bool {
		return secretsSort(out, sortField, sortDirection, i, j)
	})

	return &SecretsResponse{
		Secrets:       out,
		SortField:     sortField,
		SortDirection: sortDirection,
	}, nil
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
		CreatedAt: int64(util.TimeToMillis(s.CreatedAt)),
		UpdatedAt: int64(util.TimeToMillis(s.UpdatedAt)),
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
		CreatedAt: util.TimeFromMillis(int64(s.CreatedAt)),
		UpdatedAt: util.TimeFromMillis(int64(s.UpdatedAt)),
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

func secretsSort(secrets []*Secret, sortField string, sortDirection SortDirection, i, j int) bool {
	switch sortField {
	case "id":
		if sortDirection == SortDesc {
			return secrets[i].ID > secrets[j].ID
		}
		return secrets[i].ID < secrets[j].ID
	case "name":
		if secrets[i].Name == secrets[j].Name {
			return secretsSort(secrets, "id", sortDirection, i, j)
		}
		if sortDirection == SortDesc {
			return secrets[i].Name > secrets[j].Name
		}
		return secrets[i].Name < secrets[j].Name
	default:
		return secretsSort(secrets, "name", sortDirection, i, j)
	}
}
