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
	StateVector() StateVector
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	Subscribe() chan SyncUpdate
}

type CoreConfig struct {
	SyncPrefix enc.Name
}

func NewCore(app *eng.Engine, config *CoreConfig, constants *Constants) Core {
	return newTwoStateCore(app, config, constants)
}
