package db

import (
	"bytes"

	"github.com/keys-pub/keys"
	"github.com/minio/sio"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	ldbutil "github.com/syndtr/goleveldb/leveldb/util"
)

type sdb struct {
	db  *leveldb.DB
	key keys.SecretKey
	cfg sio.Config
}

type siter struct {
	iterator.Iterator
	db *sdb
}

func (i *siter) Value() []byte {
	b := i.Iterator.Value()
	if b == nil {
		return nil
	}
	decrypted, err := i.db.decrypt(b)
	if err != nil {
		return nil
	}
	return decrypted
}

func newSDB(db *leveldb.DB, key keys.SecretKey) *sdb {
	cfg := sio.Config{
		Key:        key[:],
		MinVersion: sio.Version20,
		MaxVersion: sio.Version20,
	}
	return &sdb{db: db, key: key, cfg: cfg}
}

func (d *sdb) Close() error {
	return d.db.Close()
}

func (d *sdb) Has(path string) (bool, error) {
	return d.db.Has([]byte(path), nil)
}

func (d *sdb) Delete(path string) error {
	return d.db.Delete([]byte(path), nil)
}

func (d *sdb) NewIterator(prefix string) iterator.Iterator {
	iter := d.db.NewIterator(ldbutil.BytesPrefix([]byte(prefix)), nil)
	return &siter{iter, d}
}

func (d *sdb) Put(path string, b []byte) error {
	encrypted, err := d.encrypt(b)
	if err != nil {
		return err
	}
	if err := d.db.Put([]byte(path), encrypted, nil); err != nil {
		return err
	}
	return nil
}

func (d *sdb) Get(path string) ([]byte, error) {
	b, err := d.db.Get([]byte(path), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	decrypted, err := d.decrypt(b)
	if err != nil {
		return nil, err
	}
	return decrypted, nil
}

func (d *sdb) encrypt(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	_, err := sio.Encrypt(&buf, bytes.NewReader(b), d.cfg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (d *sdb) decrypt(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	_, err := sio.Decrypt(&buf, bytes.NewReader(b), d.cfg)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// func (d *DB) encrypt(b []byte) ([]byte, error) {
// 	return keys.SecretBoxSeal(b, d.key), nil
// }

// func (d *DB) decrypt(b []byte) ([]byte, error) {
// 	return keys.SecretBoxOpen(b, d.key)
// }
