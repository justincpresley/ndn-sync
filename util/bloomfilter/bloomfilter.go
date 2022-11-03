/*
 This module is a modified version of the original work.
 Original work can be found at:
          https://github.com/reddragon/bloomfilter.go

 No license was provided. However, the changes or modifications
 done are described in changes.md.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.

 Copyright (c) 2013, Gaurav Menghani
 All rights reserved.
*/

package bloomfilter

import (
	"bytes"
	"hash"
	"math"

	bitset "github.com/justincpresley/ndn-sync/util/bitset"
	xxhash3 "github.com/zeebo/xxh3"
)

// The standard bloom-filter, which allows adding of
// elements, and checking for their existence
type Filter struct {
	k      uint           // Number of hash functions
	m      uint           // Size of the bloom-filter
	hashfn hash.Hash64    // The hash function
	idxes  []uint         // Current indexes for data
	bitmap *bitset.BitSet // The bloom-filter bitmap
}

// Returns a new Filter object, if you pass the
// number of Hash Functions to use and the maximum
// size of the bloom-filter
func NewFilter(numHashFuncs uint, bfSize uint) *Filter {
	return &Filter{
		k:      numHashFuncs,
		m:      bfSize,
		hashfn: xxhash3.New(), // TODO: possibly seed the hash
		idxes:  make([]uint, numHashFuncs),
		bitmap: bitset.New(uint32(bfSize)),
	}
}

// Parses bytes into a bloom-filter
func ParseFilter(b []byte) (*Filter, error) {
	k := uint(b[0])
	bs, err := bitset.Parse(b[1:])
	if err != nil {
		return nil, err
	}
	return &Filter{
		k:      k,
		m:      uint(bs.Len()),
		hashfn: xxhash3.New(), // TODO: possibly seed the hash
		idxes:  make([]uint, k),
		bitmap: bs,
	}, nil
}

func (f *Filter) getHash(d []byte) (uint32, uint32) {
	f.hashfn.Reset()
	f.hashfn.Write(d)
	hash64 := f.hashfn.Sum64()
	return uint32(hash64 & ((1 << 32) - 1)),
		uint32(hash64 >> 32)
}

func (f *Filter) setIndexes(d []byte) {
	h1, h2 := f.getHash(d)
	tries := uint(0)
	for i := uint(0); i < f.k; i++ {
	Retry:
		idx := (h1 + uint32(i+tries)*h2) % uint32(f.m)
		for j := i; j != 0; j-- {
			if f.idxes[j-1] == uint(idx) {
				tries++
				goto Retry
			}
		}
		f.idxes[i] = uint(idx)
	}
}

// Adds an element (in byte-array form) to the bloom-filter
func (f *Filter) Add(d []byte) {
	f.setIndexes(d)
	for _, idx := range f.idxes {
		f.bitmap.Set(idx, true)
	}
}

// Checks if an element (in byte-array form) exists in the
// bloom-filter
func (f *Filter) Check(d []byte) bool {
	f.setIndexes(d)
	result := true
	for _, idx := range f.idxes {
		result = result && f.bitmap.Test(idx)
	}
	return result
}

// Turns the bloom-filter into bytes
func (f *Filter) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(int(1 + f.bitmap.WriteSize()))
	err := buf.WriteByte(byte(f.k))
	if err != nil {
		return nil, err
	}
	err = f.bitmap.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}

// Returns the current False Positive Rate of the bloom-filter
func (f *Filter) FalsePositiveRate(n uint) float64 {
	return math.Pow((1 - math.Exp(-float64(f.k*n)/
		float64(f.m))), float64(f.k))
}
