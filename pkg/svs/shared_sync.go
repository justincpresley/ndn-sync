package svs

import (
	"sync/atomic"
	"time"

	log "github.com/apex/log"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
	utl "github.com/zjkmxy/go-ndn/pkg/utils"
)

type sharedFetchItem struct {
	source  string
	seqno   uint64
	retries uint
	cache   bool
}

type sharedHandlerData struct {
	done  chan struct{}
	cache bool
}

type sharedSync struct {
	app         *eng.Engine
	core        Core
	constants   *Constants
	groupPrefix enc.Name
	source      enc.Name
	sourceStr   string
	storage     Database
	intCfg      *ndn.InterestConfig
	datCfg      *ndn.DataConfig
	dataComp    enc.Component
	logger      *log.Entry
	dataCall    func(source string, seqno uint64, data ndn.Data)
	fetchQueue  chan *sharedFetchItem
	handleData  *sharedHandlerData
	numFetches  *int32
	isListening bool
}

func newSharedSync(app *eng.Engine, config *SharedConfig, constants *Constants) *sharedSync {
	var s *sharedSync
	logger := log.WithField("module", "svs")
	syncComp, _ := enc.ComponentFromStr("sync")
	dataComp, _ := enc.ComponentFromStr("data")
	syncPrefix := append(config.GroupPrefix, syncComp)

	if config.DataCallback == nil {
		logger.Error("Fetcher based on NativeConfig needs DataCallback.")
		return nil
	}
	coreConfig := &CoreConfig{
		Source:     config.Source,
		SyncPrefix: syncPrefix,
	}
	storage, err := NewBoltDB(config.StoragePath, []byte("svs-packets"))
	if err != nil {
		logger.Errorf("Unable to create storage: %+v", err)
		return nil
	}
	s = &sharedSync{
		app:         app,
		core:        NewCore(app, coreConfig, constants),
		constants:   constants,
		groupPrefix: config.GroupPrefix,
		source:      config.Source,
		sourceStr:   config.Source.String(),
		storage:     storage,
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(time.Duration(constants.DataInterestLifeTime) * time.Millisecond),
		},
		datCfg: &ndn.DataConfig{
			ContentType: utl.IdPtr(ndn.ContentTypeBlob),
			Freshness:   utl.IdPtr(time.Duration(constants.DataPacketFressness) * time.Millisecond),
		},
		dataComp:   dataComp,
		logger:     logger,
		dataCall:   config.DataCallback,
		fetchQueue: make(chan *sharedFetchItem, constants.InitialFetchQueueLength),
		numFetches: new(int32),
	}

	hData := &sharedHandlerData{
		done:  make(chan struct{}),
		cache: config.CacheOthers,
	}
	if config.HandlingOption != NoHandling {
		s.handleData = hData
	}
	switch config.HandlingOption {
	case SourceCentricHandling:
		s.newSourceCentricHandling(hData)
	case EqualTrafficHandling:
		s.newEqualTrafficHandling(hData)
	default:
	}

	return s
}

func (s *sharedSync) Listen() {
	dataPrefix := append(s.groupPrefix, s.dataComp)
	err := s.app.AttachHandler(dataPrefix, s.onInterest)
	if err != nil {
		s.logger.Errorf("Unable to register handler: %+v", err)
		return
	}
	err = s.app.RegisterRoute(dataPrefix)
	if err != nil {
		s.logger.Errorf("Unable to register route: %+v", err)
		return
	}
	s.isListening = true
	s.logger.Info("Data-side Registered and Handled.")
	s.core.Listen()
}

func (s *sharedSync) Activate(immediateStart bool) {
	s.core.Activate(immediateStart)
	s.logger.Info("Sync Activated.")
}

func (s *sharedSync) Shutdown() {
	s.core.Shutdown()
	if s.isListening {
		dataPrefix := append(s.groupPrefix, s.dataComp)
		err := s.app.DetachHandler(dataPrefix)
		if err != nil {
			s.logger.Errorf("Detech handler error: %+v", err)
		}
		err = s.app.UnregisterRoute(dataPrefix)
		if err != nil {
			s.logger.Errorf("Unregister route error: %+v", err)
		}
	}
	if s.handleData != nil {
		<-s.handleData.done
	}
	s.logger.Info("Sync Shutdown.")
}

func (s *sharedSync) NeedData(source string, seqno uint64, cache bool) {
	i := &sharedFetchItem{
		source:  source,
		seqno:   seqno,
		retries: s.constants.DataInterestRetries,
		cache:   cache,
	}
	if s.constants.MaxConcurrentDataInterests == 0 || atomic.LoadInt32(s.numFetches) < s.constants.MaxConcurrentDataInterests {
		atomic.AddInt32(s.numFetches, 1)
		s.sendInterest(i)
		return
	}
	s.fetchQueue <- i
}

func (s *sharedSync) PublishData(content []byte) {
	seqno := s.core.Seqno() + 1
	name := s.getDataName(s.sourceStr, seqno)
	wire, _, err := s.app.Spec().MakeData(
		name,
		s.datCfg,
		enc.Wire{content},
		sec.NewSha256Signer())
	if err != nil {
		s.logger.Errorf("unable to encode data: %+v", err)
		return
	}
	bytes := wire.Join()
	if len(bytes) > 8800 {
		s.logger.Warn("publication too large to be published")
		return
	}
	s.logger.Info("Publishing data " + name.String())
	s.storage.Set(name.Bytes(), bytes)
	s.core.SetSeqno(seqno)
}

func (s *sharedSync) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	s.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (s *sharedSync) Core() Core {
	return s.core
}

func (s *sharedSync) sendInterest(item *sharedFetchItem) {
	wire, _, finalName, err := s.app.Spec().MakeInterest(s.getDataName(item.source, item.seqno), s.intCfg, nil, nil)
	if err != nil {
		s.logger.Errorf("Unable to make Interest: %+v", err)
		return
	}
	err = s.app.Express(finalName, s.intCfg, wire,
		func(result ndn.InterestResult, data ndn.Data, rawData, sigCovered enc.Wire, nackReason uint64) {
			if result == ndn.InterestResultData || result == ndn.InterestResultNack || item.retries == 0 {
				if item.cache && result == ndn.InterestResultData {
					s.storage.Set(finalName.Bytes(), rawData.Join())
				}
				s.dataCall(item.source, item.seqno, data)
				atomic.AddInt32(s.numFetches, -1)
				s.processQueue()
			} else {
				item.retries--
				s.sendInterest(item)
			}
		})
	if err != nil {
		s.logger.Errorf("Unable to send Interest: %+v", err)
		return
	}
}

func (s *sharedSync) processQueue() {
	if s.constants.MaxConcurrentDataInterests == 0 || atomic.LoadInt32(s.numFetches) < s.constants.MaxConcurrentDataInterests {
		select {
		case f := <-s.fetchQueue:
			atomic.AddInt32(s.numFetches, 1)
			s.sendInterest(f)
			return
		default:
		}
	}
}

func (s *sharedSync) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	dataPkt := s.storage.Get(interest.Name().Bytes())
	if dataPkt != nil {
		s.logger.Info("Serving data " + interest.Name().String())
		err := reply(enc.Wire{dataPkt})
		if err != nil {
			s.logger.Errorf("unable to reply with data: %+v", err)
			return
		}
	}
}

func (s *sharedSync) getDataName(source string, seqno uint64) enc.Name {
	dataName := append(s.groupPrefix, s.dataComp)
	src, _ := enc.NameFromStr(source)
	dataName = append(dataName, src...)
	dataName = append(dataName, enc.NewSequenceNumComponent(seqno))
	return dataName
}

func (s *sharedSync) newSourceCentricHandling(data *sharedHandlerData) {
	go func() {
		missingChan := s.Core().Chan()
		for {
			select {
			case missing, ok := <-missingChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for _, m := range missing {
					for m.LowSeqno() <= m.HighSeqno() {
						s.NeedData(m.Source(), m.LowSeqno(), data.cache)
						m.Increment()
					}
				}
			}
		}
	}()
}

func (s *sharedSync) newEqualTrafficHandling(data *sharedHandlerData) {
	go func() {
		missingChan := s.Core().Chan()
		allFetched := true
		for {
			select {
			case missing, ok := <-missingChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for {
					allFetched = true
					for _, m := range missing {
						if m.LowSeqno() <= m.HighSeqno() {
							s.NeedData(m.Source(), m.LowSeqno(), data.cache)
							m.Increment()
							allFetched = false
						}
					}
					if allFetched {
						break
					}
				}
			}
		}
	}()
}
