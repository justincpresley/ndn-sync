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

type nativeFetchItem struct {
	source  enc.Name
	seqno   uint64
	retries uint
}

type nativeHandlerData struct {
	done chan struct{}
}

type nativeSync struct {
	app          *eng.Engine
	core         Core
	constants    *Constants
	missChan     chan SyncUpdate
	namingScheme NamingScheme
	groupPrefix  enc.Name
	srcName      enc.Name
	storage      Database
	intCfg       *ndn.InterestConfig
	datCfg       *ndn.DataConfig
	logger       *log.Entry
	dataCall     func(enc.Name, uint64, ndn.Data)
	fetchQueue   chan *nativeFetchItem
	handleData   *nativeHandlerData
	numFetches   *int32
	isListening  bool
}

func newNativeSync(app *eng.Engine, config *NativeConfig, constants *Constants) *nativeSync {
	var s *nativeSync
	logger := log.WithField("module", "svs")
	syncPrefix := config.GroupPrefix
	if config.NamingScheme != BareSourceOrientedNaming {
		syncPrefix = append(syncPrefix, constants.SyncComponent)
	}

	coreConfig := &TwoStateCoreConfig{
		Source:         config.Source,
		SyncPrefix:     syncPrefix,
		FormalEncoding: config.FormalEncoding,
	}
	storage, err := NewBoltDB(config.StoragePath, []byte("svs-packets"))
	if err != nil {
		logger.Errorf("Unable to create storage: %+v", err)
		return nil
	}
	s = &nativeSync{
		app:          app,
		core:         NewCore(app, coreConfig, constants),
		constants:    constants,
		namingScheme: config.NamingScheme,
		groupPrefix:  config.GroupPrefix,
		srcName:      config.Source,
		storage:      storage,
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(constants.DataInterestLifeTime),
		},
		datCfg: &ndn.DataConfig{
			ContentType: utl.IdPtr(ndn.ContentTypeBlob),
			Freshness:   utl.IdPtr(constants.DataPacketFreshness),
		},
		logger:     logger,
		dataCall:   config.DataCallback,
		fetchQueue: make(chan *nativeFetchItem, constants.InitialFetchQueueLength),
		numFetches: new(int32),
	}
	s.missChan = s.core.Subscribe()

	hData := &nativeHandlerData{
		done: make(chan struct{}),
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

func (s *nativeSync) Listen() {
	dataPrefix := s.groupPrefix
	if s.namingScheme != BareSourceOrientedNaming {
		dataPrefix = append(dataPrefix, s.constants.DataComponent)
	}
	if s.namingScheme == GroupOrientedNaming {
		dataPrefix = append(dataPrefix, s.srcName...)
	} else {
		dataPrefix = append(s.srcName, dataPrefix...)
	}
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

func (s *nativeSync) Activate(immediateStart bool) {
	s.core.Activate(immediateStart)
	s.logger.Info("Sync Activated.")
}

func (s *nativeSync) Shutdown() {
	s.core.Shutdown()
	if s.isListening {
		dataPrefix := s.groupPrefix
		if s.namingScheme != BareSourceOrientedNaming {
			dataPrefix = append(dataPrefix, s.constants.DataComponent)
		}
		if s.namingScheme == GroupOrientedNaming {
			dataPrefix = append(dataPrefix, s.srcName...)
		} else {
			dataPrefix = append(s.srcName, dataPrefix...)
		}
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

func (s *nativeSync) NeedData(source enc.Name, seqno uint64) {
	i := &nativeFetchItem{
		source:  source,
		seqno:   seqno,
		retries: s.constants.DataInterestRetries,
	}
	if s.constants.MaxConcurrentDataInterests == 0 || atomic.LoadInt32(s.numFetches) < s.constants.MaxConcurrentDataInterests {
		atomic.AddInt32(s.numFetches, 1)
		s.sendInterest(i)
		return
	}
	s.fetchQueue <- i
}

func (s *nativeSync) PublishData(content []byte) {
	seqno := s.core.Seqno() + 1
	name := s.getDataName(s.srcName, seqno)
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

func (s *nativeSync) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	s.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (s *nativeSync) Core() Core {
	return s.core
}

func (s *nativeSync) sendInterest(item *nativeFetchItem) {
	wire, _, finalName, err := s.app.Spec().MakeInterest(s.getDataName(item.source, item.seqno), s.intCfg, nil, nil)
	if err != nil {
		s.logger.Errorf("Unable to make Interest: %+v", err)
		return
	}
	err = s.app.Express(finalName, s.intCfg, wire,
		func(result ndn.InterestResult, data ndn.Data, rawData, sigCovered enc.Wire, nackReason uint64) {
			if result == ndn.InterestResultData || result == ndn.InterestResultNack || item.retries == 0 {
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

func (s *nativeSync) processQueue() {
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

func (s *nativeSync) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
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

func (s *nativeSync) getDataName(source enc.Name, seqno uint64) enc.Name {
	dataName := s.groupPrefix
	if s.namingScheme != BareSourceOrientedNaming {
		dataName = append(dataName, s.constants.DataComponent)
	}
	if s.namingScheme == GroupOrientedNaming {
		dataName = append(dataName, source...)
	} else {
		dataName = append(source, dataName...)
	}
	dataName = append(dataName, enc.NewSequenceNumComponent(seqno))
	return dataName
}

func (s *nativeSync) newSourceCentricHandling(data *nativeHandlerData) {
	go func() {
		for {
			select {
			case missing, ok := <-s.missChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for _, m := range missing {
					for m.LowSeqno() <= m.HighSeqno() {
						s.NeedData(m.Source(), m.LowSeqno())
						m.Increment()
					}
				}
			}
		}
	}()
}

func (s *nativeSync) newEqualTrafficHandling(data *nativeHandlerData) {
	go func() {
		var allFetched bool
		for {
			select {
			case missing, ok := <-s.missChan:
				if !ok {
					data.done <- struct{}{}
					return
				}
				for {
					allFetched = true
					for _, m := range missing {
						if m.LowSeqno() <= m.HighSeqno() {
							s.NeedData(m.Source(), m.LowSeqno())
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
