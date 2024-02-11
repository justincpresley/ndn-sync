package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type NativeSync interface {
	Listen()
	Activate(bool)
	Shutdown()
	NeedData(enc.Name, uint64)
	PublishData([]byte)
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	Core() Core
}

type NativeConfig struct {
	Source               enc.Name
	GroupPrefix          enc.Name
	NamingScheme         NamingScheme
	HandlingOption       HandlingOption
	StoragePath          string
	DataCallback         func(source enc.Name, seqno uint64, data ndn.Data)
	FormalEncoding       bool
	EfficientSuppression bool
}

func NewNativeSync(app *eng.Engine, config *NativeConfig, constants *Constants) NativeSync {
	return newNativeSync(app, config, constants)
}

func GetBasicNativeConfig(source enc.Name, group enc.Name, callback func(source enc.Name, seqno uint64, data ndn.Data)) *NativeConfig {
	return &NativeConfig{
		Source:               source,
		GroupPrefix:          group,
		NamingScheme:         SourceOrientedNaming,
		HandlingOption:       SourceCentricHandling,
		StoragePath:          "./" + source.String() + "_bolt.db",
		DataCallback:         callback,
		FormalEncoding:       false,
		EfficientSuppression: true,
	}
}
