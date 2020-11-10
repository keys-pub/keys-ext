package firestore

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/dstore"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const testURL = "firestore://chilltest-3297b"

func testFirestore(t *testing.T) *Firestore {
	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
	fs, err := New(testURL, opts...)
	require.NoError(t, err)
	return fs
}

func TestEmptyIterator(t *testing.T) {
	iter := &docsIterator{}
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()
}

func testPath() string {
	return dstore.Path("test", time.Now().Format(time.RFC3339Nano))
}

func testCollection() string {
	return dstore.Path("test", time.Now().Format(time.RFC3339Nano), "root")
}

func TestDocumentStore(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection1 := testCollection()
	collection2 := testCollection()

	for i := 10; i <= 30; i = i + 10 {
		p := dstore.Path(collection2, fmt.Sprintf("key%d", i))
		err := ds.Create(ctx, p, dstore.Data([]byte(fmt.Sprintf("value%d", i))))
		require.NoError(t, err)
	}
	for i := 10; i <= 30; i = i + 10 {
		p := dstore.Path(collection1, fmt.Sprintf("key%d", i))
		err := ds.Create(ctx, p, dstore.Data([]byte(fmt.Sprintf("value%d", i))))
		require.NoError(t, err)
	}

	iter, err := ds.DocumentIterator(ctx, collection1)
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, dstore.Path(collection1, "key10"), doc.Path)
	require.Equal(t, "value10", string(doc.Data()))
	iter.Release()

	ok, err := ds.Exists(ctx, dstore.Path(collection1, "key10"))
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = ds.Get(ctx, dstore.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "value10", string(doc.Data()))

	err = ds.Create(ctx, dstore.Path(collection1, "key10"), dstore.Data([]byte{}))
	require.EqualError(t, err, "path already exists "+dstore.Path(collection1, "key10"))
	err = ds.Set(ctx, dstore.Path(collection1, "key10"), dstore.Data([]byte("overwrite")))
	require.NoError(t, err)
	err = ds.Create(ctx, dstore.Path(collection1, "key10"), dstore.Data([]byte("overwrite")))
	require.EqualError(t, err, "path already exists "+dstore.Path(collection1, "key10"))
	doc, err = ds.Get(ctx, dstore.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "overwrite", string(doc.Data()))

	out, err := ds.GetAll(ctx, []string{dstore.Path(collection1, "key10"), dstore.Path(collection2, "key20")})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, dstore.Path(collection1, "key10"), out[0].Path)
	require.Equal(t, dstore.Path(collection2, "key20"), out[1].Path)

	ok, err = ds.Delete(ctx, dstore.Path(collection2, "key10"))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = ds.Delete(ctx, dstore.Path(collection2, "key10"))
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = ds.Exists(ctx, dstore.Path(collection2, "key10"))
	require.NoError(t, err)
	require.False(t, ok)

	expected := collection1 + "/key10 overwrite\n" + collection1 + "/key20 value20\n" + collection1 + "/key30 value30\n"
	var b bytes.Buffer
	iter, err = ds.DocumentIterator(context.TODO(), collection1)
	require.NoError(t, err)
	err = dstore.SpewOut(iter, &b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection1)
	require.NoError(t, err)
	spew, err := dstore.Spew(iter)
	require.NoError(t, err)
	require.Equal(t, b.String(), spew.String())
	require.Equal(t, expected, spew.String())
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection1, dstore.Prefix("key1"), dstore.NoData())
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, dstore.Path(collection1, "key10"), doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()

	err = ds.Create(ctx, "", dstore.Data([]byte{}))
	require.EqualError(t, err, "invalid path /")
	err = ds.Set(ctx, "", dstore.Data([]byte{}))
	require.EqualError(t, err, "invalid path /")

	cols, err := ds.Collections(ctx, "")
	require.NoError(t, err)
	expectedCols := []*dstore.Collection{
		&dstore.Collection{Path: "/channels"},
		&dstore.Collection{Path: "/inbox"},
		&dstore.Collection{Path: "/rkl"},
		&dstore.Collection{Path: "/sigchain"},
		&dstore.Collection{Path: "/test"},
		&dstore.Collection{Path: "/vaults"},
		&dstore.Collection{Path: "/vaults-rm"},
	}
	require.Equal(t, expectedCols, cols)

	_, err = ds.Collections(ctx, "/foo")
	require.EqualError(t, err, "only root collections supported")
}

func TestDocumentStorePath(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, dstore.Path(collection, "key1"), dstore.Data([]byte("value1")))
	require.NoError(t, err)

	doc, err := ds.Get(ctx, dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, []byte("value1"), doc.Data())

	ok, err := ds.Exists(ctx, dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = ds.Exists(ctx, dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)

	err = ds.Create(ctx, dstore.Path(collection, "key2", "col2", "key3"), dstore.Data([]byte("value3")))
	require.NoError(t, err)

	doc, err = ds.Get(ctx, dstore.Path(collection, "key2", "col2", "key3"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, []byte("value3"), doc.Data())

	// citer, err := ds.Collections(ctx, "")
	// require.NoError(t, err)
	// cols, err := dstore.CollectionsFromIterator(citer)
	// require.NoError(t, err)
	// require.Equal(t, "/test", cols[0].Path)
}

func TestDocumentStoreListOptions(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, dstore.Path(collection, "key1"), dstore.Data([]byte("val1")))
	require.NoError(t, err)
	err = ds.Create(ctx, dstore.Path(collection, "key2"), dstore.Data([]byte("val2")))
	require.NoError(t, err)
	err = ds.Create(ctx, dstore.Path(collection, "key3"), dstore.Data([]byte("val3")))
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, dstore.Path(collection+"a", fmt.Sprintf("e%d", i)), dstore.Data([]byte("🤓")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, dstore.Path(collection+"b", fmt.Sprintf("ea%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, dstore.Path(collection+"b", fmt.Sprintf("eb%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, dstore.Path(collection+"b", fmt.Sprintf("ec%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, dstore.Path(collection+"c", fmt.Sprintf("e%d", i)), dstore.Data([]byte("😎")))
		require.NoError(t, err)
	}

	iter, err := ds.DocumentIterator(ctx, collection)
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
	require.Equal(t, []string{dstore.Path(collection, "key1"), dstore.Path(collection, "key2"), dstore.Path(collection, "key3")}, paths)
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection)
	require.NoError(t, err)
	b, err := dstore.Spew(iter)
	require.NoError(t, err)
	expected := collection + "/key1 val1\n" + collection + "/key2 val2\n" + collection + "/key3 val3\n"

	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = ds.DocumentIterator(ctx, dstore.Path(collection+"b"), dstore.Prefix("eb"))
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
	require.Equal(t, []string{collection + "b/eb1", collection + "b/eb2"}, paths)
}

func TestMetadata(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, dstore.Path(collection, "key1"), dstore.Data([]byte("value1")))
	require.NoError(t, err)

	doc, err := ds.Get(ctx, dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	createTime := tsutil.Millis(doc.CreatedAt)
	require.True(t, createTime > 0)

	err = ds.Set(ctx, dstore.Path(collection, "key1"), dstore.Data([]byte("value1b")))
	require.NoError(t, err)

	doc, err = ds.Get(ctx, dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, createTime, tsutil.Millis(doc.CreatedAt))
	require.True(t, tsutil.Millis(doc.UpdatedAt) > createTime)
}

func TestSigchains(t *testing.T) {
	clock := tsutil.NewTestClock()
	fs := testFirestore(t)
	scs := keys.NewSigchains(fs)

	kids := []keys.ID{}
	for i := 0; i < 6; i++ {
		key := keys.GenerateEdX25519Key()
		sc := keys.NewSigchain(key.ID())
		st, err := keys.NewSigchainStatement(sc, []byte("test"), key, "", clock.Now())
		require.NoError(t, err)
		err = sc.Add(st)
		require.NoError(t, err)
		err = scs.Save(sc)
		require.NoError(t, err)
		kids = append(kids, key.ID())
	}

	sc, err := scs.Sigchain(kids[0])
	require.NoError(t, err)
	require.Equal(t, kids[0], sc.KID())
}

func ExampleNew() {
	url := "firestore://chilltest-3297b"
	collection := "test"

	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
	fs, err := New(url, opts...)
	if err != nil {
		log.Fatal(err)
	}

	exists, err := fs.Exists(context.TODO(), dstore.Path(collection, "key1"))
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		if _, err := fs.Delete(context.TODO(), dstore.Path(collection, "key1")); err != nil {
			log.Fatal(err)
		}
	}

	if err := fs.Create(context.TODO(), dstore.Path(collection, "key1"), dstore.Data([]byte("value1"))); err != nil {
		log.Fatal(err)
	}

	entry, err := fs.Get(context.TODO(), dstore.Path(collection, "key1"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", entry.Path)
	fmt.Printf("%s\n", entry.Data())
	// Output:
	// /test/key1
	// value1
}

func TestDeleteAll(t *testing.T) {
	fs := testFirestore(t)
	collection := testCollection()

	err := fs.Set(context.TODO(), dstore.Path(collection, "key1"), dstore.Data([]byte("val1")))
	require.NoError(t, err)
	err = fs.Set(context.TODO(), dstore.Path(collection, "key2"), dstore.Data([]byte("val2")))
	require.NoError(t, err)

	err = fs.DeleteAll(context.TODO(), []string{dstore.Path(collection, "key1"), dstore.Path(collection, "key2"), dstore.Path(collection, "key3")})
	require.NoError(t, err)

	doc, err := fs.Get(context.TODO(), dstore.Path(collection, "key1"))
	require.NoError(t, err)
	require.Nil(t, doc)
	doc, err = fs.Get(context.TODO(), dstore.Path(collection, "key2"))
	require.NoError(t, err)
	require.Nil(t, doc)
}

func TestUpdate(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, dstore.Path(collection, "key1"), dstore.Data([]byte("val1")))
	require.NoError(t, err)

	err = ds.Set(ctx, dstore.Path(collection, "key1"), map[string]interface{}{"index": 1, "info": "testinfo"}, dstore.MergeAll())
	require.NoError(t, err)

	time.Sleep(time.Second)

	doc, err := ds.Get(ctx, dstore.Path(collection, "key1"))
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
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	path := dstore.Path(collection, "key1")
	err := ds.Create(ctx, path, dstore.Data([]byte("value1")))
	require.NoError(t, err)

	err = ds.Create(ctx, path, dstore.Data([]byte("value1")))
	require.EqualError(t, err, fmt.Sprintf("path already exists %s", path))
}
