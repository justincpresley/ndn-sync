package svs

import (
	"errors"
	"strconv"
	"strings"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(string, enc.Name, uint64)
	Get(string) uint64
	String() string
	Len() int
	Entries() *om.OrderedMap[uint64]
	ToComponent() enc.Component
	EncodingLength() int
	EncodeInto([]byte) int
}

type stateVector struct {
	entries *om.OrderedMap[uint64]
}

func NewStateVector() StateVector {
	return stateVector{entries: om.New[uint64]()}
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
		if enc.TLNum(typ) != TypeEntrySeqno {
			return NewStateVector(), errors.New("encoding.ParseStatevector: incorrect tlv type")
		}
		length, temp = enc.ParseTLNum(buf[pos:])
		pos += temp
		seqno, _ = enc.ParseNat(buf[pos : pos+int(length)])
		pos += int(length)
		// add the entry
		ret.Set(source.String(), source, uint64(seqno))
	}
	return ret, nil
}

func (sv stateVector) Set(source string, sname enc.Name, seqno uint64) {
	sv.entries.Set(source, sname, seqno)
}

func (sv stateVector) Get(source string) uint64 {
	if val, pres := sv.entries.Get(source); pres {
		return val
	}
	return 0
}

func (sv stateVector) String() string {
	var b strings.Builder
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		b.WriteString(p.Kstring)
		b.WriteString(":")
		b.WriteString(strconv.FormatUint(p.Value, 10))
		b.WriteString(" ")
	}
	if b.Len() <= 0 {
		return ""
	}
	return b.String()[:b.Len()-1]
}

func (sv stateVector) Len() int {
	return sv.entries.Len()
}

func (sv stateVector) Entries() *om.OrderedMap[uint64] {
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
	)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		t = p.Kname.EncodingLength()
		// source
		entry = enc.TypeName.EncodingLength()
		entry += enc.TLNum(t).EncodingLength()
		entry += t
		// seqno
		entry += TypeEntrySeqno.EncodingLength()
		entry += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodingLength()
		entry += enc.Nat(p.Value).EncodingLength()
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
		t        int
	)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		t = p.Kname.EncodingLength()
		// entry length
		entryLen = enc.TypeName.EncodingLength()
		entryLen += enc.TLNum(t).EncodingLength()
		entryLen += t
		entryLen += TypeEntrySeqno.EncodingLength()
		entryLen += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodingLength()
		entryLen += enc.Nat(p.Value).EncodingLength()
		offset = TypeEntry.EncodingLength() + enc.TLNum(entryLen).EncodingLength()
		entryLen = offset
		// source
		entryLen += enc.TypeName.EncodeInto(buf[pos+entryLen:])
		entryLen += enc.TLNum(t).EncodeInto(buf[pos+entryLen:])
		entryLen += p.Kname.EncodeInto(buf[pos+entryLen:])
		// seqno
		entryLen += TypeEntrySeqno.EncodeInto(buf[pos+entryLen:])
		entryLen += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodeInto(buf[pos+entryLen:])
		entryLen += enc.Nat(p.Value).EncodeInto(buf[pos+entryLen:])
		// entry
		entryLen -= offset
		pos += TypeEntry.EncodeInto(buf[pos:])
		pos += enc.TLNum(entryLen).EncodeInto(buf[pos:])
		pos += entryLen
	}
	return pos
}
