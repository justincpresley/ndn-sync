package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type SharedSync interface {
	Listen()
	Activate(bool)
	Shutdown()
	NeedData(enc.Name, uint64, bool)
	PublishData([]byte)
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	Core() Core
}

type SharedConfig struct {
	Source               enc.Name
	GroupPrefix          enc.Name
	HandlingOption       HandlingOption
	StoragePath          string
	DataCallback         func(enc.Name, uint64, ndn.Data)
	FormalEncoding       bool
	EfficientSuppression bool
	// high-level only
	CacheOthers bool
}

func NewSharedSync(app *eng.Engine, config *SharedConfig, constants *Constants) SharedSync {
	return newSharedSync(app, config, constants)
}

func GetBasicSharedConfig(source enc.Name, group enc.Name, callback func(source enc.Name, seqno uint64, data ndn.Data)) *SharedConfig {
	return &SharedConfig{
		Source:         source,
		GroupPrefix:    group,
		HandlingOption: SourceCentricHandling,
		StoragePath:    "./" + source.String() + "_bolt.db",
		DataCallback:   callback,
		FormalEncoding: false,
		CacheOthers:    true,
	}
}
