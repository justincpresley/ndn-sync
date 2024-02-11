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

func CopyStateVector(sv stateVector) StateVector {
	return stateVector{entries: sv.entries.Copy()}
}

func ParseStateVector(reader enc.ParseReader, formal bool) (StateVector, error) {
	if formal {
		return parseFormalStateVector(reader)
	} else {
		return parseInformalStateVector(reader)
	}
}

func (sv stateVector) Set(dsstr string, dsname enc.Name, seqno uint64, old bool) {
	sv.entries.Set(dsstr, dsname, seqno, om.MetaV{Old: old})
}

func (sv stateVector) Get(dsstr string) uint64 {
	if val, ok := sv.entries.Get(dsstr); ok {
		return val
	}
	return 0
}

func (sv stateVector) String() string {
	var ret strings.Builder
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		ret.WriteString(p.Kstr)
		ret.WriteString(":")
		ret.WriteString(strconv.FormatUint(p.Val, 10))
		ret.WriteString(" ")
	}
	if ret.Len() <= 0 {
		return ""
	}
	return ret.String()[:ret.Len()-1]
}

func (sv stateVector) Len() int {
	return sv.entries.Len()
}

func (sv stateVector) Total() uint64 {
	var ret uint64
	for p := sv.entries.Front(); p != nil; p = p.Next() {
		ret += p.Val
	}
	return ret
}

func (sv stateVector) Entries() *om.OrderedMap[uint64] {
	return sv.entries
}

func (sv stateVector) Encode(formal bool) enc.Wire {
	if formal {
		tl, ls := sv.formalEncodingLengths()
		// length
		pos := TypeVector.EncodingLength()
		pos += enc.TLNum(tl).EncodingLength()
		pos += tl
		// space
		ret := make(enc.Wire, 1)
		ret[0] = make([]byte, pos)
		buf := ret[0]
		// encode
		pos = TypeVector.EncodeInto(buf)
		pos += enc.TLNum(tl).EncodeInto(buf[pos:])
		sv.formalEncodeInto(buf[pos:], ls)
		return ret
	} else {
		tl := sv.informalEncodingLength()
		// length
		pos := TypeVector.EncodingLength()
		pos += enc.TLNum(tl).EncodingLength()
		pos += tl
		// space
		ret := make(enc.Wire, 1)
		ret[0] = make([]byte, pos)
		buf := ret[0]
		// encode
		pos = TypeVector.EncodeInto(buf)
		pos += enc.TLNum(tl).EncodeInto(buf[pos:])
		sv.informalEncodeInto(buf[pos:])
		return ret
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
		e += enc.TLNum(enc.Nat(p.Val).EncodingLength()).EncodingLength()
		e += enc.Nat(p.Val).EncodingLength()
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
		off += enc.TLNum(enc.Nat(p.Val).EncodingLength()).EncodeInto(buf[pos+off:])
		enc.Nat(p.Val).EncodeInto(buf[pos+off:])
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
		e += enc.TLNum(enc.Nat(p.Val).EncodingLength()).EncodingLength()
		e += enc.Nat(p.Val).EncodingLength()
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
		pos += enc.TLNum(enc.Nat(p.Val).EncodingLength()).EncodeInto(buf[pos:])
		pos += enc.Nat(p.Val).EncodeInto(buf[pos:])
	}
	return pos
}

func parseFormalStateVector(reader enc.ParseReader) (StateVector, error) {
	if reader == nil {
		return NewStateVector(), enc.ErrBufferOverflow
	}
	var (
		dsname enc.Name
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
		// dsname
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
		dsname, err = enc.ReadName(reader.Delegate(int(l)))
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
		ret.Set(dsname.String(), dsname, uint64(seqno), true)
	}
	return ret, nil
}

func parseInformalStateVector(reader enc.ParseReader) (StateVector, error) {
	if reader == nil {
		return NewStateVector(), enc.ErrBufferOverflow
	}
	var (
		dsname enc.Name
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
		// dsname
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
		dsname, err = enc.ReadName(reader.Delegate(int(l)))
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
		ret.Set(dsname.String(), dsname, uint64(seqno), true)
	}
	return ret, nil
}
