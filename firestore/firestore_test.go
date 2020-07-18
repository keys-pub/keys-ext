package firestore

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys/docs"
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
	return docs.Path("test", time.Now().Format(time.RFC3339Nano))
}

func testCollection() string {
	return docs.Path("test", time.Now().Format(time.RFC3339Nano), "root")
}

func TestDocumentStore(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection1 := testCollection()
	collection2 := testCollection()

	for i := 10; i <= 30; i = i + 10 {
		p := docs.Path(collection2, fmt.Sprintf("key%d", i))
		err := ds.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}
	for i := 10; i <= 30; i = i + 10 {
		p := docs.Path(collection1, fmt.Sprintf("key%d", i))
		err := ds.Create(ctx, p, []byte(fmt.Sprintf("value%d", i)))
		require.NoError(t, err)
	}

	iter, err := ds.DocumentIterator(ctx, collection1)
	require.NoError(t, err)
	doc, err := iter.Next()
	require.NoError(t, err)
	require.Equal(t, docs.Path(collection1, "key10"), doc.Path)
	require.Equal(t, "value10", string(doc.Data))
	iter.Release()

	ok, err := ds.Exists(ctx, docs.Path(collection1, "key10"))
	require.NoError(t, err)
	require.True(t, ok)
	doc, err = ds.Get(ctx, docs.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "value10", string(doc.Data))

	err = ds.Create(ctx, docs.Path(collection1, "key10"), []byte{})
	require.EqualError(t, err, "path already exists "+docs.Path(collection1, "key10"))
	err = ds.Set(ctx, docs.Path(collection1, "key10"), []byte("overwrite"))
	require.NoError(t, err)
	err = ds.Create(ctx, docs.Path(collection1, "key10"), []byte("overwrite"))
	require.EqualError(t, err, "path already exists "+docs.Path(collection1, "key10"))
	doc, err = ds.Get(ctx, docs.Path(collection1, "key10"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, "overwrite", string(doc.Data))

	out, err := ds.GetAll(ctx, []string{docs.Path(collection1, "key10"), docs.Path(collection2, "key20")})
	require.NoError(t, err)
	require.Equal(t, 2, len(out))
	require.Equal(t, docs.Path(collection1, "key10"), out[0].Path)
	require.Equal(t, docs.Path(collection2, "key20"), out[1].Path)

	ok, err = ds.Delete(ctx, docs.Path(collection2, "key10"))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = ds.Delete(ctx, docs.Path(collection2, "key10"))
	require.NoError(t, err)
	require.False(t, ok)

	ok, err = ds.Exists(ctx, docs.Path(collection2, "key10"))
	require.NoError(t, err)
	require.False(t, ok)

	expected := collection1 + "/key10 overwrite\n" + collection1 + "/key20 value20\n" + collection1 + "/key30 value30\n"
	var b bytes.Buffer
	iter, err = ds.DocumentIterator(context.TODO(), collection1)
	require.NoError(t, err)
	err = docs.SpewOut(iter, &b)
	require.NoError(t, err)
	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection1)
	require.NoError(t, err)
	spew, err := docs.Spew(iter)
	require.NoError(t, err)
	require.Equal(t, b.String(), spew.String())
	require.Equal(t, expected, spew.String())
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection1, docs.Prefix("key1"), docs.NoData())
	require.NoError(t, err)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Equal(t, docs.Path(collection1, "key10"), doc.Path)
	doc, err = iter.Next()
	require.NoError(t, err)
	require.Nil(t, doc)
	iter.Release()

	err = ds.Create(ctx, "", []byte{})
	require.EqualError(t, err, "invalid path /")
	err = ds.Set(ctx, "", []byte{})
	require.EqualError(t, err, "invalid path /")

	cols, err := ds.Collections(ctx, "")
	require.NoError(t, err)
	expectedCols := []*docs.Collection{
		&docs.Collection{Path: "/msgs"},
		&docs.Collection{Path: "/sigchain"},
		&docs.Collection{Path: "/test"},
		&docs.Collection{Path: "/vaults-rm"},
	}
	require.Equal(t, expectedCols, cols)

	_, err = ds.Collections(ctx, "/foo")
	require.EqualError(t, err, "only root collections supported")
}

func TestDocumentStorePath(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, docs.Path(collection, "key1"), []byte("value1"))
	require.NoError(t, err)

	doc, err := ds.Get(ctx, docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, []byte("value1"), doc.Data)

	ok, err := ds.Exists(ctx, docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = ds.Exists(ctx, docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.True(t, ok)

	err = ds.Create(ctx, docs.Path(collection, "key2", "col2", "key3"), []byte("value3"))
	require.NoError(t, err)

	doc, err = ds.Get(ctx, docs.Path(collection, "key2", "col2", "key3"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, []byte("value3"), doc.Data)

	// citer, err := ds.Collections(ctx, "")
	// require.NoError(t, err)
	// cols, err := docs.CollectionsFromIterator(citer)
	// require.NoError(t, err)
	// require.Equal(t, "/test", cols[0].Path)
}

func TestDocumentStoreListOptions(t *testing.T) {
	ds := testFirestore(t)
	ctx := context.TODO()
	collection := testCollection()

	err := ds.Create(ctx, docs.Path(collection, "key1"), []byte("val1"))
	require.NoError(t, err)
	err = ds.Create(ctx, docs.Path(collection, "key2"), []byte("val2"))
	require.NoError(t, err)
	err = ds.Create(ctx, docs.Path(collection, "key3"), []byte("val3"))
	require.NoError(t, err)

	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, docs.Path(collection+"a", fmt.Sprintf("e%d", i)), []byte("🤓"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, docs.Path(collection+"b", fmt.Sprintf("ea%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, docs.Path(collection+"b", fmt.Sprintf("eb%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, docs.Path(collection+"b", fmt.Sprintf("ec%d", i)), []byte("😎"))
		require.NoError(t, err)
	}
	for i := 1; i < 3; i++ {
		err := ds.Create(ctx, docs.Path(collection+"c", fmt.Sprintf("e%d", i)), []byte("😎"))
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
	require.Equal(t, []string{docs.Path(collection, "key1"), docs.Path(collection, "key2"), docs.Path(collection, "key3")}, paths)
	iter.Release()

	iter, err = ds.DocumentIterator(context.TODO(), collection)
	require.NoError(t, err)
	b, err := docs.Spew(iter)
	require.NoError(t, err)
	expected := collection + "/key1 val1\n" + collection + "/key2 val2\n" + collection + "/key3 val3\n"

	require.Equal(t, expected, b.String())
	iter.Release()

	iter, err = ds.DocumentIterator(ctx, docs.Path(collection+"b"), docs.Prefix("eb"))
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

	err := ds.Create(ctx, docs.Path(collection, "key1"), []byte("value1"))
	require.NoError(t, err)

	doc, err := ds.Get(ctx, docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	createTime := tsutil.Millis(doc.CreatedAt)
	require.True(t, createTime > 0)

	err = ds.Set(ctx, docs.Path(collection, "key1"), []byte("value1b"))
	require.NoError(t, err)

	doc, err = ds.Get(ctx, docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.Equal(t, createTime, tsutil.Millis(doc.CreatedAt))
	require.True(t, tsutil.Millis(doc.UpdatedAt) > createTime)
}

func TestSigchains(t *testing.T) {
	clock := tsutil.NewTestClock()
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

	exists, err := fs.Exists(context.TODO(), docs.Path(collection, "key1"))
	if err != nil {
		log.Fatal(err)
	}
	if exists {
		if _, err := fs.Delete(context.TODO(), docs.Path(collection, "key1")); err != nil {
			log.Fatal(err)
		}
	}

	if err := fs.Create(context.TODO(), docs.Path(collection, "key1"), []byte("value1")); err != nil {
		log.Fatal(err)
	}

	entry, err := fs.Get(context.TODO(), docs.Path(collection, "key1"))
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

	err := fs.Set(context.TODO(), docs.Path(collection, "key1"), []byte("val1"))
	require.NoError(t, err)
	err = fs.Set(context.TODO(), docs.Path(collection, "key2"), []byte("val2"))
	require.NoError(t, err)

	err = fs.DeleteAll(context.TODO(), []string{docs.Path(collection, "key1"), docs.Path(collection, "key2"), docs.Path(collection, "key3")})
	require.NoError(t, err)

	doc, err := fs.Get(context.TODO(), docs.Path(collection, "key1"))
	require.NoError(t, err)
	require.Nil(t, doc)
	doc, err = fs.Get(context.TODO(), docs.Path(collection, "key2"))
	require.NoError(t, err)
	require.Nil(t, doc)
}

// func TestDeleteCollections(t *testing.T) {
// 	fs := testFirestore(t)
// 	iter, err := fs.Collections(context.TODO(), "")
// 	require.NoError(t, err)
// 	cols, err := docs.CollectionsFromIterator(iter)
// 	require.NoError(t, err)
// 	iter.Release()

// 	for _, col := range cols {
// 		if strings.HasPrefix(col.Path, "/test") {
// 			t.Logf("Col: %s", col)
// 			diter, err := fs.Documents(context.TODO(), col.Path)
// 			require.NoError(t, err)
// 			docs, err := docs.DocumentsFromIterator(diter)
// 			require.NoError(t, err)
// 			diter.Release()
// 			for _, doc := range docs {
// 				t.Logf("Delete: %s", doc.Path)
// 				_, err = fs.Delete(context.TODO(), doc.Path)
// 				require.NoError(t, err)
// 			}
// 		}
// 	}
// }
