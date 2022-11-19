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
	EncodingLengths() (int, []int)
	EncodeInto(buf []byte, lens []int) int
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
		source      enc.Name
		seqno       enc.Nat
		length, typ enc.TLNum
		buf         []byte = comp.Val
		pos, temp   int
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
	if val, pres := sv.entries.Get(source); pres {
		return val
	}
	return 0
}

func (sv stateVector) String() string {
	str := ""
	for p := sv.entries.First(); p != nil; p = p.Next() {
		str += p.Key + ":" + strconv.FormatUint(p.Value, 10) + " "
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
	for p := sv.entries.First(); p != nil; p = p.Next() {
		total += p.Value
	}
	return total
}

func (sv stateVector) Entries() *om.OrderedMap[string, uint64] {
	return sv.entries
}

func (sv stateVector) ToComponent() enc.Component {
	tl, ls := sv.EncodingLengths()
	comp := enc.Component{
		Typ: TypeVector,
		Val: make([]byte, tl),
	}
	sv.EncodeInto(comp.Val, ls)
	return comp
}

func (sv stateVector) EncodingLengths() (int, []int) {
	var (
		e, tl, nl, i int
		ls           []int = make([]int, sv.entries.Len())
		n            enc.Name
	)
	for p := sv.entries.First(); p != nil; p = p.Next() {
		n, _ = enc.NameFromStr(p.Key)
		nl = n.EncodingLength()
		// source
		e = enc.TypeName.EncodingLength()
		e += enc.TLNum(nl).EncodingLength()
		e += nl
		// seqno
		e += TypeEntrySeqno.EncodingLength()
		e += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodingLength()
		e += enc.Nat(p.Value).EncodingLength()
		// entry
		tl += TypeEntry.EncodingLength()
		tl += enc.TLNum(e).EncodingLength()
		tl += e
		ls[i] = e
		i++
	}
	return tl, ls
}

func (sv stateVector) EncodeInto(buf []byte, ls []int) int {
	var (
		el, off, pos, i int
		n               enc.Name
	)
	for p := sv.entries.First(); p != nil; p = p.Next() {
		n, _ = enc.NameFromStr(p.Key)
		el = ls[i]
		off = TypeEntry.EncodingLength() + enc.TLNum(el).EncodingLength()
		// source
		off += enc.TypeName.EncodeInto(buf[pos+off:])
		off += enc.TLNum(n.EncodingLength()).EncodeInto(buf[pos+off:])
		off += n.EncodeInto(buf[pos+off:])
		// seqno
		off += TypeEntrySeqno.EncodeInto(buf[pos+off:])
		off += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodeInto(buf[pos+off:])
		enc.Nat(p.Value).EncodeInto(buf[pos+off:])
		// entry
		pos += TypeEntry.EncodeInto(buf[pos:])
		pos += enc.TLNum(el).EncodeInto(buf[pos:])
		pos += el
		i++
	}
	return pos
}
