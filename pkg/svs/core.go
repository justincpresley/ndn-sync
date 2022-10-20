/*
 Copyright (C) 2022-2025, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 This library is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 This library is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. To see more details about the authors and
 contributors, please see AUTHORS.md. If absent, Both of which can be
 found within the GitHub repository:
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

type CoreState uint8

const (
	CoreStateSteady      CoreState = 0
	CoreStateSuppression CoreState = 1
)

type CoreConfig struct {
	Source         enc.Name
	SyncPrefix     enc.Name
	UpdateCallback func([]MissingData)
}

type Core struct {
	state          CoreState
	app            *eng.Engine
	constants      *Constants
	updateCallback func([]MissingData)
	syncPrefix     enc.Name
	sourceStr      string
	vector         StateVector
	record         StateVector
	scheduler      *Scheduler
	logger         *log.Entry
	intCfg         *ndn.InterestConfig
	vectorMutex    sync.Mutex
	recordMutex    sync.Mutex
	isListening    bool
	isActive       bool
}

func NewCore(app *eng.Engine, config *CoreConfig, constants *Constants) *Core {
	c := &Core{
		state:          CoreStateSteady,
		app:            app,
		constants:      constants,
		updateCallback: config.UpdateCallback,
		syncPrefix:     config.SyncPrefix,
		sourceStr:      config.Source.String(),
		vector:         NewStateVector(),
		record:         NewStateVector(),
		logger:         log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(time.Duration(constants.SyncInterestLifeTime) * time.Millisecond),
		},
	}
	c.scheduler = NewScheduler(c.target, constants.Interval, constants.IntervalRandomness)
	return c
}

func (c *Core) Listen() {
	err := c.app.AttachHandler(c.syncPrefix, c.onInterest)
	if err != nil {
		c.logger.Errorf("Unable to register handler: %+v", err)
		return
	}
	err = c.app.RegisterRoute(c.syncPrefix)
	if err != nil {
		c.logger.Errorf("Unable to register route: %+v", err)
		return
	}
	c.isListening = true
	c.logger.Info("Sync-side Registered and Handled.")
}

func (c *Core) Activate(immediateStart bool) {
	c.scheduler.Start(immediateStart)
	c.isActive = true
	c.logger.Info("Core Activated.")
}

func (c *Core) Shutdown() {
	if c.isActive {
		c.scheduler.Stop()
	}
	if c.isListening {
		c.app.DetachHandler(c.syncPrefix)
		c.app.UnregisterRoute(c.syncPrefix)
	}
	c.logger.Info("Core Shutdown.")
}

func (c *Core) SetSeqno(seqno uint) {
	if seqno <= c.vector.Get(c.sourceStr) {
		c.logger.Warn("The Core was updated with a lower seqno.")
		return
	}
	c.vectorMutex.Lock()
	c.vector.Set(c.sourceStr, seqno)
	c.vectorMutex.Unlock()
	c.scheduler.Skip()
}

func (c *Core) GetSeqno() uint {
	return c.vector.Get(c.sourceStr)
}

func (c *Core) GetStateVector() StateVector {
	return c.vector
}

func (c *Core) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	c.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (c *Core) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	// TODO: VERIFY THE INTEREST
	incomingVector, err := ParseStateVector(interest.Name()[len(interest.Name())-2])
	if err != nil {
		c.logger.Warnf("Received unparsable statevector: %+v", err)
		return
	}
	localNewer := c.mergeStateVector(incomingVector)
	if c.recordStateVector(incomingVector) {
		return
	}
	if !localNewer {
		c.scheduler.Reset()
	} else {
		c.state = CoreStateSuppression
		delay := AddRandomness(c.constants.BriefInterval, c.constants.BriefIntervalRandomness)
		if uint(c.scheduler.TimeLeft().Milliseconds()) > delay {
			c.scheduler.Set(delay)
		}
	}
}

func (c *Core) target() {
	c.recordMutex.Lock()
	defer c.recordMutex.Unlock()
	localNewer := c.mergeStateVector(c.record)
	if c.state == CoreStateSteady || localNewer {
		c.sendInterest()
	}
	c.state = CoreStateSteady
	c.record = NewStateVector()
}

func (c *Core) sendInterest() {
	// make the interest
	// TODO: SIGN THE INTEREST WITH AUTHENTICATABLE KEY
	// WARNING: SHA SIGNER PROVIDES NOTHING (signature only includes the appParams) & IS ONLY PLACEHOLDER
	c.vectorMutex.Lock()
	name := append(c.syncPrefix, c.vector.ToComponent())
	c.vectorMutex.Unlock()
	wire, _, finalName, err := c.app.Spec().MakeInterest(
		name, c.intCfg, enc.Wire{}, sec.NewSha256IntSigner(c.app.Timer()),
	)
	if err != nil {
		c.logger.Errorf("Unable to make Sync Interest: %+v", err)
		return
	}
	// send the interest
	err = c.app.Express(finalName, c.intCfg, wire,
		func(result ndn.InterestResult, data ndn.Data, rawData, sigCovered enc.Wire, nackReason uint64) {},
	)
	if err != nil {
		c.logger.Errorf("Unable to send Sync Interest: %+v", err)
		return
	}
}

func (c *Core) mergeStateVector(incomingVector StateVector) bool {
	c.vectorMutex.Lock()
	defer c.vectorMutex.Unlock()
	var (
		missing []MissingData = make([]MissingData, 0)
		temp    uint
		isNewer bool
	)
	for nid, seqno := range incomingVector.Entries() {
		temp = c.vector.Get(nid)
		if temp < seqno {
			missing = append(missing, NewMissingData(nid, temp+1, seqno))
			c.vector.Set(nid, seqno)
		} else if nid != c.sourceStr && temp > seqno {
			isNewer = true
		}
	}
	if incomingVector.Len() < c.vector.Len() {
		isNewer = true
	}
	if len(missing) != 0 {
		c.updateCallback(missing)
	}
	return isNewer
}

func (c *Core) recordStateVector(incomingVector StateVector) bool {
	if c.state != CoreStateSuppression {
		return false
	}
	c.recordMutex.Lock()
	defer c.recordMutex.Unlock()
	for nid, seqno := range incomingVector.Entries() {
		if c.record.Get(nid) < seqno {
			c.record.Set(nid, seqno)
		}
	}
	return true
}
