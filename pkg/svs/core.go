package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type CoreState int32

const (
	Steady      CoreState = 0
	Suppression CoreState = 1
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
	GetSeqno() uint64
	GetStateVector() StateVector
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	MissingChan() chan *[]MissingData
}

func NewCore(app *eng.Engine, config *CoreConfig, constants *Constants) Core {
	return newTwoStateCore(app, config, constants)
}
