package svs

import (
	"slices"
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
	selfsets    []string
	local       *StateVector
	scheduler   Scheduler
	logger      *log.Entry
	intCfg      *ndn.InterestConfig
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
		selfsets:   make([]string, 0),
		local:      NewStateVector(),
		logger:     log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(constants.SyncInterestLifeTime),
		},
		formal: config.FormalEncoding,
	}
	c.scheduler = NewScheduler(c.sendInterest)
	c.scheduler.ApplyBounds(JitterToBounds(constants.SyncInterval, constants.SyncIntervalJitter))
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

func (c *oneStateCore) Update(dsname enc.Name, seqno uint64) {
	if seqno == 0 {
		c.logger.Warn("The Core was updated with a seqno of 0.")
		return
	}
	dsstr := dsname.String()
	if seqno <= c.local.Get(dsstr) {
		c.logger.Warn("The Core was updated with a non-new seqno.")
		return
	}
	if c.local.Get(dsstr) == 0 {
		c.selfsets = append(c.selfsets, dsstr)
	} else {
		if !slices.Contains(c.selfsets, dsstr) {
			c.logger.Warn("The Core was updated with a dataset not previously updated by the node.")
			return
		}
	}
	c.local.Lock()
	c.local.Update(dsstr, dsname, seqno, false)
	c.local.Unlock()
	c.scheduler.Skip()
}

func (c *oneStateCore) Subscribe() chan SyncUpdate {
	ch := make(chan SyncUpdate, c.constants.InitialMissingChannelSize)
	c.subs = append(c.subs, ch)
	return ch
}

func (c *oneStateCore) StateVector() *StateVector {
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
		c.scheduler.Skip()
	}
}

func (c *oneStateCore) sendInterest() {
	// make the interest
	// TODO: SIGN THE INTEREST WITH AUTHENTICATABLE KEY
	// WARNING: SHA SIGNER PROVIDES NOTHING (signature only includes the appParams) & IS ONLY PLACEHOLDER
	c.local.RLock()
	appP := c.local.Encode(c.formal)
	c.local.RUnlock()
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

func (c *oneStateCore) mergeVectorToLocal(vector *StateVector) bool {
	var (
		missing = make(SyncUpdate, 0)
		lVal    uint64
		lNewer  bool
	)
	c.local.Lock()
	for p := vector.Entries().Back(); p != nil; p = p.Prev() {
		lVal = c.local.Get(p.Kstr)
		if lVal < p.Val {
			missing = append(missing, MissingData{Dataset: p.Kname, StartSeq: lVal + 1, EndSeq: p.Val})
			c.local.Set(p.Kstr, p.Kname, p.Val, false)
		} else if lVal > p.Val {
			if slices.Contains(c.selfsets, p.Kstr) && time.Since(c.local.LastUpdated(p.Kstr)) < c.constants.SuppressionInterval {
				continue
			}
			lNewer = true
		}
	}
	if vector.Len() < c.local.Len() {
		lNewer = true
	}
	c.local.Unlock()
	if len(missing) != 0 {
		for _, sub := range c.subs {
			sub <- missing
		}
	}
	return lNewer
}
