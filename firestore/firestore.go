package firestore

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/keys-pub/keys"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ keys.DocumentStore = &Firestore{}
var _ keys.Changes = &Firestore{}

// Firestore ...
type Firestore struct {
	uri    string
	client *firestore.Client
	test   bool
}

// NewFirestore creates a Firestore
func NewFirestore(uri string, opts ...option.ClientOption) (*Firestore, error) {
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
	return keys.RetryE(fn)
}

// Set document.
func (f *Firestore) Set(ctx context.Context, path string, b []byte) error {
	fn := func() error { return f.set(ctx, path, b) }
	return keys.RetryE(fn)
}

func normalizePath(p string) string {
	path := keys.Path(p)
	path = strings.TrimPrefix(path, "/")
	return path
}

func (f *Firestore) create(ctx context.Context, path string, b []byte) error {
	path = keys.Path(path)
	if len(keys.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	_, err := doc.Create(context.TODO(), m)
	if err != nil {
		st, ok := status.FromError(err)
		if !ok {
			return errors.Wrapf(processError(err), "failed to create firestore value")
		}
		switch st.Code() {
		case codes.AlreadyExists:
			return keys.NewErrPathExists(path)
		default:
			return errors.Wrapf(processError(err), "failed to create firestore value")
		}
	}
	return nil
}

func (f *Firestore) set(ctx context.Context, path string, b []byte) error {
	path = keys.Path(path)
	if len(keys.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(normalizePath(path))
	m := map[string]interface{}{
		"data": b,
	}

	_, err := doc.Set(context.TODO(), m)
	if err != nil {
		return errors.Wrapf(processError(err), "failed to set firestore value")
	}
	return nil
}

func (f *Firestore) setValue(ctx context.Context, path string, m map[string]interface{}) error {
	path = normalizePath(path)
	if len(keys.PathComponents(path)) != 2 {
		return errors.Errorf("invalid path %s", path)
	}

	logger.Infof(ctx, "Set (Firestore) %s", path)
	doc := f.client.Doc(path)
	_, err := doc.Create(context.TODO(), m)
	if err != nil {
		return errors.Wrapf(processError(err), "failed to set firestore value")
	}
	return nil
}

// timestampField should match firestore tag on keys.Change.
const timestampField = "ts"

func changePath(name string, ref string) string {
	s := strings.ReplaceAll(ref, "/", "-")
	return keys.Path(name, s)
}

// ChangeAdd adds Change.
func (f *Firestore) ChangeAdd(ctx context.Context, name string, ref string) error {
	path := changePath(name, ref)
	// Map should match keys.Change json format
	m := map[string]interface{}{
		"path":         ref,
		timestampField: firestore.ServerTimestamp,
	}
	return f.setValue(ctx, path, m)
}

// Change for name and ref.
func (f *Firestore) Change(ctx context.Context, name string, ref string) (*keys.Change, error) {
	path := changePath(name, ref)
	var change keys.Change
	ok, err := f.getValue(ctx, path, &change)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return &change, nil
}

// Changes ...
func (f *Firestore) Changes(ctx context.Context, name string, from time.Time, limit int) ([]*keys.Change, time.Time, error) {
	col := f.client.Collection(name)
	if col == nil {
		return nil, time.Time{}, nil
	}
	var q firestore.Query
	if from.IsZero() {
		logger.Infof(ctx, "List changes...")
		q = col.Offset(0)
	} else {
		logger.Infof(ctx, "List changes >= %s", from)
		q = col.Where(timestampField, ">=", from)
	}
	iter := q.Documents(context.TODO())

	if limit == 0 {
		limit = 100
	}

	docs := make([]*keys.Change, 0, limit)
	to := from
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, time.Time{}, err
		}
		var change keys.Change
		if err := doc.DataTo(&change); err != nil {
			return nil, time.Time{}, err
		}

		docs = append(docs, &change)
		if change.Timestamp.After(to) {
			to = change.Timestamp
		}
		if len(docs) >= limit {
			break
		}
	}
	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Timestamp == docs[j].Timestamp {
			return docs[i].Path < docs[j].Path
		}
		return docs[i].Timestamp.Before(docs[j].Timestamp)
	})

	return docs, to, nil
}

// GetAll paths.
func (f *Firestore) GetAll(ctx context.Context, paths []string) ([]*keys.Document, error) {
	refs := make([]*firestore.DocumentRef, 0, len(paths))
	for _, p := range paths {
		p = normalizePath(p)
		ref := f.client.Doc(p)
		refs = append(refs, ref)
	}

	docs, err := f.client.GetAll(ctx, refs)
	if err != nil {
		return nil, err
	}
	out := make([]*keys.Document, 0, len(docs))
	for _, doc := range docs {
		if !doc.Exists() {
			continue
		}
		m := doc.Data()
		b, ok := m["data"].([]byte)
		if !ok {
			return nil, errors.Errorf("firestore value missing data")
		}
		// Is there an easier way to get the path?
		path := keys.Path(doc.Ref.Parent.ID + doc.Ref.Path[len(doc.Ref.Parent.Path):])
		d := keys.NewDocument(path, b)
		d.CreatedAt = doc.CreateTime
		d.UpdatedAt = doc.UpdateTime
		out = append(out, d)
	}
	return out, nil
}

// Get ...
func (f *Firestore) Get(ctx context.Context, path string) (*keys.Document, error) {
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

	out := keys.NewDocument(path, b)
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
	if len(keys.PathComponents(path)) != 2 {
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
func (f *Firestore) Documents(ctx context.Context, parent string, opts *keys.DocumentsOpts) (keys.DocumentIterator, error) {
	if opts == nil {
		opts = &keys.DocumentsOpts{}
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

	iter := q.Documents(context.TODO())
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
func (f *Firestore) Collections(ctx context.Context, parent string) (keys.CollectionIterator, error) {
	if keys.Path(parent) != "/" {
		return nil, errors.Errorf("only root collections supported")
	}

	iter := f.client.Collections(ctx)
	return &colsIterator{iter: iter}, nil
}

func (f *Firestore) deleteAll(ctx context.Context) error {
	iter := f.client.Collections(ctx)
	for {
		col, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		docIter := col.DocumentRefs(ctx)
		for {
			doc, err := docIter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			if _, err := doc.Delete(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

// Delete ...
func (f *Firestore) Delete(ctx context.Context, path string) (bool, error) {
	if f.test && keys.Path(path) == "/" {
		if err := f.deleteAll(ctx); err != nil {
			return false, err
		}
		return true, nil
	}

	path = normalizePath(path)
	if len(keys.PathComponents(path)) != 2 {
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
	_, err = doc.Delete(context.TODO())
	if err != nil {
		return false, err
	}
	return true, nil
}
