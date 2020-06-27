package firestore

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ ds.DocumentStore = &Firestore{}
var _ ds.Events = &Firestore{}

// Firestore is a DocumentStore implemented on Google Cloud Firestore.
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
//
// Paths can be nested as long as they are even length components.
// For example,
//
//   collection1/key1 (OK)
//   collection1/key1/collection2/key2 (OK)
//   collection1 (INVALID)
//   collection1/key1/collection2 (INVALID)
//
func (f *Firestore) Create(ctx context.Context, path string, b []byte) error {
	fn := func() error { return f.create(ctx, path, b) }
	return keys.RetryE(fn)
}

// Set document.
//
// Paths can be nested as long as they are even length components.
// For example,
//
//   collection1/key1 (OK)
//   collection1/key1/collection2/key2 (OK)
//   collection1 (INVALID)
//   collection1/key1/collection2 (INVALID)
//
func (f *Firestore) Set(ctx context.Context, path string, b []byte) error {
	fn := func() error { return f.set(ctx, path, b) }
	return keys.RetryE(fn)
}

func normalizePath(p string) string {
	path := ds.Path(p)
	path = strings.TrimPrefix(path, "/")
	return path
}

func (f *Firestore) create(ctx context.Context, path string, b []byte) error {
	logger.Infof(ctx, "Create (Firestore) %s", path)
	path, err := checkPath(path)
	if err != nil {
		return err
	}

	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	if _, err := doc.Create(ctx, m); err != nil {
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

func checkPath(path string) (string, error) {
	path = ds.Path(path)
	if len(ds.PathComponents(path))%2 != 0 {
		return "", errors.Errorf("invalid path %s", path)
	}
	if path == "/" {
		return "", errors.Errorf("invalid path /")
	}
	return path, nil
}

func (f *Firestore) set(ctx context.Context, path string, b []byte) error {
	logger.Infof(ctx, "Set (Firestore) %s", path)
	path, err := checkPath(path)
	if err != nil {
		return err
	}
	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	if _, err := doc.Set(ctx, m); err != nil {
		return errors.Wrapf(processError(err), "failed to set firestore value")
	}
	return nil
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

		pc := ds.PathComponents(doc.Ref.Path)
		path := ds.Path(pc[5:])
		m := doc.Data()
		b, ok := m["data"].([]byte)
		if !ok {
			logger.Warningf(ctx, "firestore value missing data at %s", path)
		}
		// Is there an easier way to get the path?
		newDoc := ds.NewDocument(path, b)
		newDoc.CreatedAt = doc.CreateTime
		newDoc.UpdatedAt = doc.UpdateTime
		out = append(out, newDoc)

	}
	return out, nil
}

// Get ...
func (f *Firestore) Get(ctx context.Context, path string) (*ds.Document, error) {
	logger.Infof(ctx, "Get (Firestore) %s", path)
	path, err := checkPath(path)
	if err != nil {
		return nil, err
	}

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
		logger.Warningf(ctx, "firestore value missing data at %s", path)
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
	doc, err := f.client.Doc(path).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return doc, nil
}

// DocumentIterator ...
func (f *Firestore) DocumentIterator(ctx context.Context, parent string, opt ...ds.DocumentsOption) (ds.DocumentIterator, error) {
	opts := ds.NewDocumentsOptions(opt...)

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
	return &docsIterator{iter: iter, parent: path, prefix: opts.Prefix, pathOnly: opts.NoData}, nil
}

// Documents not implemented on Firestore, use DocumentIterator.
func (f *Firestore) Documents(ctx context.Context, parent string, opt ...ds.DocumentsOption) ([]*ds.Document, error) {
	return nil, errors.Errorf("not implemented")
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
func (f *Firestore) Collections(ctx context.Context, parent string) ([]*ds.Collection, error) {
	if ds.Path(parent) != "/" {
		return nil, errors.Errorf("only root collections supported")
	}

	iter := f.client.Collections(ctx)
	cols := []*ds.Collection{}
	for {
		col, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		cols = append(cols, &ds.Collection{Path: ds.Path(col.ID)})
	}
	return cols, nil
}

// Delete ...
func (f *Firestore) Delete(ctx context.Context, path string) (bool, error) {
	return f.delete(ctx, path)
}

func (f *Firestore) delete(ctx context.Context, path string) (bool, error) {
	path = normalizePath(path)
	if len(ds.PathComponents(path))%2 != 0 {
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
