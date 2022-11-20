package bloomfilter

import (
	"bufio"
	"bytes"
	"io"
)

type bitset struct {
	set []uint8
	len uint32
}

func newBitset(n uint32) *bitset {
	return &bitset{
		set: make([]uint8, (n+7)/8),
		len: n,
	}
}

func parseBitset(buf []byte) (*bitset, error) {
	r := bytes.NewBuffer(buf)
	l, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	b := r.Bytes()
	return &bitset{
		set: b,
		len: (uint32(len(b)) * 8) - uint32(l),
	}, nil
}

func (b *bitset) copy() *bitset {
	rtn := newBitset(b.len)
	copy(rtn.set, b.set)
	return rtn
}

func (b *bitset) test(index uint) bool {
	return (b.set[index/8] & (uint8(1) << (index % 8))) != 0
}

func (b *bitset) operate(index uint, value bool) {
	if value {
		b.set[index/8] |= (uint8(1) << (index % 8))
	} else {
		b.set[index/8] &= ^(uint8(1) << (index % 8))
	}
}

func (b *bitset) length() uint32 {
	return b.len
}

func (b *bitset) clear() {
	b.set = make([]uint8, (b.len+7)/8)
}

func (b *bitset) writeTo(stream io.Writer) error {
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

func (b *bitset) writeSize() uint32 {
	return 1 + (b.len+7)/8
}

func (b *bitset) bytes() ([]byte, error) {
	var buf bytes.Buffer
	buf.Grow(int(b.writeSize()))
	err := b.writeTo(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), err
}
