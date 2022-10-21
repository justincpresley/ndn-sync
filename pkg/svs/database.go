/*
 Copyright (C) 2022-2030, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 ndn-sync is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 ndn-sync is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. If absent, it can be found within the
 GitHub repository:
          https://github.com/justincpresley/ndn-sync
*/

package svs

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"

	bolt "go.etcd.io/bbolt"
)

type Database interface {
	Get(key []byte) (val []byte)
	Set(key []byte, value []byte) error
	Remove(key []byte) error
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

func (fs BoltDB) Set(key []byte, value []byte) error {
	return fs.handle.Update(func(tx *bolt.Tx) error {
		buc := tx.Bucket(fs.bucket)
		return buc.Put(key, value)
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
	if _, err := os.Stat(dir); err != nil {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolvePath(path string) string {
	usr, _ := user.Current()
	if path == "~" {
		path = usr.HomeDir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(usr.HomeDir, path[2:])
	} else if strings.HasPrefix(path, "./") {
		path, _ = filepath.Abs(path)
	}
	return path
}
