package bitset

import (
	"bufio"
	"bytes"
)

type BitSet struct {
	set []uint8
	len uint32
}

func New(n uint32) *BitSet {
	return &BitSet{
		set: make([]uint8, (n+7)/8),
		len: n,
	}
}

func Parse(buf []byte) (*BitSet, error) {
	r := bytes.NewBuffer(buf)
	l, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	b := r.Bytes()
	return &BitSet{
		set: b,
		len: (uint32(len(b)) * 8) - uint32(l),
	}, nil
}

func (b *BitSet) Copy() *BitSet {
	rtn := New(b.len)
	copy(rtn.set, b.set)
	return rtn
}

func (b *BitSet) Test(index uint) bool {
	return (b.set[index/8] & (uint8(1) << (index % 8))) != 0
}

func (b *BitSet) Set(index uint, value bool) {
	if value {
		b.set[index/8] |= (uint8(1) << (index % 8))
	} else {
		b.set[index/8] &= ^(uint8(1) << (index % 8))
	}
}

func (b *BitSet) Len() uint32 {
	return b.len
}

func (b *BitSet) Clear() {
	b.set = make([]uint8, (b.len+7)/8)
}

func (b *BitSet) WriteTo(stream *bytes.Buffer) error {
	writer := bufio.NewWriter(stream)
	// pad len
	if err := writer.WriteByte(byte(8 - (b.len % 8))); err != nil {
		return err
	}
	// set
	if _, err := writer.Write(b.set); err != nil {
		return err
	}
	return writer.Flush()
}

func (b *BitSet) WriteSize() uint32 {
	return 1 + (b.len+7)/8
}

func (b *BitSet) Bytes() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(int(b.WriteSize()))
	err := b.WriteTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
