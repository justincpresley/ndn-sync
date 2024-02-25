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
	Update(enc.Name, uint64)
	StateVector() *StateVector
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	Subscribe() chan SyncUpdate
}

type OneStateCoreConfig struct {
	SyncPrefix     enc.Name
	FormalEncoding bool
}

type TwoStateCoreConfig struct {
	SyncPrefix           enc.Name
	FormalEncoding       bool
	EfficientSuppression bool
}

// TODO: change after threestatecore is added
func NewCore(app *eng.Engine, config interface{}, constants *Constants) Core {
	switch config.(type) {
	case *OneStateCoreConfig:
		return newOneStateCore(app, config.(*OneStateCoreConfig), constants)
	case *TwoStateCoreConfig:
		return newTwoStateCore(app, config.(*TwoStateCoreConfig), constants)
	default:
		return newNullCore()
	}
}
