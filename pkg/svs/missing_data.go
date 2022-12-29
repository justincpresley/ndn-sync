package svs

type MissingData interface {
	Source() string
	LowSeqno() uint64
	HighSeqno() uint64
}

type missingData struct {
	source    string
	lowSeqno  uint64
	highSeqno uint64
}

func NewMissingData(source string, low uint64, high uint64) MissingData {
	return missingData{source: source, lowSeqno: low, highSeqno: high}
}

func (md missingData) Source() string {
	return md.source
}

func (md missingData) LowSeqno() uint64 {
	return md.lowSeqno
}

func (md missingData) HighSeqno() uint64 {
	return md.highSeqno
}
