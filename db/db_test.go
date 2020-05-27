package db

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

// testDB returns DB for testing.
// You should defer Close() the result.
func testDB(t *testing.T) (*DB, func()) {
	db := New()
	db.SetTimeNow(tsutil.NewClock().Now)
	path := testPath()
	ctx := context.TODO()
	key := keys.Rand32()
	err := db.OpenAtPath(ctx, path, key)
	require.NoError(t, err)

	return db, func() {
		db.Close()
		os.Remove(path)
	}
}

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("db-test-%s.leveldb", keys.Rand3262()))
}

func TestDB(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()
	testDocumentStore(t, db)
}

func TestDBPath(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	testDocumentStorePath(t, db)
}

func TestDBListOptions(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	testDocumentStoreListOptions(t, db)
}

func TestDBMetadata(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	testMetadata(t, db)
}

func testDocumentStore(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()

	for i := 10; i <= 30; i = i + 10 {
		p := ds.Path("test1", fmt.Sprintf("key%d", i))
		err := dst.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}
	for i := 10; i <= 30; i = i + 10 {
		p := ds.Path("test0", fmt.Sprintf("key%d", i))
		err := dst.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}

	iter, err := dst.Documents(ctx, "test0", nil)
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, "/test0/key10", doc.Path)
	require.Equal(t, "value10", string(doc.Data))
	iter.Release()

	ok, err := dst.Exists(ctx, "/test0/key10")
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = dst.Get(ctx, "/test0/key10")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "value10", string(doc.Data))

	err = dst.Create(ctx, "/test0/key10", []byte{})
	require.EqualError(t, err, "path already exists /test0/key10")
	err = dst.Set(ctx, "/test0/key10", []byte("overwrite"))
	require.NoError(t, err)
	err = dst.Create(ctx, "/test0/key10", []byte("overwrite"))
	require.EqualError(t, err, "path already exists /test0/key10")
	doc, err = dst.Get(ctx, "/test0/key10")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "overwrite", string(doc.Data))

	out, err := dst.GetAll(ctx, []string{"/test0/key10", "/test0/key20"})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, "/test0/key10", out[0].Path)
	require.Equal(t, "/test0/key20", out[1].Path)

	ok, err = dst.Delete(ctx, "/test1/key10")
	require.True(t, ok)
	require.NoError(t, err)
	ok, err = dst.Delete(ctx, "/test1/key10")
	require.False(t, ok)
	require.NoError(t, err)

	ok, err = dst.Exists(ctx, "/test1/key10")
	require.NoError(t, err)
	require.False(t, ok)

	expected := `/test0/key10 overwrite
/test0/key20 value20
/test0/key30 value30
`
	var b bytes.Buffer
	iter, err = dst.Documents(context.TODO(), "test0", nil)
	require.NoError(t, err)
	err = ds.SpewOut(iter, nil, &b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = dst.Documents(context.TODO(), "test0", nil)
	require.NoError(t, err)
	spew, err := ds.Spew(iter, nil)
	require.NoError(t, err)
	require.Equal(t, b.String(), spew.String())
	require.Equal(t, expected, spew.String())
	iter.Release()

	iter, err = dst.Documents(context.TODO(), "test0", &ds.DocumentsOpts{Prefix: "key1", PathOnly: true})
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, "/test0/key10", doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()

	err = dst.Create(ctx, "", []byte{})
	require.EqualError(t, err, "invalid path /")
	err = dst.Set(ctx, "", []byte{})
	require.EqualError(t, err, "invalid path /")

	citer, err := dst.Collections(ctx, "")
	require.NoError(t, err)
	col, err := citer.Next()
	require.NoError(t, err)
	require.Equal(t, "/test0", col.Path)
	col, err = citer.Next()
	require.NoError(t, err)
	require.Equal(t, "/test1", col.Path)
	col, err = citer.Next()
	require.NoError(t, err)
	require.Nil(t, col)
	citer.Release()

	_, err = dst.Collections(ctx, "/test0")
	require.EqualError(t, err, "only root collections supported")
}

func testDocumentStorePath(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()

	err := dst.Create(ctx, "test/1", []byte("value1"))
	require.NoError(t, err)

	doc, err := dst.Get(ctx, "/test/1")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ok, err := dst.Exists(ctx, "/test/1")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = dst.Exists(ctx, "test/1")
	require.NoError(t, err)
	require.True(t, ok)
}

func testDocumentStoreListOptions(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()

	err := dst.Create(ctx, "/test/1", []byte("val1"))
	require.NoError(t, err)
	err = dst.Create(ctx, "/test/2", []byte("val2"))
	require.NoError(t, err)
	err = dst.Create(ctx, "/test/3", []byte("val3"))
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("a", fmt.Sprintf("e%d", i)), []byte("🤓"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b", fmt.Sprintf("ea%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b", fmt.Sprintf("eb%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b", fmt.Sprintf("ec%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("c", fmt.Sprintf("e%d", i)), []byte("😎"))
		require.NoError(t, err)
	}

	iter, err := dst.Documents(ctx, "test", nil)
	require.NoError(t, err)
	paths := []string{}
	for {
		doc, err := iter.Next()
		require.NoError(t, err)
		if doc == nil {
			break
		}
		paths = append(paths, doc.Path)
	}
	require.Equal(t, []string{"/test/1", "/test/2", "/test/3"}, paths)
	iter.Release()

	iter, err = dst.Documents(context.TODO(), "test", nil)
	require.NoError(t, err)
	b, err := ds.Spew(iter, nil)
	require.NoError(t, err)
	expected := `/test/1 val1
/test/2 val2
/test/3 val3
`
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = dst.Documents(ctx, "b", &ds.DocumentsOpts{Prefix: "eb"})
	require.NoError(t, err)
	paths = []string{}
	for {
		doc, err := iter.Next()
		require.NoError(t, err)
		if doc == nil {
			break
		}
		paths = append(paths, doc.Path)
	}
	iter.Release()
	require.Equal(t, []string{"/b/eb1", "/b/eb2"}, paths)
}

func testMetadata(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()

	err := dst.Create(ctx, "/test/key1", []byte("value1"))
	require.NoError(t, err)

	doc, err := dst.Get(ctx, "/test/key1")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, int64(1234567890001), tsutil.Millis(doc.CreatedAt))

	err = dst.Set(ctx, "/test/key1", []byte("value1b"))
	require.NoError(t, err)

	doc, err = dst.Get(ctx, "/test/key1")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, int64(1234567890001), tsutil.Millis(doc.CreatedAt))
	require.Equal(t, int64(1234567890002), tsutil.Millis(doc.UpdatedAt))
}

func ExampleDB_OpenAtPath() {
	db := New()
	defer db.Close()

	key := keys.Rand32()
	ctx := context.TODO()
	path := filepath.Join(os.TempDir(), "example-db-open.db")
	if err := db.OpenAtPath(ctx, path, key); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_Create() {
	db := New()
	defer db.Close()

	key := keys.Rand32()
	ctx := context.TODO()
	path := filepath.Join(os.TempDir(), "example-db-create.db")
	if err := db.OpenAtPath(ctx, path, key); err != nil {
		log.Fatal(err)
	}

	if err := db.Create(context.TODO(), "/test/1", []byte{0x01, 0x02, 0x03}); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_Get() {
	db := New()
	defer db.Close()

	key := keys.Rand32()
	ctx := context.TODO()
	path := filepath.Join(os.TempDir(), "example-db-get.db")
	if err := db.OpenAtPath(ctx, path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	if err := db.Set(ctx, ds.Path("collection1", "doc1"), []byte("hi")); err != nil {
		log.Fatal(err)
	}

	doc, err := db.Get(ctx, ds.Path("collection1", "doc1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got %s\n", string(doc.Data))
	// Output:
	// Got hi
}

func ExampleDB_Set() {
	db := New()
	defer db.Close()

	key := keys.Rand32()
	ctx := context.TODO()
	path := filepath.Join(os.TempDir(), "example-db-set.db")
	if err := db.OpenAtPath(ctx, path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	if err := db.Set(ctx, ds.Path("collection1", "doc1"), []byte("hi")); err != nil {
		log.Fatal(err)
	}

	doc, err := db.Get(ctx, ds.Path("collection1", "doc1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got %s\n", string(doc.Data))
	// Output:
	// Got hi
}

func ExampleDB_Documents() {
	db := New()
	defer db.Close()

	key := keys.Rand32()
	ctx := context.TODO()
	path := filepath.Join(os.TempDir(), "example-db-documents.db")
	if err := db.OpenAtPath(ctx, path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	if err := db.Set(ctx, ds.Path("collection1", "doc1"), []byte("hi")); err != nil {
		log.Fatal(err)
	}

	iter, err := db.Documents(ctx, ds.Path("collection1"), nil)
	if err != nil {
		log.Fatal(err)
	}
	for {
		doc, err := iter.Next()
		if err != nil {
			log.Fatal(err)
		}
		if doc == nil {
			break
		}
		fmt.Printf("%s: %s\n", doc.Path, string(doc.Data))
	}
	// Output:
	// /collection1/doc1: hi
}

func TestDBGetSetLarge(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	large := bytes.Repeat([]byte{0x01}, 10*1024*1024)

	err := db.Set(context.TODO(), "/test/key1", large)
	require.NoError(t, err)

	doc, err := db.Get(context.TODO(), "/test/key1")
	require.NoError(t, err)
	require.Equal(t, large, doc.Data)
}

func TestDBGetSetEmpty(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	err := db.Set(context.TODO(), "/test/key1", []byte{})
	require.NoError(t, err)

	doc, err := db.Get(context.TODO(), "/test/key1")
	require.NoError(t, err)
	require.Equal(t, []byte{}, doc.Data)
}

func TestDeleteAll(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	err := db.Set(context.TODO(), "/test/key1", []byte("val1"))
	require.NoError(t, err)
	err = db.Set(context.TODO(), "/test/key2", []byte("val2"))
	require.NoError(t, err)

	err = db.DeleteAll(context.TODO(), []string{"/test/key1", "/test/key2", "/test/key3"})
	require.NoError(t, err)

	doc, err := db.Get(context.TODO(), "/test/key1")
	require.NoError(t, err)
	require.Nil(t, doc)
	doc, err = db.Get(context.TODO(), "/test/key2")
	require.NoError(t, err)
	require.Nil(t, doc)
}
