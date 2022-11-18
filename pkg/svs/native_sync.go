/*
 Copyright (C) 2022-2030, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 ndn-sync is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 ndn-sync is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. If absent, it can be found within the
 GitHub repository:
          https://github.com/justincpresley/ndn-sync
*/

package svs

import (
	"sync"
	"time"

	log "github.com/apex/log"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
	utl "github.com/zjkmxy/go-ndn/pkg/utils"
)

type NativeConfig struct {
	Source       enc.Name
	GroupPrefix  enc.Name
	NamingScheme NamingScheme
	StoragePath  string
	DataCallback func(source string, seqno uint64, data ndn.Data)
	// low-level only
	UpdateCallback func(sync *NativeSync, missing []MissingData)
}

func GetBasicNativeConfig(source enc.Name, group enc.Name, callback func(source string, seqno uint64, data ndn.Data)) *NativeConfig {
	return &NativeConfig{
		Source:       source,
		GroupPrefix:  group,
		NamingScheme: HostOrientedNaming,
		StoragePath:  "./" + source.String() + "_bolt.db",
		DataCallback: callback,
	}
}

type NativeSync struct {
	app             *eng.Engine
	core            Core
	constants       *Constants
	namingScheme    NamingScheme
	groupPrefix     enc.Name
	source          enc.Name
	sourceStr       string
	storage         Database
	intCfg          *ndn.InterestConfig
	datCfg          *ndn.DataConfig
	dataComp        enc.Component
	logger          *log.Entry
	dataCall        func(source string, seqno uint64, data ndn.Data)
	updateCall      func(sync *NativeSync, missing []MissingData)
	fetchQueue      chan func() (string, uint64, uint)
	numFetches      uint
	numFetchesMutex sync.Mutex
	isListening     bool
}

func NewNativeSync(app *eng.Engine, config *NativeConfig, constants *Constants) *NativeSync {
	var s *NativeSync
	var callback func(missing []MissingData)
	logger := log.WithField("module", "svs")
	syncComp, _ := enc.ComponentFromStr("sync")
	dataComp, _ := enc.ComponentFromStr("data")
	syncPrefix := append(config.GroupPrefix, syncComp)

	if config.DataCallback == nil {
		logger.Error("Fetcher based on NativeConfig needs DataCallback.")
		return nil
	}
	if config.UpdateCallback == nil {
		callback = func(missing []MissingData) {
			var curr uint64
			for _, m := range missing {
				curr = m.LowSeqno()
				for curr <= m.HighSeqno() {
					s.FetchData(m.Source(), curr)
					curr++
				}
			}
		}
	} else {
		callback = func(missing []MissingData) {
			s.updateCall(s, missing)
		}
	}

	coreConfig := &CoreConfig{
		Source:         config.Source,
		SyncPrefix:     syncPrefix,
		UpdateCallback: callback,
	}
	storage, err := NewBoltDB(config.StoragePath, []byte("svs-packets"))
	if err != nil {
		logger.Errorf("Unable to create storage: %+v", err)
		return nil
	}
	s = &NativeSync{
		app:          app,
		core:         NewCore(app, coreConfig, constants),
		constants:    constants,
		namingScheme: config.NamingScheme,
		groupPrefix:  config.GroupPrefix,
		source:       config.Source,
		sourceStr:    config.Source.String(),
		storage:      storage,
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
		updateCall: config.UpdateCallback,
		fetchQueue: make(chan func() (string, uint64, uint), constants.InitialFetchQueueLength),
	}
	return s
}

func (s *NativeSync) Listen() {
	dataPrefix := append(s.groupPrefix, s.dataComp)
	if s.namingScheme == GroupOrientedNaming {
		dataPrefix = append(dataPrefix, s.source...)
	} else {
		dataPrefix = append(s.source, dataPrefix...)
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

func (s *NativeSync) Activate(immediateStart bool) {
	s.core.Activate(immediateStart)
	s.logger.Info("Sync Activated.")
}

func (s *NativeSync) Shutdown() {
	s.core.Shutdown()
	if s.isListening {
		dataPrefix := append(s.groupPrefix, s.dataComp)
		if s.namingScheme == GroupOrientedNaming {
			dataPrefix = append(dataPrefix, s.source...)
		} else {
			dataPrefix = append(s.source, dataPrefix...)
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
	s.logger.Info("Sync Shutdown.")
}

func (s *NativeSync) FetchData(source string, seqno uint64) {
	s.numFetchesMutex.Lock()
	if s.constants.MaxConcurrentDataInterests == 0 || s.numFetches < s.constants.MaxConcurrentDataInterests {
		s.numFetches++
		s.numFetchesMutex.Unlock()
		s.sendInterest(source, seqno, s.constants.DataInterestRetries)
		return
	}
	s.numFetchesMutex.Unlock()
	s.fetchQueue <- func() (string, uint64, uint) { return source, seqno, s.constants.DataInterestRetries }
}

func (s *NativeSync) PublishData(content []byte) {
	seqno := s.core.GetSeqno() + 1
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

func (s *NativeSync) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	s.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (s *NativeSync) GetCore() Core {
	return s.core
}

func (s *NativeSync) sendInterest(source string, seqno uint64, retries uint) {
	wire, _, finalName, err := s.app.Spec().MakeInterest(s.getDataName(source, seqno), s.intCfg, nil, nil)
	if err != nil {
		s.logger.Errorf("Unable to make Interest: %+v", err)
		return
	}
	err = s.app.Express(finalName, s.intCfg, wire,
		func(result ndn.InterestResult, data ndn.Data, rawData, sigCovered enc.Wire, nackReason uint64) {
			if result == ndn.InterestResultData || result == ndn.InterestResultNack || retries == 0 {
				s.dataCall(source, seqno, data)
				s.numFetchesMutex.Lock()
				s.numFetches--
				s.numFetchesMutex.Unlock()
				s.processQueue()
			} else {
				retries--
				s.sendInterest(source, seqno, retries)
			}
		})
	if err != nil {
		s.logger.Errorf("Unable to send Interest: %+v", err)
		return
	}
}

func (s *NativeSync) processQueue() {
	s.numFetchesMutex.Lock()
	if s.constants.MaxConcurrentDataInterests == 0 || s.numFetches < s.constants.MaxConcurrentDataInterests {
		select {
		case f := <-s.fetchQueue:
			s.numFetches++
			s.numFetchesMutex.Unlock()
			s.sendInterest(f())
			return
		default:
		}
	}
	s.numFetchesMutex.Unlock()
}

func (s *NativeSync) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
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

func (s *NativeSync) getDataName(source string, seqno uint64) enc.Name {
	dataName := append(s.groupPrefix, s.dataComp)
	src, _ := enc.NameFromStr(source)
	if s.namingScheme == GroupOrientedNaming {
		dataName = append(dataName, src...)
	} else {
		dataName = append(src, dataName...)
	}
	dataName = append(dataName, enc.NewSequenceNumComponent(seqno))
	return dataName
}
