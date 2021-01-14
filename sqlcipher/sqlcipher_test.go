package sqlcipher_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/sqlcipher"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
)

// testDB returns DB for testing.
// You should defer Close() the result.
func testDB(t *testing.T) (*sqlcipher.DB, func()) {
	path := testPath()
	key := keys.Rand32()
	return testDBWithOpts(t, path, key)
}

func testDBWithOpts(t *testing.T, path string, key sqlcipher.SecretKey) (*sqlcipher.DB, func()) {
	db := sqlcipher.New()
	db.SetClock(tsutil.NewTestClock())
	ctx := context.TODO()
	err := db.OpenAtPath(ctx, path, key)
	require.NoError(t, err)

	return db, func() {
		db.Close()
		os.Remove(path)
	}
}

func testPath() string {
	return filepath.Join(os.TempDir(), fmt.Sprintf("%s.db", keys.RandFileName()))
}

func TestDB(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	require.True(t, db.IsOpen())

	ctx := context.TODO()

	for i := 10; i <= 30; i = i + 10 {
		p := dstore.Path("test1", fmt.Sprintf("key%d", i))
		err := db.Create(ctx, p, dstore.Data([]byte(fmt.Sprintf("value%d", i))))
		require.NoError(t, err)
	}
	for i := 10; i <= 30; i = i + 10 {
		p := dstore.Path("test0", fmt.Sprintf("key%d", i))
		err := db.Create(ctx, p, dstore.Data([]byte(fmt.Sprintf("value%d", i))))
		require.NoError(t, err)
	}

	iter, err := db.DocumentIterator(ctx, "test0")
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "/test0/key10", doc.Path)
	require.Equal(t, "value10", string(doc.Data()))
	iter.Release()

	out, err := db.Documents(ctx, "test0")
	require.NoError(t, err)
	require.Equal(t, 3, len(out))
	require.Equal(t, "/test0/key10", out[0].Path)
	require.Equal(t, "value10", string(out[0].Data()))

	ok, err := db.Exists(ctx, "/test0/key10")
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = db.Get(ctx, "/test0/key10")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "value10", string(doc.Data()))

	err = db.Create(ctx, "/test0/key10", dstore.Data([]byte{}))
	require.EqualError(t, err, "path already exists /test0/key10")
	err = db.Set(ctx, "/test0/key10", dstore.Data([]byte("overwrite")))
	require.NoError(t, err)
	err = db.Create(ctx, "/test0/key10", dstore.Data([]byte("overwrite")))
	require.EqualError(t, err, "path already exists /test0/key10")
	doc, err = db.Get(ctx, "/test0/key10")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "overwrite", string(doc.Data()))

	out, err = db.GetAll(ctx, []string{"/test0/key10", "/test0/key20"})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, "/test0/key10", out[0].Path)
	require.Equal(t, "/test0/key20", out[1].Path)

	ok, err = db.Delete(ctx, "/test1/key10")
	require.True(t, ok)
	require.NoError(t, err)
	ok, err = db.Delete(ctx, "/test1/key10")
	require.False(t, ok)
	require.NoError(t, err)

	ok, err = db.Exists(ctx, "/test1/key10")
	require.NoError(t, err)
	require.False(t, ok)

	expected := `/test0/key10 overwrite
/test0/key20 value20
/test0/key30 value30
`
	var b bytes.Buffer
	iter, err = db.DocumentIterator(context.TODO(), "test0")
	require.NoError(t, err)
	err = dstore.SpewOut(iter, &b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = db.DocumentIterator(context.TODO(), "test0")
	require.NoError(t, err)
	spew, err := dstore.Spew(iter)
	require.NoError(t, err)
	require.Equal(t, b.String(), spew.String())
	require.Equal(t, expected, spew.String())
	iter.Release()

	iter, err = db.DocumentIterator(context.TODO(), "test0", dstore.Prefix("key1"), dstore.NoData())
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, "/test0/key10", doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()

	err = db.Create(ctx, "", dstore.Data([]byte{}))
	require.EqualError(t, err, "invalid path /")
	err = db.Set(ctx, "", dstore.Data([]byte{}))
	require.EqualError(t, err, "invalid path /")

	cols, err := db.Collections(ctx, "")
	require.NoError(t, err)
	require.Equal(t, "/test0", cols[0].Path)
	require.Equal(t, "/test1", cols[1].Path)

	_, err = db.Collections(ctx, "/test0")
	require.EqualError(t, err, "only root collections supported")
}

func TestDocumentsPath(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	ctx := context.TODO()

	err := db.Create(ctx, "test/1", dstore.Data([]byte("value1")))
	require.NoError(t, err)

	doc, err := db.Get(ctx, "/test/1")
	require.NoError(t, err)
	require.NotNil(t, doc)

	ok, err := db.Exists(ctx, "/test/1")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = db.Exists(ctx, "test/1")
	require.NoError(t, err)
	require.True(t, ok)

	err = db.Create(ctx, dstore.Path("test", "key2", "col2", "key3"), dstore.Data([]byte("value3")))
	require.NoError(t, err)

	doc, err = db.Get(ctx, dstore.Path("test", "key2", "col2", "key3"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, []byte("value3"), doc.Data())

	cols, err := db.Collections(ctx, "")
	require.NoError(t, err)
	require.Equal(t, "/test", cols[0].Path)
}

func TestDBListOptions(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()

	ctx := context.TODO()

	err := db.Create(ctx, "/test/1", dstore.Data([]byte("val1")))
	require.NoError(t, err)
	err = db.Create(ctx, "/test/2", dstore.Data([]byte("val2")))
	require.NoError(t, err)
	err = db.Create(ctx, "/test/3", dstore.Data([]byte("val3")))
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		err := db.Create(ctx, dstore.Path("a", fmt.Sprintf("e%d", i)), dstore.Data([]byte("🤓")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := db.Create(ctx, dstore.Path("b", fmt.Sprintf("ea%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := db.Create(ctx, dstore.Path("b", fmt.Sprintf("eb%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := db.Create(ctx, dstore.Path("b", fmt.Sprintf("ec%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := db.Create(ctx, dstore.Path("c", fmt.Sprintf("e%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}

	iter, err := db.DocumentIterator(ctx, "test")
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

	iter, err = db.DocumentIterator(context.TODO(), "test")
	require.NoError(t, err)
	b, err := dstore.Spew(iter)
	require.NoError(t, err)
	expected := `/test/1 val1
/test/2 val2
/test/3 val3
`
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = db.DocumentIterator(context.TODO(), "test", dstore.NoData())
	require.NoError(t, err)
	b, err = dstore.Spew(iter)
	require.NoError(t, err)
	expected = `/test/1
/test/2
/test/3
`
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = db.DocumentIterator(ctx, "b", dstore.Prefix("eb"))
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

func TestMetadata(t *testing.T) {
	ctx := context.TODO()
	db, closeFn := testDB(t)
	defer closeFn()

	err := db.Create(ctx, "/test/key1", dstore.Data([]byte("value1")))
	require.NoError(t, err)

	doc, err := db.Get(ctx, "/test/key1")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, int64(1234567890001), tsutil.Millis(doc.CreatedAt))

	err = db.Set(ctx, "/test/key1", dstore.Data([]byte("value1b")))
	require.NoError(t, err)

	doc, err = db.Get(ctx, "/test/key1")
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, int64(1234567890001), tsutil.Millis(doc.CreatedAt))
	require.Equal(t, int64(1234567890002), tsutil.Millis(doc.UpdatedAt))
}

func TestDocumentSetTo(t *testing.T) {
	ctx := context.TODO()
	db, closeFn := testDB(t)
	defer closeFn()

	type Test struct {
		Int    int    `json:"n,omitempty"`
		String string `json:"s,omitempty"`
		Bytes  []byte `json:"b,omitempty"`
	}
	val := &Test{
		Int:    1,
		String: "teststring",
		Bytes:  []byte("testbytes"),
	}

	path := dstore.Path("test", "key1")
	err := db.Create(ctx, path, dstore.From(val))
	require.NoError(t, err)

	doc, err := db.Get(ctx, path)
	require.NoError(t, err)

	var out Test
	err = doc.To(&out)
	require.NoError(t, err)
	require.Equal(t, val, &out)
}

func TestDocumentMerge(t *testing.T) {
	ctx := context.TODO()
	db, closeFn := testDB(t)
	defer closeFn()

	type Test struct {
		Int    int    `json:"n,omitempty"`
		String string `json:"s,omitempty"`
		Bytes  []byte `json:"b,omitempty"`
	}
	val := &Test{
		Int:    1,
		String: "teststring",
		Bytes:  []byte("testbytes"),
	}

	path := dstore.Path("test", "key1")
	err := db.Set(ctx, path, dstore.From(val))
	require.NoError(t, err)

	val2 := &Test{String: "teststring-merge"}
	err = db.Set(ctx, path, dstore.From(val2), dstore.MergeAll())
	require.NoError(t, err)

	doc, err := db.Get(ctx, path)
	require.NoError(t, err)

	var out Test
	err = doc.To(&out)
	require.NoError(t, err)
	expected := &Test{
		Int:    1,
		String: "teststring-merge",
		Bytes:  []byte("testbytes"),
	}
	require.Equal(t, expected, &out)
}

func TestDBGetSetLarge(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	large := bytes.Repeat([]byte{0x01}, 10*1024*1024)

	err := db.Set(context.TODO(), "/test/key1", dstore.Data(large))
	require.NoError(t, err)

	doc, err := db.Get(context.TODO(), "/test/key1")
	require.NoError(t, err)
	require.Equal(t, large, doc.Data())
}

func TestDBGetSetEmpty(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	err := db.Set(context.TODO(), "/test/key1", dstore.Data([]byte{}))
	require.NoError(t, err)

	doc, err := db.Get(context.TODO(), "/test/key1")
	require.NoError(t, err)
	require.Equal(t, []byte{}, doc.Data())
}

func TestDeleteAll(t *testing.T) {
	// SetLogger(NewLogger(DebugLevel))
	db, closeFn := testDB(t)
	defer closeFn()

	err := db.Set(context.TODO(), "/test/key1", dstore.Data([]byte("val1")))
	require.NoError(t, err)
	err = db.Set(context.TODO(), "/test/key2", dstore.Data([]byte("val2")))
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

func TestUpdate(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	ctx := context.TODO()

	err := db.Create(ctx, dstore.Path("test", "key1"), dstore.Data([]byte("val1")))
	require.NoError(t, err)

	err = db.Set(ctx, dstore.Path("test", "key1"), map[string]interface{}{"index": 1, "info": "testinfo"}, dstore.MergeAll())
	require.NoError(t, err)

	doc, err := db.Get(ctx, dstore.Path("test", "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)

	b := doc.Bytes("data")
	require.Equal(t, []byte("val1"), b)

	index, _ := doc.Int("index")
	require.Equal(t, 1, index)

	info, _ := doc.String("info")
	require.Equal(t, "testinfo", info)
}

func TestCreate(t *testing.T) {
	db, closeFn := testDB(t)
	defer closeFn()
	ctx := context.TODO()

	path := dstore.Path("test", "key1")
	err := db.Create(ctx, path, dstore.Data([]byte("value1")))
	require.NoError(t, err)

	err = db.Create(ctx, path, dstore.Data([]byte("value1")))
	require.EqualError(t, err, fmt.Sprintf("path already exists %s", path))
}
