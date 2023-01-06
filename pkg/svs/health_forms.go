package svs

import (
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
)

type HealthSync interface {
	Listen()
	Activate(bool)
	Shutdown()
	// TODO: NeedStatus + PublishStatus (publish/pull at a rate)
	Core() Core
	Tracker() Tracker
}

type HealthConfig struct {
	Source      enc.Name
	GroupPrefix enc.Name
}

func NewHealthSync(app *eng.Engine, config *HealthConfig, constants *Constants) HealthSync {
	return newHealthSync(app, config, constants)
}
