package svs

import (
	"time"

	log "github.com/apex/log"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
)

type healthHandlerData struct {
	done chan struct{}
}

type healthSync struct {
	app         *eng.Engine
	core        Core
	tracker     Tracker
	constants   *Constants
	groupPrefix enc.Name
	source      enc.Name
	sourceStr   string
	logger      *log.Entry
	handleData  *healthHandlerData
}

func newHealthSync(app *eng.Engine, config *HealthConfig, constants *Constants) *healthSync {
	var s *healthSync
	logger := log.WithField("module", "svs")
	syncComp, _ := enc.ComponentFromStr("sync")
	syncPrefix := append(config.GroupPrefix, syncComp)

	// TODO: switch to new core type (check then switch)
	coreConfig := &TwoStateCoreConfig{
		Source:     config.Source,
		SyncPrefix: syncPrefix,
	}
	s = &healthSync{
		app:         app,
		core:        NewCore(app, coreConfig, constants),
		tracker:     NewTracker(config.Source.String(), constants),
		constants:   constants,
		groupPrefix: config.GroupPrefix,
		source:      config.Source,
		sourceStr:   config.Source.String(),
		logger:      logger,
	}

	hData := &healthHandlerData{
		done: make(chan struct{}),
	}
	s.handleData = hData
	s.newHandling(hData)

	return s
}

func (s *healthSync) Listen() {
	s.core.Listen()
}

func (s *healthSync) Activate(immediateStart bool) {
	s.core.Activate(immediateStart)
	s.logger.Info("Sync Activated.")
}

func (s *healthSync) Shutdown() {
	s.core.Shutdown()
	if s.handleData != nil {
		<-s.handleData.done
	}
	s.logger.Info("Sync Shutdown.")
}

func (s *healthSync) Tracker() Tracker {
	return s.tracker
}

func (s *healthSync) Core() Core {
	return s.core
}

func (s *healthSync) newHandling(data *healthHandlerData) {
	go func() {
		missingChan := s.Core().Chan()
		for {
			if s.tracker.UntilBeat() < s.constants.MonitorInterval {
				s.core.SetSeqno(s.core.Seqno() + 1)
				s.tracker.Reset(s.sourceStr)
			}
			s.tracker.Detect()
			select {
			case missing, ok := <-missingChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for _, m := range missing {
					s.tracker.Reset(m.Source())
				}
			default:
			}
			time.Sleep(s.constants.MonitorInterval)
		}
	}()
}
