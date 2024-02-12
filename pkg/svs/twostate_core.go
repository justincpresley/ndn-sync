package svs

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/apex/log"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
	sec "github.com/zjkmxy/go-ndn/pkg/security"
	utl "github.com/zjkmxy/go-ndn/pkg/utils"
)

type twoStateCore struct {
	app         *eng.Engine
	state       *int32
	constants   *Constants
	subs        []chan SyncUpdate
	syncPrefix  enc.Name
	selfsets    []string
	updateTimes map[string]time.Time
	local       StateVector
	record      StateVector
	scheduler   Scheduler
	logger      *log.Entry
	intCfg      *ndn.InterestConfig
	localMtx    sync.Mutex
	recordMtx   sync.Mutex
	formal      bool
	effSuppress bool
	isListening bool
	isActive    bool
}

func newTwoStateCore(app *eng.Engine, config *TwoStateCoreConfig, constants *Constants) *twoStateCore {
	c := &twoStateCore{
		app:         app,
		state:       new(int32),
		constants:   constants,
		subs:        make([]chan SyncUpdate, 0),
		syncPrefix:  config.SyncPrefix,
		selfsets:    make([]string, 0),
		updateTimes: make(map[string]time.Time),
		local:       NewStateVector(),
		record:      NewStateVector(),
		logger:      log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(constants.SyncInterestLifeTime),
		},
		formal:      config.FormalEncoding,
		effSuppress: config.EfficientSuppression,
	}
	c.scheduler = NewScheduler(c.onTimer, constants.Interval, constants.IntervalRandomness)
	return c
}

func (c *twoStateCore) Listen() {
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

func (c *twoStateCore) Activate(immediateStart bool) {
	c.scheduler.Start(immediateStart)
	c.isActive = true
	c.logger.Info("Core Activated.")
}

func (c *twoStateCore) Shutdown() {
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

func (c *twoStateCore) Update(dsname enc.Name, seqno uint64) {
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
	c.localMtx.Lock()
	c.local.Set(dsstr, dsname, seqno, false)
	c.localMtx.Unlock()
	c.updateTimes[dsstr] = time.Now()
	c.scheduler.Skip()
}

func (c *twoStateCore) StateVector() StateVector {
	return c.local
}

func (c *twoStateCore) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	c.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (c *twoStateCore) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	// TODO: VERIFY THE INTEREST
	remote, err := ParseStateVector(enc.NewWireReader(interest.AppParam()), c.formal)
	if err != nil {
		c.logger.Warnf("Received unparsable statevector: %+v", err)
		return
	}
	if atomic.LoadInt32(c.state) == suppressionState {
		c.recordVector(remote)
		return
	}
	localNewer := c.mergeVectorToLocal(remote)
	if !localNewer {
		c.scheduler.Reset()
	} else {
		atomic.StoreInt32(c.state, suppressionState)
		delay := AddRandomness(c.constants.BriefInterval, c.constants.BriefIntervalRandomness)
		if c.scheduler.TimeLeft() > delay {
			c.scheduler.Set(delay)
		}
	}
}

func (c *twoStateCore) onTimer() {
	if atomic.LoadInt32(c.state) == steadyState {
		c.sendInterest()
	} else {
		c.recordMtx.Lock()
		localNewer := c.mergeRecordToLocal()
		c.record = NewStateVector()
		c.recordMtx.Unlock()
		atomic.StoreInt32(c.state, steadyState)
		if localNewer {
			c.sendInterest()
		}
	}
}

func (c *twoStateCore) sendInterest() {
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

func (c *twoStateCore) mergeVectorToLocal(vector StateVector) bool {
	var (
		missing = make(SyncUpdate, 0)
		lVal    uint64
		isNewer bool
	)
	c.localMtx.Lock()
	for p := vector.Entries().Back(); p != nil; p = p.Prev() {
		lVal = c.local.Get(p.Kstr)
		if lVal < p.Val {
			missing = append(missing, NewMissingData(p.Kname, lVal+1, p.Val))
			c.local.Set(p.Kstr, p.Kname, p.Val, false)
			c.updateTimes[p.Kstr] = time.Now()
		} else if !slices.Contains(c.selfsets, p.Kstr) && lVal > p.Val {
			if !c.effSuppress || time.Since(c.updateTimes[p.Kstr]) >= c.constants.BriefInterval {
				isNewer = true
			}
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

func (c *twoStateCore) recordVector(vector StateVector) {
	var (
		missing = make(SyncUpdate, 0)
		lVal    uint64
	)
	c.recordMtx.Lock()
	for p := vector.Entries().Back(); p != nil; p = p.Prev() {
		lVal = c.local.Get(p.Kstr)
		if c.record.Get(p.Kstr) < p.Val {
			c.record.Set(p.Kstr, p.Kname, p.Val, true)
		}
		if lVal < p.Val {
			missing = append(missing, NewMissingData(p.Kname, lVal+1, p.Val))
			c.local.Set(p.Kstr, p.Kname, p.Val, false)
			c.updateTimes[p.Kstr] = time.Now()
		}
	}
	c.recordMtx.Unlock()
	if len(missing) != 0 {
		for _, sub := range c.subs {
			sub <- missing
		}
	}
}

func (c *twoStateCore) mergeRecordToLocal() bool {
	var isNewer bool
	c.localMtx.Lock()
	for p := c.record.Entries().Back(); p != nil; p = p.Prev() {
		if c.local.Get(p.Kstr) > p.Val && !slices.Contains(c.selfsets, p.Kstr) {
			if !c.effSuppress || time.Since(c.updateTimes[p.Kstr]) >= c.constants.BriefInterval {
				isNewer = true
			}
		}
	}
	if c.record.Len() < c.local.Len() {
		isNewer = true
	}
	c.localMtx.Unlock()
	return isNewer
}

func (c *twoStateCore) Subscribe() chan SyncUpdate {
	ch := make(chan SyncUpdate, c.constants.InitialMissingChannelSize)
	c.subs = append(c.subs, ch)
	return ch
}
