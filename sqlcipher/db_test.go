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

type Message struct {
	Text      string  `json:"text"`
	Sender    keys.ID `json:"sender"`
	Recipient keys.ID `json:"recipient"`
	Timestamp float64 `json:"ts"`
}

func testMessage(text string, sender keys.ID, recipient keys.ID, ts int64) *Message {
	return &Message{Text: text, Sender: sender, Recipient: recipient, Timestamp: float64(ts)}
}

func TestDB(t *testing.T) {
	// sqlcipher.SetLogger(sqlcipher.NewLogger(sqlcipher.DebugLevel))

	db, closeFn := testDB(t)
	defer closeFn()

	require.True(t, db.IsOpen())

	ctx := context.TODO()
	alice := keys.NewEdX25519KeyFromSeed(testSeed(0x01))
	bob := keys.NewEdX25519KeyFromSeed(testSeed(0x02))

	for i := 0; i < 3; i++ {
		p := dstore.Path("channel0", fmt.Sprintf("id%d", i))
		msg := testMessage(fmt.Sprintf("channel0 test message %d", i), alice.ID(), bob.ID(), tsutil.NowMillis())
		err := db.Create(ctx, p, dstore.From(msg))
		require.NoError(t, err)
	}
	for i := 3; i < 6; i++ {
		p := dstore.Path("channel0", fmt.Sprintf("id%d", i))
		msg := testMessage(fmt.Sprintf("channel0 test message %d", i), bob.ID(), alice.ID(), tsutil.NowMillis())
		err := db.Create(ctx, p, dstore.From(msg))
		require.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		p := dstore.Path("channel1", fmt.Sprintf("id%d", i))
		msg := testMessage(fmt.Sprintf("channel1 test message %d", i), alice.ID(), bob.ID(), tsutil.NowMillis())
		err := db.Create(ctx, p, dstore.From(msg))
		require.NoError(t, err)
	}

	var out Message

	iter, err := db.DocumentIterator(ctx, "channel0")
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	iter.Release()
	require.NotNil(t, doc)
	require.Equal(t, "/channel0/id0", doc.Path)
	err = doc.To(&out)
	require.NoError(t, err)
	require.Equal(t, "channel0 test message 0", out.Text)

	docs, err := db.Documents(ctx, "channel0")
	require.NoError(t, err)
	require.Equal(t, 6, len(docs))
	require.Equal(t, "/channel0/id0", docs[0].Path)
	err = doc.To(&out)
	require.NoError(t, err)
	require.Equal(t, "channel0 test message 0", out.Text)

	ok, err := db.Exists(ctx, "/channel0/id0")
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = db.Get(ctx, "/channel0/id0")
	require.NoError(t, err)
	require.NotNil(t, doc)
	err = doc.To(&out)
	require.NoError(t, err)
	require.Equal(t, "channel0 test message 0", out.Text)

	empty := map[string]interface{}{}
	err = db.Create(ctx, "/channel0/id0", empty)
	require.EqualError(t, err, "path already exists /channel0/id0")

	overwrite := testMessage("channel0 test message overwrite", alice.ID(), bob.ID(), tsutil.NowMillis())
	err = db.Set(ctx, "/channel0/id0", dstore.From(overwrite))
	require.NoError(t, err)

	err = db.Create(ctx, "/channel0/id0", empty)
	require.EqualError(t, err, "path already exists /channel0/id0")

	doc, err = db.Get(ctx, "/channel0/id0")
	require.NoError(t, err)
	require.NotNil(t, doc)
	err = doc.To(&out)
	require.NoError(t, err)
	require.Equal(t, &out, overwrite)

	docs, err = db.GetAll(ctx, []string{"/channel0/id0", "/channel0/id1"})
	require.NoError(t, err)
	require.Equal(t, 2, len(docs))
	require.Equal(t, "/channel0/id0", docs[0].Path)
	require.Equal(t, "/channel0/id1", docs[1].Path)

	ok, err = db.Delete(ctx, "/channel1/id0")
	require.True(t, ok)
	require.NoError(t, err)
	ok, err = db.Delete(ctx, "/channel1/id0")
	require.False(t, ok)
	require.NoError(t, err)

	ok, err = db.Exists(ctx, "/channel1/id0")
	require.NoError(t, err)
	require.False(t, ok)

	iter, err = db.DocumentIterator(context.TODO(), "channel0", dstore.Prefix("id0"), dstore.NoData())
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "/channel0/id0", doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()

	err = db.Create(ctx, "", empty)
	require.EqualError(t, err, "invalid path /")
	err = db.Set(ctx, "", empty)
	require.EqualError(t, err, "invalid path /")

	cols, err := db.Collections(ctx, "")
	require.NoError(t, err)
	require.Equal(t, 2, len(cols))
	require.Equal(t, "/channel0", cols[0].Path)
	require.Equal(t, "/channel1", cols[1].Path)

	_, err = db.Collections(ctx, "/channel0")
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
	require.Equal(t, 1, len(cols))
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

func testSeed(b byte) *[32]byte {
	return keys.Bytes32(bytes.Repeat([]byte{b}, 32))
}
