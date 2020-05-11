package firestore

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/util"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ ds.DocumentStore = &Firestore{}
var _ ds.Changes = &Firestore{}

// Firestore ...
type Firestore struct {
	uri    string
	client *firestore.Client
}

// New creates a Firestore
func New(uri string, opts ...option.ClientOption) (*Firestore, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "firestore" {
		return nil, errors.Errorf("invalid scheme, should be like firestore://projectid")
	}
	projectID := u.Host

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create firestore client")
	}
	fs := &Firestore{
		uri:    uri,
		client: client,
	}
	return fs, nil
}

// URI ...
func (f *Firestore) URI() string {
	return f.uri
}

// Create document.
func (f *Firestore) Create(ctx context.Context, path string, b []byte) error {
	fn := func() error { return f.create(ctx, path, b) }
	return util.RetryE(fn)
}

// Set document.
func (f *Firestore) Set(ctx context.Context, path string, b []byte) error {
	fn := func() error { return f.set(ctx, path, b) }
	return util.RetryE(fn)
}

func normalizePath(p string) string {
	path := ds.Path(p)
	path = strings.TrimPrefix(path, "/")
	return path
}

func (f *Firestore) create(ctx context.Context, path string, b []byte) error {
	path = ds.Path(path)
	if len(ds.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	_, err := doc.Create(ctx, m)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return errors.Wrapf(processError(err), "failed to create firestore value")
		}
		switch st.Code() {
		case codes.AlreadyExists:
			return ds.NewErrPathExists(path)
		default:
			return errors.Wrapf(processError(err), "failed to create firestore value")
		}
	}
	return nil
}

func (f *Firestore) set(ctx context.Context, path string, b []byte) error {
	path = ds.Path(path)
	if len(ds.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	_, err := doc.Set(ctx, m)
	if err != nil {
		return errors.Wrapf(processError(err), "failed to set firestore value")
	}
	return nil
}

func (f *Firestore) createValue(ctx context.Context, path string, m map[string]interface{}) error {
	path = normalizePath(path)
	if len(ds.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(path)
	_, err := doc.Create(ctx, m)
	if err != nil {
		return errors.Wrapf(processError(err), "failed to set firestore value")
	}
	return nil
}

// timestampField should match firestore tag on keys.Change.
const timestampField = "ts"

// ChangeAdd adds Change.
func (f *Firestore) ChangeAdd(ctx context.Context, name string, id string, ref string) error {
	path := ds.Path(name, id)
	// Map should match keys.Change json format
	m := map[string]interface{}{
		"path":         ref,
		timestampField: firestore.ServerTimestamp,
	}
	return f.createValue(ctx, path, m)
}

// Changes ...
func (f *Firestore) Changes(ctx context.Context, name string, ts time.Time, limit int, direction ds.Direction) ([]*ds.Change, time.Time, error) {
	col := f.client.Collection(name)
	if col == nil {
		return nil, time.Time{}, nil
	}

	var q firestore.Query
	switch direction {
	case ds.Ascending:
		if ts.IsZero() {
			logger.Infof(ctx, "List changes (asc)...")
			q = col.OrderBy(timestampField, firestore.Asc)
		} else {
			logger.Infof(ctx, "List changes (asc >= %s)", ts)
			q = col.OrderBy(timestampField, firestore.Asc).Where(timestampField, ">=", ts)
		}
	case ds.Descending:
		if ts.IsZero() {
			logger.Infof(ctx, "List changes (desc)...")
			q = col.OrderBy(timestampField, firestore.Desc)
		} else {
			logger.Infof(ctx, "List changes (desc <= %s)", ts)
			q = col.OrderBy(timestampField, firestore.Desc).Where(timestampField, "<=", ts)
		}
	}

	iter := q.Documents(ctx)

	if limit == 0 {
		limit = 100
	}

	out := make([]*ds.Change, 0, limit)

	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, time.Time{}, err
		}
		var change ds.Change
		if err := doc.DataTo(&change); err != nil {
			return nil, time.Time{}, err
		}

		out = append(out, &change)
		if len(out) >= limit {
			break
		}
	}

	to := ts
	if len(out) != 0 {
		to = out[len(out)-1].Timestamp
	}

	return out, to, nil
}

// GetAll paths.
func (f *Firestore) GetAll(ctx context.Context, paths []string) ([]*ds.Document, error) {
	refs := make([]*firestore.DocumentRef, 0, len(paths))
	for _, p := range paths {
		p = normalizePath(p)
		ref := f.client.Doc(p)
		refs = append(refs, ref)
	}

	res, err := f.client.GetAll(ctx, refs)
	if err != nil {
		return nil, err
	}
	out := make([]*ds.Document, 0, len(res))
	for _, doc := range res {
		if !doc.Exists() {
			continue
		}
		m := doc.Data()
		b, ok := m["data"].([]byte)
		if !ok {
			return nil, errors.Errorf("firestore value missing data")
		}
		// Is there an easier way to get the path?
		path := ds.Path(doc.Ref.Parent.ID + doc.Ref.Path[len(doc.Ref.Parent.Path):])
		newDoc := ds.NewDocument(path, b)
		newDoc.CreatedAt = doc.CreateTime
		newDoc.UpdatedAt = doc.UpdateTime
		out = append(out, newDoc)

	}
	return out, nil
}

// Get ...
func (f *Firestore) Get(ctx context.Context, path string) (*ds.Document, error) {
	logger.Infof(ctx, "Get %s", path)
	doc, err := f.get(ctx, path)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, nil
	}
	m := doc.Data()
	b, ok := m["data"].([]byte)
	if !ok {
		return nil, errors.Errorf("firestore value missing data")
	}

	logger.Debugf(ctx, "Create time %s", doc.CreateTime)
	logger.Debugf(ctx, "Update time %s", doc.UpdateTime)

	out := ds.NewDocument(path, b)
	out.CreatedAt = doc.CreateTime
	out.UpdatedAt = doc.UpdateTime
	return out, nil
}

// Exists returns true if path exists.
func (f *Firestore) Exists(ctx context.Context, path string) (bool, error) {
	doc, err := f.get(ctx, path)
	if err != nil {
		return false, err
	}
	return doc != nil, nil
}

func (f *Firestore) get(ctx context.Context, path string) (*firestore.DocumentSnapshot, error) {
	path = normalizePath(path)
	if len(ds.PathComponents(path)) != 2 {
		return nil, errors.Errorf("invalid path %s", path)
	}

	doc, err := f.client.Doc(path).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return doc, nil
}

func (f *Firestore) getValue(ctx context.Context, path string, v interface{}) (bool, error) {
	doc, err := f.get(ctx, path)
	if err != nil {
		return false, err
	}
	if doc == nil {
		return false, nil
	}
	if err := doc.DataTo(v); err != nil {
		return false, err
	}
	return true, nil
}

// Documents ...
func (f *Firestore) Documents(ctx context.Context, parent string, opts *ds.DocumentsOpts) (ds.DocumentIterator, error) {
	if opts == nil {
		opts = &ds.DocumentsOpts{}
	}
	// TODO: Handle context Done()
	path := normalizePath(parent)

	if path == "" {
		return nil, errors.Errorf("list root not supported")
	}

	logger.Infof(ctx, "Query (firestore) %q (%+v)...", path, opts)
	col := f.client.Collection(path)
	if col == nil {
		return &docsIterator{parent: path}, nil
	}
	q := col.Offset(0)

	if opts.Prefix != "" {
		q = q.Where(firestore.DocumentID, ">=", col.Doc(opts.Prefix))
	}

	// if opts.OrderBy != "" {
	// 	q = col.OrderBy(opts.OrderBy, firestore.Asc)
	// }
	// if opts.StartAt != "" {
	// 	q = q.StartAt(opts.StartAt)
	// }
	if opts.Index > 0 {
		q = q.Offset(opts.Index)
	}
	if opts.Limit > 0 {
		q = q.Limit(opts.Limit)
	}

	iter := q.Documents(ctx)
	return &docsIterator{iter: iter, parent: path, prefix: opts.Prefix, pathOnly: opts.PathOnly}, nil
}

// processError tries to unmarshal Firebase JSON error, if it fails it returns
// what was passed in.
func processError(ferr error) error {
	if strings.HasPrefix(ferr.Error(), "{") {
		var jsonErr struct{ Error string }
		if err := json.Unmarshal([]byte(ferr.Error()), &jsonErr); err == nil {
			if jsonErr.Error != "" {
				return errors.Errorf("firestore error: %s", jsonErr.Error)
			}
		}
	}
	return ferr
}

// Collections ...
func (f *Firestore) Collections(ctx context.Context, parent string) (ds.CollectionIterator, error) {
	if ds.Path(parent) != "/" {
		return nil, errors.Errorf("only root collections supported")
	}

	iter := f.client.Collections(ctx)
	return &colsIterator{iter: iter}, nil
}

// Delete ...
func (f *Firestore) Delete(ctx context.Context, path string) (bool, error) {
	return f.delete(ctx, path)
}

func (f *Firestore) delete(ctx context.Context, path string) (bool, error) {
	path = normalizePath(path)
	if len(ds.PathComponents(path)) != 2 {
		return false, errors.Errorf("invalid path %s", path)
	}

	exists, err := f.Exists(ctx, path)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, nil
	}

	doc := f.client.Doc(path)
	_, err = doc.Delete(ctx)
	if err != nil {
		return false, err
	}
	return true, nil
}

// DeleteAll ...
func (f *Firestore) DeleteAll(ctx context.Context, paths []string) error {
	for _, p := range paths {
		if _, err := f.delete(ctx, p); err != nil {
			return err
		}
	}
	return nil
}
