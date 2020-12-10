package sdb_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/sdb"
	"github.com/keys-pub/keys/dstore"
)

func ExampleNew() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}

	type Message struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	msg := &Message{ID: "id1", Content: "hi"}

	if err := db.Set(context.TODO(), dstore.Path("collection1", "doc1"), dstore.From(msg)); err != nil {
		log.Fatal(err)
	}

	iter, err := db.DocumentIterator(context.TODO(), dstore.Path("collection1"))
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Release()
	for {
		doc, err := iter.Next()
		if err != nil {
			log.Fatal(err)
		}
		if doc == nil {
			break
		}
		var msg Message
		if err := doc.To(&msg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", doc.Path, msg.Content)
	}

	// Output:
	// /collection1/doc1: hi
}

func ExampleDB_OpenAtPath() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_Create() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}

	if err := db.Create(context.TODO(), "/test/1", dstore.Data([]byte{0x01, 0x02, 0x03})); err != nil {
		log.Fatal(err)
	}
}

func ExampleDB_Get() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	if err := db.Set(context.TODO(), dstore.Path("collection1", "doc1"), dstore.Data([]byte("hi"))); err != nil {
		log.Fatal(err)
	}

	doc, err := db.Get(context.TODO(), dstore.Path("collection1", "doc1"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Got %s\n", string(doc.Data()))
	// Output:
	// Got hi
}

func ExampleDB_Set() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	type Message struct {
		ID      string `msgpack:"id"`
		Content string `msgpack:"content"`
	}
	msg := &Message{ID: "id1", Content: "hi"}

	if err := db.Set(context.TODO(), dstore.Path("collection1", "doc1"), dstore.From(msg)); err != nil {
		log.Fatal(err)
	}

	doc, err := db.Get(context.TODO(), dstore.Path("collection1", "doc1"))
	if err != nil {
		log.Fatal(err)
	}
	var out Message
	if err := doc.To(&out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Message: %s\n", out.Content)
	// Output:
	// Message: hi
}

func ExampleDB_Documents() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	if err := db.Set(context.TODO(), dstore.Path("collection1", "doc1"), dstore.Data([]byte("hi"))); err != nil {
		log.Fatal(err)
	}

	docs, err := db.Documents(context.TODO(), dstore.Path("collection1"))
	if err != nil {
		log.Fatal(err)
	}
	for _, doc := range docs {
		fmt.Printf("%s: %s\n", doc.Path, string(doc.Data()))
	}
	// Output:
	// /collection1/doc1: hi
}

func ExampleDB_DocumentIterator() {
	db := sdb.New()
	defer db.Close()

	key := keys.Rand32()
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatal(err)
	}
	path := filepath.Join(dir, "my.sdb")
	if err := db.OpenAtPath(context.TODO(), path, key); err != nil {
		log.Fatal(err)
	}
	// Don't remove db in real life
	defer os.RemoveAll(path)

	type Message struct {
		ID      string `json:"id"`
		Content string `json:"content"`
	}
	msg := &Message{ID: "id1", Content: "hi"}

	if err := db.Set(context.TODO(), dstore.Path("collection1", "doc1"), dstore.From(msg)); err != nil {
		log.Fatal(err)
	}

	iter, err := db.DocumentIterator(context.TODO(), dstore.Path("collection1"))
	if err != nil {
		log.Fatal(err)
	}
	defer iter.Release()
	for {
		doc, err := iter.Next()
		if err != nil {
			log.Fatal(err)
		}
		if doc == nil {
			break
		}
		var msg Message
		if err := doc.To(&msg); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s: %s\n", doc.Path, msg.Content)
	}
	// Output:
	// /collection1/doc1: hi
}
