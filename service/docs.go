package service

import (
	"context"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
)

// Collections (RPC) ...
func (s *service) Collections(ctx context.Context, req *CollectionsRequest) (*CollectionsResponse, error) {
	iter, err := s.db.Collections(ctx, req.Path)
	if err != nil {
		return nil, err
	}
	cols := make([]*Collection, 0, 100)
	for {
		col, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if col == nil {
			break
		}
		// if strings.HasPrefix(col.Path, "/.") {
		// 	continue
		// }
		cols = append(cols, &Collection{
			Path: col.Path,
		})
	}
	iter.Release()
	return &CollectionsResponse{
		Collections: cols,
	}, nil
}

// Documents (RPC) lists local document store.
func (s *service) Documents(ctx context.Context, req *DocumentsRequest) (*DocumentsResponse, error) {
	iter, err := s.db.Documents(ctx, req.Path,
		&ds.DocumentsOpts{
			Prefix: req.Prefix,
			// Index:  int(req.Index),
			// Limit:  int(req.Length),
		})
	if err != nil {
		return nil, err
	}
	out := make([]*Document, 0, 100)
	for {
		doc, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if doc == nil {
			break
		}
		// if req.Pretty {
		// 	if pretty := doc.Pretty(); pretty != nil {
		// 		b = pretty
		// 	}
		// }
		var val string
		if !utf8.Valid(doc.Data) {
			val = string(spew.Sdump(doc.Data))
		} else {
			val = string(doc.Data)
		}
		out = append(out, &Document{
			Path:      doc.Path,
			Value:     val,
			CreatedAt: int64(keys.TimeToMillis(doc.CreatedAt)),
			UpdatedAt: int64(keys.TimeToMillis(doc.UpdatedAt)),
		})
	}
	iter.Release()
	return &DocumentsResponse{
		Documents: out,
	}, nil
}

// DocumentDelete (RPC) ...
func (s *service) DocumentDelete(ctx context.Context, req *DocumentDeleteRequest) (*DocumentDeleteResponse, error) {
	return nil, errors.Errorf("not implemented")
}
