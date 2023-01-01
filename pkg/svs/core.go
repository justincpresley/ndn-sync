package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type CoreConfig struct {
	Source     enc.Name
	SyncPrefix enc.Name
}

type Core interface {
	Listen()
	Activate(bool)
	Shutdown()
	SetSeqno(uint64)
	Seqno() uint64
	StateVector() StateVector
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	Chan() chan []MissingData
}

func NewCore(app *eng.Engine, config *CoreConfig, constants *Constants) Core {
	return newTwoStateCore(app, config, constants)
}
