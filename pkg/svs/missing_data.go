package svs

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type MissingData interface {
	Source() enc.Name
	Increment()
	LowSeqno() uint64
	HighSeqno() uint64
}

type missingData struct {
	source    enc.Name
	lowSeqno  uint64
	highSeqno uint64
}

func NewMissingData(source enc.Name, low uint64, high uint64) MissingData {
	return &missingData{source: source, lowSeqno: low, highSeqno: high}
}

func (md *missingData) Source() enc.Name  { return md.source }
func (md *missingData) Increment()        { md.lowSeqno++ }
func (md *missingData) LowSeqno() uint64  { return md.lowSeqno }
func (md *missingData) HighSeqno() uint64 { return md.highSeqno }
