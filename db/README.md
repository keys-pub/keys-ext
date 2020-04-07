# DB

This package provides a leveldb database encrypted with [github.com/minio/sio](https://github.com/minio/sio) (DARE).

**Only values are encrypted.**

```go
db := NewDB()
defer db.Close()

key := keys.RandKey()
if err := db.OpenAtPath(context.TODO(), "my.db", key); err != nil {
    log.Fatal(err)
}

if err := db.Set(context.TODO(), "/collection1/doc1", []byte("hi")); err != nil {
    log.Fatal(err)
}

doc, err  := db.Get(context.TODO(), "/collection1/doc1")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("%s: %s\n", doc.Path, string(doc.Data))

iter, err := db.Documents(context.TODO(), keys.Path("collection1"), nil)
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
```
