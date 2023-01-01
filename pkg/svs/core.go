package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

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

type CoreConfig interface {*TwoStateCoreConfig}

type TwoStateCoreConfig struct {
	Source     enc.Name
	SyncPrefix enc.Name
}

func NewCore[T CoreConfig](app *eng.Engine, config T, constants *Constants) Core {
	return newTwoStateCore(app, config, constants)
}
