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
	List() map[string]uint
	ToComponent() enc.Component
}

type statevector struct {
	Entries map[string]uint
}

func NewStateVector() StateVector {
	return statevector{Entries: make(map[string]uint)}
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

func (sv statevector) Set(source string, seqno uint) {
	sv.Entries[source] = seqno
}

func (sv statevector) List() map[string]uint {
	return sv.Entries
}

func (sv statevector) Get(source string) uint {
	return sv.Entries[source]
}

func (sv statevector) String() string {
	str := ""
	for key, ele := range sv.Entries {
		str += key + ":" + strconv.FormatUint(uint64(ele), 10) + " "
	}
	return str // has an extra space
}

func (sv statevector) Len() int {
	return len(sv.Entries)
}

func (sv statevector) ToComponent() enc.Component {
	var (
		length uint = 2 * uint(len(sv.Entries))
		pos    uint
	)
	// component value space
	for key, ele := range sv.Entries {
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
	for key, ele := range sv.Entries {
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
