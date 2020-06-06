package firestore

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/ds"
	"github.com/keys-pub/keys/tsutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/api/option"
)

const testURL = "firestore://chilltest-3297b"

var ctx = context.TODO()

func testCollection() string {
	return "test-" + time.Now().Format(time.RFC3339Nano)
}

func testFirestore(t *testing.T) *Firestore {
	opts := []option.ClientOption{option.WithCredentialsFile("credentials.json")}
	fs, err := New(testURL, opts...)
	require.NoError(t, err)
	return fs
}

func TestFirestore(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	fs := testFirestore(t)
	testDocumentStore(t, fs)
}

func TestFirestorePath(t *testing.T) {
	fs := testFirestore(t)
	testDocumentStorePath(t, fs)
}

func TestFirestoreListOptions(t *testing.T) {
	fs := testFirestore(t)
	testDocumentStoreListOptions(t, fs)
}

func TestFirestoreMetadata(t *testing.T) {
	// SetContextLogger(NewContextLogger(DebugLevel))
	fs := testFirestore(t)
	testMetadata(t, fs)
}

func TestEmptyIterator(t *testing.T) {
	iter := &docsIterator{}
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()
}

func testDocumentStore(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()
	collection1 := testCollection()
	collection2 := testCollection()

	for i := 10; i <= 30; i = i + 10 {
		p := ds.Path(collection2, fmt.Sprintf("key%d", i))
		err := dst.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}
	for i := 10; i <= 30; i = i + 10 {
		p := ds.Path(collection1, fmt.Sprintf("key%d", i))
		err := dst.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}

	iter, err := dst.Documents(ctx, collection1)
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, ds.Path(collection1, "key10"), doc.Path)
	require.Equal(t, "value10", string(doc.Data))
	iter.Release()

	ok, err := dst.Exists(ctx, ds.Path(collection1, "key10"))
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = dst.Get(ctx, ds.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "value10", string(doc.Data))

	err = dst.Create(ctx, ds.Path(collection1, "key10"), []byte{})
	require.EqualError(t, err, "path already exists "+ds.Path(collection1, "key10"))
	err = dst.Set(ctx, ds.Path(collection1, "key10"), []byte("overwrite"))
	require.NoError(t, err)
	err = dst.Create(ctx, ds.Path(collection1, "key10"), []byte("overwrite"))
	require.EqualError(t, err, "path already exists "+ds.Path(collection1, "key10"))
	doc, err = dst.Get(ctx, ds.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "overwrite", string(doc.Data))

	out, err := dst.GetAll(ctx, []string{ds.Path(collection1, "key10"), ds.Path(collection2, "key20")})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, ds.Path(collection1, "key10"), out[0].Path)
	require.Equal(t, ds.Path(collection2, "key20"), out[1].Path)

	ok, err = dst.Delete(ctx, ds.Path(collection2, "key10"))
	require.True(t, ok)
	require.NoError(t, err)
	ok, err = dst.Delete(ctx, ds.Path(collection2, "key10"))
	require.False(t, ok)
	require.NoError(t, err)

	ok, err = dst.Exists(ctx, ds.Path(collection2, "key10"))
	require.NoError(t, err)
	require.False(t, ok)

	expected := "/" + collection1 + "/key10 overwrite\n/" + collection1 + "/key20 value20\n/" + collection1 + "/key30 value30\n"
	var b bytes.Buffer
	iter, err = dst.Documents(context.TODO(), collection1)
	require.NoError(t, err)
	err = ds.SpewOut(iter, &b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = dst.Documents(context.TODO(), collection1)
	require.NoError(t, err)
	spew, err := ds.Spew(iter)
	require.NoError(t, err)
	require.Equal(t, b.String(), spew.String())
	require.Equal(t, expected, spew.String())
	iter.Release()

	iter, err = dst.Documents(context.TODO(), collection1, ds.Prefix("key1"), ds.NoData())
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, ds.Path(collection1, "key10"), doc.Path)
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
	require.NotEmpty(t, col.Path)
	citer.Release()

	_, err = dst.Collections(ctx, "/foo")
	require.EqualError(t, err, "only root collections supported")
}

func testDocumentStorePath(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()
	collection := testCollection()

	err := dst.Create(ctx, ds.Path(collection, "key1"), []byte("value1"))
	require.NoError(t, err)

	doc, err := dst.Get(ctx, ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)

	ok, err := dst.Exists(ctx, ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = dst.Exists(ctx, ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)
}

func testDocumentStoreListOptions(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()
	collection := testCollection()

	err := dst.Create(ctx, ds.Path(collection, "key1"), []byte("val1"))
	require.NoError(t, err)
	err = dst.Create(ctx, ds.Path(collection, "key2"), []byte("val2"))
	require.NoError(t, err)
	err = dst.Create(ctx, ds.Path(collection, "key3"), []byte("val3"))
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("a"+collection, fmt.Sprintf("e%d", i)), []byte("🤓"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b"+collection, fmt.Sprintf("ea%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b"+collection, fmt.Sprintf("eb%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("b"+collection, fmt.Sprintf("ec%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := dst.Create(ctx, ds.Path("c"+collection, fmt.Sprintf("e%d", i)), []byte("😎"))
		require.NoError(t, err)
	}

	iter, err := dst.Documents(ctx, collection)
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
	require.Equal(t, []string{ds.Path(collection, "key1"), ds.Path(collection, "key2"), ds.Path(collection, "key3")}, paths)
	iter.Release()

	iter, err = dst.Documents(context.TODO(), collection)
	require.NoError(t, err)
	b, err := ds.Spew(iter)
	require.NoError(t, err)
	expected := "/" + collection + "/key1 val1\n" + "/" + collection + "/key2 val2\n" + "/" + collection + "/key3 val3\n"

	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = dst.Documents(ctx, "b"+collection, ds.Prefix("eb"))
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
	require.Equal(t, []string{"/b" + collection + "/eb1", "/b" + collection + "/eb2"}, paths)
}

func testMetadata(t *testing.T, dst ds.DocumentStore) {
	ctx := context.TODO()
	collection := testCollection()

	err := dst.Create(ctx, ds.Path(collection, "key1"), []byte("value1"))
	require.NoError(t, err)

	doc, err := dst.Get(ctx, ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	createTime := tsutil.Millis(doc.CreatedAt)
	require.True(t, createTime > 0)

	err = dst.Set(ctx, ds.Path(collection, "key1"), []byte("value1b"))
	require.NoError(t, err)

	doc, err = dst.Get(ctx, ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, createTime, tsutil.Millis(doc.CreatedAt))
	require.True(t, tsutil.Millis(doc.UpdatedAt) > createTime)
}

func TestSigchains(t *testing.T) {
	clock := tsutil.NewClock()
	fs := testFirestore(t)
	scs := keys.NewSigchainStore(fs)

	kids := []keys.ID{}
	for i := 0; i < 6; i++ {
		key := keys.GenerateEdX25519Key()
		sc := keys.NewSigchain(key.ID())
		st, err := keys.NewSigchainStatement(sc, []byte("test"), key, "", clock.Now())
		require.NoError(t, err)
		err = sc.Add(st)
		require.NoError(t, err)
		err = scs.SaveSigchain(sc)
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

	exists, err := fs.Exists(context.TODO(), ds.Path(collection, "key1"))
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		if _, err := fs.Delete(context.TODO(), ds.Path(collection, "key1")); err != nil {
			log.Fatal(err)
		}
	}

	if err := fs.Create(context.TODO(), ds.Path(collection, "key1"), []byte("value1")); err != nil {
		log.Fatal(err)
	}

	entry, err := fs.Get(context.TODO(), ds.Path(collection, "key1"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", entry.Path)
	fmt.Printf("%s\n", entry.Data)
	// Output:
	// /test/key1
	// value1
}

func TestDeleteAll(t *testing.T) {
	fs := testFirestore(t)
	collection := testCollection()

	err := fs.Set(context.TODO(), ds.Path(collection, "key1"), []byte("val1"))
	require.NoError(t, err)
	err = fs.Set(context.TODO(), ds.Path(collection, "key2"), []byte("val2"))
	require.NoError(t, err)

	err = fs.DeleteAll(context.TODO(), []string{ds.Path(collection, "key1"), ds.Path(collection, "key2"), ds.Path(collection, "key3")})
	require.NoError(t, err)

	doc, err := fs.Get(context.TODO(), ds.Path(collection, "key1"))
	require.NoError(t, err)
	require.Nil(t, doc)
	doc, err = fs.Get(context.TODO(), ds.Path(collection, "key2"))
	require.NoError(t, err)
	require.Nil(t, doc)
}
