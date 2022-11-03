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
	tlvh "github.com/justincpresley/ndn-sync/util/tlvhelp"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(source string, seqno uint, oldData bool)
	Get(source string) uint
	String() string
	Len() int
	Total() uint
	Entries() *om.OrderedMap[string, uint]
	ToComponent() enc.Component
	EncodingLength() int
	EncodeInto(buf []byte) int
}

type stateVector struct {
	// WARNING: ideally enc.Name, unable due to not implementing comparable
	entries *om.OrderedMap[string, uint]
}

func NewStateVector() StateVector {
	return stateVector{entries: om.New[string, uint]()}
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
		seqno  uint
		buf    []byte = comp.Val
		pos    uint
		length uint
		temp   uint
		typ    uint
	)
	// verify type
	if comp.Typ != TypeVector {
		return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
	}
	// decode components
	ret = NewStateVector()
	for pos < uint(len(buf)) {
		// entry
		typ, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != TypeEntry {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		_, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		// source
		typ, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != enc.TypeName {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		source, _ = enc.ReadName(enc.NewBufferReader(buf[pos : pos+length]))
		pos += length
		// seqno
		typ, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != TypeEntrySeqno {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = tlvh.ParseUint(buf, pos)
		pos += temp
		seqno, _ = tlvh.ParseUint(buf, pos)
		pos += length
		// add the entry
		ret.Set(source.String(), seqno, true)
	}
	return ret, nil
}

func (sv stateVector) Set(source string, seqno uint, old bool) {
	sv.entries.Set(source, seqno, old)
}

func (sv stateVector) Get(source string) uint {
	if val, present := sv.entries.Get(source); present {
		return val
	} else {
		return 0
	}
}

func (sv stateVector) String() string {
	str := ""
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		str += pair.Key + ":" + strconv.FormatUint(uint64(pair.Value), 10) + " "
	}
	if str != "" {
		return str[:len(str)-1]
	}
	return str
}

func (sv stateVector) Len() int {
	return sv.entries.Len()
}

func (sv stateVector) Total() uint {
	total := uint(0)
	for pair := sv.entries.First(); pair != nil; pair = pair.Next() {
		total += pair.Value
	}
	return total
}

func (sv stateVector) Entries() *om.OrderedMap[string, uint] {
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
		entry += tlvh.GetUintByteSize(uint(t))
		entry += t
		// seqno
		entry += TypeEntrySeqno.EncodingLength()
		entry += tlvh.GetUintByteSize(uint(tlvh.GetUintByteSize(pair.Value)))
		entry += tlvh.GetUintByteSize(pair.Value)
		// entry
		total += TypeEntry.EncodingLength()
		total += tlvh.GetUintByteSize(uint(entry))
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
		entryLen += tlvh.GetUintByteSize(uint(t))
		entryLen += t
		entryLen += TypeEntrySeqno.EncodingLength()
		entryLen += tlvh.GetUintByteSize(uint(tlvh.GetUintByteSize(pair.Value)))
		entryLen += tlvh.GetUintByteSize(pair.Value)
		offset = TypeEntry.EncodingLength() + tlvh.GetUintByteSize(uint(entryLen))
		entryLen = offset
		// source
		entryLen += enc.TypeName.EncodeInto(buf[pos+entryLen:])
		entryLen += tlvh.WriteUint(uint(t), buf, pos+entryLen)
		entryLen += n.EncodeInto(buf[pos+entryLen:])
		// seqno
		entryLen += TypeEntrySeqno.EncodeInto(buf[pos+entryLen:])
		entryLen += tlvh.WriteUint(uint(tlvh.GetUintByteSize(pair.Value)), buf, pos+entryLen)
		entryLen += tlvh.WriteUint(pair.Value, buf, pos+entryLen)
		// entry
		entryLen -= offset
		pos += TypeEntry.EncodeInto(buf[pos:])
		pos += tlvh.WriteUint(uint(entryLen), buf, pos)
		pos += entryLen
	}
	return pos
}
