package service

import (
	"context"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/pkg/errors"
)

// Collections (RPC) ...
func (s *service) Collections(ctx context.Context, req *CollectionsRequest) (*CollectionsResponse, error) {
	switch req.DB {
	case "", "service":
		return s.serviceCollections(ctx, req.Parent)
	case "vault":
		return s.vaultCollections(ctx, req.Parent)
	default:
		return nil, errors.Errorf("invalid db %s", req.DB)
	}
}

func (s *service) serviceCollections(ctx context.Context, parent string) (*CollectionsResponse, error) {
	cols, err := s.db.Collections(ctx, parent)
	if err != nil {
		return nil, err
	}
	return &CollectionsResponse{Collections: collectionsToRPC(cols)}, nil
}

func (s *service) vaultCollections(ctx context.Context, parent string) (*CollectionsResponse, error) {
	cols, err := vault.Collections(s.vault.Store(), "")
	if err != nil {
		return nil, err
	}
	out := make([]*Collection, 0, len(cols))
	for _, c := range cols {
		out = append(out, &Collection{Path: c})
	}
	return &CollectionsResponse{Collections: out}, nil
}

func collectionsToRPC(cols []*dstore.Collection) []*Collection {
	out := make([]*Collection, 0, len(cols))
	for _, c := range cols {
		out = append(out, &Collection{Path: c.Path})
	}
	return out
}

// Documents (RPC) lists document from db or vault.
func (s *service) Documents(ctx context.Context, req *DocumentsRequest) (*DocumentsResponse, error) {
	out := make([]*Document, 0, 100)

	dataToString := func(b []byte) string {
		var val string
		if !utf8.Valid(b) {
			val = string(spew.Sdump(b))
		} else {
			val = string(b)
		}
		return val
	}

	switch req.DB {
	case "", "service":
		docs, err := s.db.Documents(ctx, "", dstore.Prefix(req.Prefix))
		if err != nil {
			return nil, err
		}
		for _, doc := range docs {
			out = append(out, &Document{
				Path:      doc.Path,
				Value:     dataToString(doc.Data()),
				CreatedAt: tsutil.Millis(doc.CreatedAt),
				UpdatedAt: tsutil.Millis(doc.UpdatedAt),
			})
		}

	case "vault":
		entries, err := s.vault.Store().List(&vault.ListOptions{Prefix: req.Prefix})
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			out = append(out, &Document{
				Path:  entry.Path,
				Value: dataToString(entry.Data),
			})
		}
	default:
		return nil, errors.Errorf("unrecognized db")
	}

	return &DocumentsResponse{
		Documents: out,
	}, nil
}

// DocumentDelete (RPC) ...
func (s *service) DocumentDelete(ctx context.Context, req *DocumentDeleteRequest) (*DocumentDeleteResponse, error) {
	if req.Path == "" {
		return nil, errors.Errorf("invalid path")
	}

	var err error
	var ok bool
	switch req.DB {
	case "", "service":
		ok, err = s.db.Delete(ctx, req.Path)
	case "vault":
		ok, err = s.vault.Store().Delete(req.Path)
	}
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("path not found %s", req.Path)
	}
	return &DocumentDeleteResponse{}, nil
}
