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
	Subscribe() chan SyncUpdate
}

type OneStateCoreConfig struct {
	Source         enc.Name
	SyncPrefix     enc.Name
	FormalEncoding bool
}

type TwoStateCoreConfig struct {
	Source         enc.Name
	SyncPrefix     enc.Name
	FormalEncoding bool
}

// TODO: change after threestatecore is added
func NewCore(app *eng.Engine, config interface{}, constants *Constants) Core {
	switch config.(type) {
	case OneStateCoreConfig:
		return newOneStateCore(app, config.(*OneStateCoreConfig), constants)
	case TwoStateCoreConfig:
		return newTwoStateCore(app, config.(*TwoStateCoreConfig), constants)
	default:
		return newNullCore()
	}
}
