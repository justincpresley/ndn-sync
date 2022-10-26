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

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(source string, seqno uint)
	Get(source string) uint
	String() string
	Len() int
	Entries() map[string]uint
	ToComponent() enc.Component
	EncodingLength() int
	EncodeInto(buf []byte) int
}

type stateVector struct {
	// WARNING: ideally enc.Name, unable due to not implementing comparable
	entries map[string]uint
}

func NewStateVector() StateVector {
	return stateVector{entries: make(map[string]uint)}
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
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != TypeEntry {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		_, temp = parse_uint(buf, pos)
		pos += temp
		// source
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != enc.TypeName {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = parse_uint(buf, pos)
		pos += temp
		source, _ = enc.ReadName(enc.NewBufferReader(buf[pos : pos+length]))
		pos += length
		// seqno
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != TypeEntrySeqno {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = parse_uint(buf, pos)
		pos += temp
		seqno, _ = parse_uint(buf, pos)
		pos += length
		// add the entry
		ret.Set(source.String(), seqno)
	}
	return ret, nil
}

func (sv stateVector) Set(source string, seqno uint) {
	sv.entries[source] = seqno
}

func (sv stateVector) Entries() map[string]uint {
	return sv.entries
}

func (sv stateVector) Get(source string) uint {
	return sv.entries[source]
}

func (sv stateVector) String() string {
	str := ""
	for key, ele := range sv.entries {
		str += key + ":" + strconv.FormatUint(uint64(ele), 10) + " "
	}
	if str != "" {
		return str[:len(str)-1]
	}
	return str
}

func (sv stateVector) Len() int {
	return len(sv.entries)
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
	for key, ele := range sv.entries {
		n, _ = enc.NameFromStr(key)
		t = n.EncodingLength()
		// source
		entry = enc.TypeName.EncodingLength()
		entry += get_uint_byte_size(uint(t))
		entry += t
		// seqno
		entry += TypeEntrySeqno.EncodingLength()
		entry += get_uint_byte_size(uint(get_uint_byte_size(ele)))
		entry += get_uint_byte_size(ele)
		// entry
		total += TypeEntry.EncodingLength()
		total += get_uint_byte_size(uint(entry))
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
	for key, ele := range sv.entries {
		n, _ = enc.NameFromStr(key)
		t = n.EncodingLength()
		// entry length
		entryLen = enc.TypeName.EncodingLength()
		entryLen += get_uint_byte_size(uint(t))
		entryLen += t
		entryLen += TypeEntrySeqno.EncodingLength()
		entryLen += get_uint_byte_size(uint(get_uint_byte_size(ele)))
		entryLen += get_uint_byte_size(ele)
		offset = TypeEntry.EncodingLength() + get_uint_byte_size(uint(entryLen))
		entryLen = offset
		// source
		entryLen += enc.TypeName.EncodeInto(buf[pos+entryLen:])
		entryLen += write_uint(uint(t), buf, pos+entryLen)
		entryLen += n.EncodeInto(buf[pos+entryLen:])
		// seqno
		entryLen += TypeEntrySeqno.EncodeInto(buf[pos+entryLen:])
		entryLen += write_uint(uint(get_uint_byte_size(ele)), buf, pos+entryLen)
		entryLen += write_uint(ele, buf, pos+entryLen)
		// entry
		entryLen -= offset
		pos += TypeEntry.EncodeInto(buf[pos:])
		pos += write_uint(uint(entryLen), buf, pos)
		pos += entryLen
	}
	return pos
}
