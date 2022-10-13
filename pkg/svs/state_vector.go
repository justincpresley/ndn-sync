/*
 Copyright (C) 2022-2025, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 This library is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 This library is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. To see more details about the authors and
 contributors, please see AUTHORS.md. If absent, Both of which can be
 found within the GitHub repository:
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
}

type stateVector struct {
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
		source string
		seqno  uint
		buf    []byte = comp.Val
		pos    uint
		length uint
		temp   uint
		typ    uint
	)
	// verify type
	if uint(comp.Typ) != TlvTypeVector {
		return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
	}
	// decode components
	ret = NewStateVector()
	for pos < uint(len(buf)) {
		// source
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if typ != TlvTypeEntrySource {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = parse_uint(buf, pos)
		pos += temp
		source = string(buf[pos : pos+length])
		pos += length
		// seqno
		typ, temp = parse_uint(buf, pos)
		pos += temp
		if typ != TlvTypeEntrySeqno {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = parse_uint(buf, pos)
		pos += temp
		seqno, _ = parse_uint(buf, pos)
		pos += length
		// add the component
		ret.Set(source, seqno)
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
	return str // has an extra space
}

func (sv stateVector) Len() int {
	return len(sv.entries)
}

func (sv stateVector) ToComponent() enc.Component {
	var (
		length uint = 2 * uint(len(sv.entries))
		pos    uint
	)
	// component value space
	for key, ele := range sv.entries {
		length += get_uint_byte_size(uint(len(key)))
		length += uint(len(key))
		length += get_uint_byte_size(get_uint_byte_size(ele))
		length += get_uint_byte_size(ele)
	}
	// make and fill the component
	comp := enc.Component{
		Typ: enc.TLNum(TlvTypeVector),
		Val: make([]byte, length),
	}
	buf := comp.Val
	for key, ele := range sv.entries {
		pos += write_uint(TlvTypeEntrySource, buf, pos)
		pos += write_uint(uint(len(key)), buf, pos)
		copy(buf[pos:], key)
		pos += uint(len(key))
		pos += write_uint(TlvTypeEntrySeqno, buf, pos)
		pos += write_uint(get_uint_byte_size(ele), buf, pos)
		pos += write_uint(ele, buf, pos)
	}
	return comp
}
