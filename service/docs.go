package service

import (
	"context"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys/docs"
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
	cols, err := s.vault.Collections(parent)
	if err != nil {
		return nil, err
	}
	return &CollectionsResponse{Collections: collectionsToRPC(cols)}, nil
}

func collectionsToRPC(cols []*docs.Collection) []*Collection {
	out := make([]*Collection, 0, len(cols))
	for _, c := range cols {
		out = append(out, &Collection{Path: c.Path})
	}
	return out
}

// Documents (RPC) lists document from db or vault.
func (s *service) Documents(ctx context.Context, req *DocumentsRequest) (*DocumentsResponse, error) {
	var ds []*docs.Document
	var dsErr error
	switch req.DB {
	case "", "service":
		ds, dsErr = s.db.Documents(ctx, "", docs.Prefix(req.Prefix))
	case "vault":
		ds, dsErr = s.vault.Documents(docs.Prefix(req.Prefix))
	}
	if dsErr != nil {
		return nil, dsErr
	}
	out := make([]*Document, 0, 100)
	for _, doc := range ds {
		var val string
		if !utf8.Valid(doc.Data) {
			val = string(spew.Sdump(doc.Data))
		} else {
			val = string(doc.Data)
		}
		out = append(out, &Document{
			Path:      doc.Path,
			Value:     val,
			CreatedAt: int64(tsutil.Millis(doc.CreatedAt)),
			UpdatedAt: int64(tsutil.Millis(doc.UpdatedAt)),
		})
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
	ok, err := s.db.Delete(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("path not found %s", req.Path)
	}
	return &DocumentDeleteResponse{}, nil
}
