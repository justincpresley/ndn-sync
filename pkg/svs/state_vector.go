package svs

import (
	"strconv"
	"strings"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type StateVector interface {
	Set(string, enc.Name, uint64, bool)
	Get(string) uint64
	String() string
	Len() int
	Total() uint64
	Entries() *om.OrderedMap[uint64]
	Encode(bool) enc.Wire
}

type stateVector struct {
	entries *om.OrderedMap[uint64]
}

func NewStateVector() StateVector {
	return stateVector{entries: om.New[uint64](om.LatestEntriesFirst)}
}

func ParseStateVector(reader enc.ParseReader, formal bool) (StateVector, error) {
	if formal {
		return parseFormalStateVector(reader)
	} else {
		return parseInformalStateVector(reader)
	}
}

func (sv stateVector) Set(source string, sname enc.Name, seqno uint64, old bool) {
	sv.entries.Set(source, sname, seqno, om.MetaV{Old: old})
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

func (sv stateVector) Total() uint64 {
	var total uint64
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		total += p.Value
	}
	return total
}

func (sv stateVector) Entries() *om.OrderedMap[uint64] {
	return sv.entries
}

func (sv stateVector) Encode(formal bool) enc.Wire {
	if formal {
		tl, ls := sv.formalEncodingLengths()
		// length
		e := TypeVector.EncodingLength()
		e += enc.TLNum(tl).EncodingLength()
		e += tl
		// space
		wire := make(enc.Wire, 1)
		wire[0] = make([]byte, e)
		buf := wire[0]
		// encode
		off := TypeVector.EncodeInto(buf)
		off += enc.TLNum(tl).EncodeInto(buf[off:])
		sv.formalEncodeInto(buf[off:], ls)
		return wire
	} else {
		tl := sv.informalEncodingLength()
		// length
		e := TypeVector.EncodingLength()
		e += enc.TLNum(tl).EncodingLength()
		e += tl
		// space
		wire := make(enc.Wire, 1)
		wire[0] = make([]byte, e)
		buf := wire[0]
		// encode
		off := TypeVector.EncodeInto(buf)
		off += enc.TLNum(tl).EncodeInto(buf[off:])
		sv.informalEncodeInto(buf[off:])
		return wire
	}
}

func (sv stateVector) formalEncodingLengths() (int, []int) {
	var (
		e, tl, nl, i int
		ls           = make([]int, sv.entries.Len())
	)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		nl = p.Kname.EncodingLength()
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

func (sv stateVector) formalEncodeInto(buf []byte, ls []int) int {
	var (
		el, off, pos, i int
	)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		el = ls[i]
		off = TypeEntry.EncodingLength() + enc.TLNum(el).EncodingLength()
		// source
		off += enc.TypeName.EncodeInto(buf[pos+off:])
		off += enc.TLNum(p.Kname.EncodingLength()).EncodeInto(buf[pos+off:])
		off += p.Kname.EncodeInto(buf[pos+off:])
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

func (sv stateVector) informalEncodingLength() int {
	var (
		e, nl int
	)
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		nl = p.Kname.EncodingLength()
		// source
		e += enc.TypeName.EncodingLength()
		e += enc.TLNum(nl).EncodingLength()
		e += nl
		// seqno
		e += TypeEntrySeqno.EncodingLength()
		e += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodingLength()
		e += enc.Nat(p.Value).EncodingLength()
	}
	return e
}

func (sv stateVector) informalEncodeInto(buf []byte) int {
	var pos int
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		// source
		pos += enc.TypeName.EncodeInto(buf[pos:])
		pos += enc.TLNum(p.Kname.EncodingLength()).EncodeInto(buf[pos:])
		pos += p.Kname.EncodeInto(buf[pos:])
		// seqno
		pos += TypeEntrySeqno.EncodeInto(buf[pos:])
		pos += enc.TLNum(enc.Nat(p.Value).EncodingLength()).EncodeInto(buf[pos:])
		pos += enc.Nat(p.Value).EncodeInto(buf[pos:])
	}
	return pos
}

func parseFormalStateVector(reader enc.ParseReader) (StateVector, error) {
	if reader == nil {
		return NewStateVector(), enc.ErrBufferOverflow
	}
	var (
		source enc.Name
		seqno  enc.Nat
		l, t   enc.TLNum
		b      enc.Buffer
		end    int
		err    error
		ret    StateVector = NewStateVector()
	)
	// vector
	t, err = enc.ReadTLNum(reader)
	if err != nil {
		return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
	}
	if t != TypeVector {
		return ret, enc.ErrUnrecognizedField{TypeNum: t}
	}
	l, err = enc.ReadTLNum(reader)
	if err != nil {
		return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
	}
	if reader.Length()-reader.Pos() < int(l) {
		return ret, enc.ErrFailToParse{TypeNum: t}
	}
	// entries
	end = int(l)
	for reader.Pos() < end {
		// entry
		t, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		if t != TypeEntry {
			return ret, enc.ErrUnrecognizedField{TypeNum: t}
		}
		_, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		// source
		t, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		if t != enc.TypeName {
			return ret, enc.ErrUnrecognizedField{TypeNum: t}
		}
		l, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		source, err = enc.ReadName(reader.Delegate(int(l)))
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		// seqno
		t, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		if t != TypeEntrySeqno {
			return ret, enc.ErrUnrecognizedField{TypeNum: t}
		}
		l, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		b, err = reader.ReadBuf(int(l))
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		seqno, _ = enc.ParseNat(b)
		// add
		ret.Set(source.String(), source, uint64(seqno), true)
	}
	return ret, nil
}

func parseInformalStateVector(reader enc.ParseReader) (StateVector, error) {
	if reader == nil {
		return NewStateVector(), enc.ErrBufferOverflow
	}
	var (
		source enc.Name
		seqno  enc.Nat
		l, t   enc.TLNum
		b      enc.Buffer
		end    int
		err    error
		ret    StateVector = NewStateVector()
	)
	// vector
	t, err = enc.ReadTLNum(reader)
	if err != nil {
		return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
	}
	if t != TypeVector {
		return ret, enc.ErrUnrecognizedField{TypeNum: t}
	}
	l, err = enc.ReadTLNum(reader)
	if err != nil {
		return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
	}
	if reader.Length()-reader.Pos() < int(l) {
		return ret, enc.ErrFailToParse{TypeNum: t}
	}
	// entries
	end = int(l)
	for reader.Pos() < end {
		// source
		t, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		if t != enc.TypeName {
			return ret, enc.ErrUnrecognizedField{TypeNum: t}
		}
		l, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		source, err = enc.ReadName(reader.Delegate(int(l)))
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		// seqno
		t, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		if t != TypeEntrySeqno {
			return ret, enc.ErrUnrecognizedField{TypeNum: t}
		}
		l, err = enc.ReadTLNum(reader)
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		b, err = reader.ReadBuf(int(l))
		if err != nil {
			return ret, enc.ErrFailToParse{TypeNum: t, Err: err}
		}
		seqno, _ = enc.ParseNat(b)
		// add
		ret.Set(source.String(), source, uint64(seqno), true)
	}
	return ret, nil
}
