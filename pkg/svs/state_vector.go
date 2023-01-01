package svs

import (
	"errors"
	"strconv"
	"strings"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(string, uint64, bool)
	Get(string) uint64
	String() string
	Len() int
	Total() uint64
	Entries() *om.OrderedMap[string, uint64]
	ToComponent() enc.Component
	EncodingLengths() (int, []int)
	EncodeInto([]byte, []int) int
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
	var b strings.Builder
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		b.WriteString(p.Key)
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

func (sv stateVector) Total() uint64 {
	total := uint64(0)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
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
	for p := sv.entries.Front(); p != nil; p = p.Next() {
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
	for p := sv.entries.Front(); p != nil; p = p.Next() {
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
