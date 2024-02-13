package svs

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	bolt "go.etcd.io/bbolt"
)

type Database interface {
	Get([]byte) []byte
	Set([]byte, []byte) error
	Remove([]byte) error
	Close()
}

type BoltDB struct {
	handle *bolt.DB
	bucket []byte
}

func NewBoltDB(path string, bucket []byte) (BoltDB, error) {
	var (
		err error
		db  *bolt.DB
	)
	path = resolvePath(path)
	err = ensureDirectory(path)
	if err != nil {
		return BoltDB{nil, nil}, err
	}
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		return BoltDB{nil, nil}, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucket)
		return err
	})
	if err != nil {
		return BoltDB{nil, nil}, err
	}
	return BoltDB{handle: db, bucket: bucket}, nil
}

func (fs BoltDB) Get(key []byte) (val []byte) {
	fs.handle.View(func(tx *bolt.Tx) error {
		buc := tx.Bucket(fs.bucket)
		val = buc.Get(key)
		return nil
	})
	return val
}

func (fs BoltDB) Set(key []byte, val []byte) error {
	return fs.handle.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket(fs.bucket)
		return buc.Put(key, val)
	})
}

func (fs BoltDB) Remove(key []byte) error {
	return fs.handle.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket(fs.bucket)
		return buc.Delete(key)
	})
}

func (fs BoltDB) Close() {
	fs.handle.Close()
}

func ensureDirectory(path string) error {
	dir := filepath.Dir(path)
	return os.MkdirAll(dir, os.ModePerm)
}

func resolvePath(path string) string {
	usr, _ := user.Current()
	switch {
	case path == "~":
		path = usr.HomeDir
	case strings.HasPrefix(path, "~/"):
		path = filepath.Join(usr.HomeDir, path[2:])
	case strings.HasPrefix(path, "./"):
		path, _ = filepath.Abs(path)
	}
	return path
}
