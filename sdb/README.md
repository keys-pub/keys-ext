# SDB

This package implements dstore.Documents backed by a leveldb database encrypted with [github.com/minio/sio](https://github.com/minio/sio) (DARE).

**Only values are encrypted.**

```go
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
```
