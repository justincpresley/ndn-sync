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
	"errors"
	"strconv"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(source string, seqno uint64, oldData bool)
	Get(source string) uint64
	String() string
	Len() int
	Total() uint64
	Entries() *om.OrderedMap[string, uint64]
	ToComponent() enc.Component
	EncodingLength() int
	EncodeInto(buf []byte) int
}

type stateVector struct {
	// WARNING: ideally enc.Name, unable due to not implementing comparable
	entries *om.OrderedMap[string, uint64]
}

func NewStateVector() StateVector {
	return stateVector{entries: om.New[string, uint64]()}
}

func ParseStateVector(comp enc.Component) (ret StateVector, err error) {
	defer func() {
		if recover() != nil {
			ret = NewStateVector()
			err = errors.New("encoding.ParseStatevector: buffer length invalid")
		}
	}()
	var (
		source enc.Name
		seqno  enc.Nat
		length enc.TLNum
		typ    enc.TLNum
		buf    []byte = comp.Val
		pos    int
		temp   int
	)
	// verify type
	if comp.Typ != TypeVector {
		return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
	}
	// decode components
	ret = NewStateVector()
	for pos < len(buf) {
		// entry
		typ, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		if typ != TypeEntry {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		_, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		// source
		typ, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		if typ != enc.TypeName {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		source, _ = enc.ReadName(enc.NewBufferReader(buf[pos : pos+int(length)]))
		pos += int(length)
		// seqno
		typ, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		if typ != TypeEntrySeqno {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		seqno, _ = enc.ParseNat(buf[pos : pos+int(length)])
		pos += int(length)
		// add the entry
		ret.Set(source.String(), uint64(seqno), true)
	}
	return ret, nil
}

func (sv stateVector) Set(source string, seqno uint64, old bool) {
	sv.entries.Set(source, seqno, old)
}

func (sv stateVector) Get(source string) uint64 {
	if val, present := sv.entries.Get(source); present {
		return val
	} else {
		return 0
	}
}

func (sv stateVector) String() string {
	str := ""
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		str += pair.Key + ":" + strconv.FormatUint(pair.Value, 10) + " "
	}
	if str != "" {
		return str[:len(str)-1]
	}
	return str
}

func (sv stateVector) Len() int {
	return sv.entries.Len()
}

func (sv stateVector) Total() uint64 {
	total := uint64(0)
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		total += pair.Value
	}
	return total
}

func (sv stateVector) Entries() *om.OrderedMap[string, uint64] {
	return sv.entries
}

func (sv stateVector) ToComponent() enc.Component {
	length := sv.EncodingLength()
	comp := enc.Component{
		Typ: TypeVector,
		Val: make([]byte, length),
	}
	sv.EncodeInto(comp.Val)
	return comp
}

func (sv stateVector) EncodingLength() int {
	var (
		entry int
		total int
		t     int
		n     enc.Name
	)
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		n, _ = enc.NameFromStr(pair.Key)
		t = n.EncodingLength()
		// source
		entry = enc.TypeName.EncodingLength()
		entry += enc.TLNum(t).EncodingLength()
		entry += t
		// seqno
		entry += TypeEntrySeqno.EncodingLength()
		entry += enc.TLNum(enc.Nat(pair.Value).EncodingLength()).EncodingLength()
		entry += enc.Nat(pair.Value).EncodingLength()
		// entry
		total += TypeEntry.EncodingLength()
		total += enc.TLNum(entry).EncodingLength()
		total += entry
	}
	return total
}

func (sv stateVector) EncodeInto(buf []byte) int {
	var (
		entryLen int
		offset   int
		pos      int
		n        enc.Name
		t        int
	)
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		n, _ = enc.NameFromStr(pair.Key)
		t = n.EncodingLength()
		// entry length
		entryLen = enc.TypeName.EncodingLength()
		entryLen += enc.TLNum(t).EncodingLength()
		entryLen += t
		entryLen += TypeEntrySeqno.EncodingLength()
		entryLen += enc.TLNum(enc.Nat(pair.Value).EncodingLength()).EncodingLength()
		entryLen += enc.Nat(pair.Value).EncodingLength()
		offset = TypeEntry.EncodingLength() + enc.TLNum(entryLen).EncodingLength()
		entryLen = offset
		// source
		entryLen += enc.TypeName.EncodeInto(buf[pos+entryLen:])
		entryLen += enc.TLNum(t).EncodeInto(buf[pos+entryLen:])
		entryLen += n.EncodeInto(buf[pos+entryLen:])
		// seqno
		entryLen += TypeEntrySeqno.EncodeInto(buf[pos+entryLen:])
		entryLen += enc.TLNum(enc.Nat(pair.Value).EncodingLength()).EncodeInto(buf[pos+entryLen:])
		entryLen += enc.Nat(pair.Value).EncodeInto(buf[pos+entryLen:])
		// entry
		entryLen -= offset
		pos += TypeEntry.EncodeInto(buf[pos:])
		pos += enc.TLNum(entryLen).EncodeInto(buf[pos:])
		pos += entryLen
	}
	return pos
}
