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

	om "github.com/justincpresley/ndn-sync/internal/orderedmap"
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
}

type stateVector struct {
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
		source string
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
		// source
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if enc.TLNum(typ) != TypeEntrySource {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = parse_uint(buf, pos)
		pos += temp
		source = string(buf[pos : pos+length])
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
		// add the component
		ret.Set(source, seqno, true)
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
	for pair := sv.entries.Oldest(); pair != nil; pair = pair.Next() {
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
	for pair := sv.entries.Oldest(); pair != nil; pair = pair.Next() {
		total += pair.Value
	}
	return total
}

func (sv stateVector) Entries() *om.OrderedMap[string, uint] {
	return sv.entries
}

func (sv stateVector) ToComponent() enc.Component {
	var (
		pos    int
		length int = 2 * sv.entries.Len()
		pair   *om.Pair[string, uint]
	)
	// component value space
	for pair = sv.entries.Oldest(); pair != nil; pair = pair.Next() {
		length += get_uint_byte_size(uint(len(pair.Key)))
		length += len(pair.Key)
		length += get_uint_byte_size(uint(get_uint_byte_size(pair.Value)))
		length += get_uint_byte_size(pair.Value)
	}
	// make and fill the component
	comp := enc.Component{
		Typ: TypeVector,
		Val: make([]byte, length),
	}
	buf := comp.Val
	for pair = sv.entries.Oldest(); pair != nil; pair = pair.Next() {
		pos += TypeEntrySource.EncodeInto(buf[pos:])
		pos += write_uint(uint(len(pair.Key)), buf, pos)
		copy(buf[pos:], pair.Key)
		pos += len(pair.Key)
		pos += TypeEntrySeqno.EncodeInto(buf[pos:])
		pos += write_uint(uint(get_uint_byte_size(pair.Value)), buf, pos)
		pos += write_uint(pair.Value, buf, pos)
	}
	return comp
}
