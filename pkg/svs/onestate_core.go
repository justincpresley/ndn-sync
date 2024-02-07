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

type oneStateCore struct {
	app         *eng.Engine
	constants   *Constants
	subs        []chan SyncUpdate
	syncPrefix  enc.Name
	srcStr      string
	srcName     enc.Name
	srcSeq      uint64
	local       StateVector
	scheduler   Scheduler
	logger      *log.Entry
	intCfg      *ndn.InterestConfig
	localMtx    sync.Mutex
	formal      bool
	isListening bool
	isActive    bool
}

func newOneStateCore(app *eng.Engine, config *OneStateCoreConfig, constants *Constants) *oneStateCore {
	c := &oneStateCore{
		app:        app,
		constants:  constants,
		subs:       make([]chan SyncUpdate, 0),
		syncPrefix: config.SyncPrefix,
		srcStr:     config.Source.String(),
		srcName:    config.Source,
		local:      NewStateVector(),
		logger:     log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(constants.SyncInterestLifeTime),
		},
		formal: config.FormalEncoding,
	}
	c.scheduler = NewScheduler(c.sendInterest, constants.Interval, constants.IntervalRandomness)
	return c
}

func (c *oneStateCore) Listen() {
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

func (c *oneStateCore) Activate(immediateStart bool) {
	c.scheduler.Start(immediateStart)
	c.isActive = true
	c.logger.Info("Core Activated.")
}

func (c *oneStateCore) Shutdown() {
	if c.isActive {
		c.scheduler.Stop()
	}
	if c.isListening {
		err := c.app.DetachHandler(c.syncPrefix)
		if err != nil {
			c.logger.Errorf("Detech handler error: %+v", err)
		}
		err = c.app.UnregisterRoute(c.syncPrefix)
		if err != nil {
			c.logger.Errorf("Unregister route error: %+v", err)
		}
	}
	for _, sub := range c.subs {
		close(sub)
	}
	c.logger.Info("Core Shutdown.")
}

func (c *oneStateCore) SetSeqno(seqno uint64) {
	if seqno <= c.srcSeq {
		c.logger.Warn("The Core was updated with a lower seqno.")
		return
	}
	c.srcSeq = seqno
	c.localMtx.Lock()
	c.local.Set(c.srcStr, c.srcName, seqno, false)
	c.localMtx.Unlock()
	c.scheduler.Skip()
}

func (c *oneStateCore) Seqno() uint64 {
	return c.srcSeq
}

func (c *oneStateCore) StateVector() StateVector {
	return c.local
}

func (c *oneStateCore) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	c.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (c *oneStateCore) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	// TODO: VERIFY THE INTEREST
	remote, err := ParseStateVector(enc.NewWireReader(interest.AppParam()), c.formal)
	if err != nil {
		c.logger.Warnf("Received unparsable statevector: %+v", err)
		return
	}
	localNewer := c.mergeVectorToLocal(remote)
	if !localNewer {
		c.scheduler.Reset()
	} else {
		delay := AddRandomness(c.constants.BriefInterval, c.constants.BriefIntervalRandomness)
		if c.scheduler.TimeLeft() > delay {
			c.scheduler.Set(delay)
		}
	}
}

func (c *oneStateCore) sendInterest() {
	// make the interest
	// TODO: SIGN THE INTEREST WITH AUTHENTICATABLE KEY
	// WARNING: SHA SIGNER PROVIDES NOTHING (signature only includes the appParams) & IS ONLY PLACEHOLDER
	c.localMtx.Lock()
	appP := c.local.Encode(c.formal)
	c.localMtx.Unlock()
	wire, _, finalName, err := c.app.Spec().MakeInterest(
		c.syncPrefix, c.intCfg, appP, sec.NewSha256IntSigner(c.app.Timer()),
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

func (c *oneStateCore) mergeVectorToLocal(vector StateVector) bool {
	var (
		missing = make(SyncUpdate, 0)
		temp    uint64
		isNewer bool
	)
	c.localMtx.Lock()
	for pair := vector.Entries().Back(); pair != nil; pair = pair.Prev() {
		temp = c.local.Get(pair.Kstring)
		if temp < pair.Value {
			missing = append(missing, NewMissingData(pair.Kname, temp+1, pair.Value))
			c.local.Set(pair.Kstring, pair.Kname, pair.Value, false)
		} else if pair.Kstring != c.srcStr && temp > pair.Value {
			isNewer = true
		}
	}
	if vector.Len() < c.local.Len() {
		isNewer = true
	}
	c.localMtx.Unlock()
	if len(missing) != 0 {
		for _, sub := range c.subs {
			sub <- missing
		}
	}
	return isNewer
}

func (c *oneStateCore) Subscribe() chan SyncUpdate {
	ch := make(chan SyncUpdate, c.constants.InitialMissingChannelSize)
	c.subs = append(c.subs, ch)
	return ch
}