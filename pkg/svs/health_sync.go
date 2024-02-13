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
	missChan    chan SyncUpdate
	tracker     Tracker
	constants   *Constants
	groupPrefix enc.Name
	srcName     enc.Name
	srcStr      string
	srcSeq      uint64
	logger      *log.Entry
	handleData  *healthHandlerData
}

func newHealthSync(app *eng.Engine, config *HealthConfig, constants *Constants) *healthSync {
	var s *healthSync
	logger := log.WithField("module", "svs")
	syncPrefix := append(config.GroupPrefix, constants.SyncComponent)

	coreConfig := &TwoStateCoreConfig{
		SyncPrefix:           syncPrefix,
		FormalEncoding:       config.FormalEncoding,
		EfficientSuppression: config.EfficientSuppression,
	}
	s = &healthSync{
		app:         app,
		core:        NewCore(app, coreConfig, constants),
		tracker:     NewTracker(config.Source.String(), constants),
		constants:   constants,
		groupPrefix: config.GroupPrefix,
		srcName:     config.Source,
		srcStr:      config.Source.String(),
		logger:      logger,
	}
	s.missChan = s.core.Subscribe()

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
		for {
			if s.tracker.UntilBeat() < s.constants.MonitorInterval {
				s.core.Update(s.srcName, s.srcSeq)
				s.tracker.Reset(s.srcStr)
			}
			s.tracker.Detect()
			select {
			case missing, ok := <-s.missChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for _, m := range missing {
					s.tracker.Reset(m.Source().String())
				}
			default:
			}
			time.Sleep(s.constants.MonitorInterval)
		}
	}()
}
