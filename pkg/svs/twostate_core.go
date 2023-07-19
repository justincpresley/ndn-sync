package svs

import (
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
	state       *CoreState
	constants   *Constants
	missingChan chan []MissingData
	syncPrefix  enc.Name
	srcStr      string
	srcName     enc.Name
	srcSeq      uint64
	vector      StateVector
	record      StateVector
	scheduler   Scheduler
	logger      *log.Entry
	intCfg      *ndn.InterestConfig
	vectorMtx   sync.Mutex
	recordMtx   sync.Mutex
	formal      bool
	isListening bool
	isActive    bool
}

func newTwoStateCore(app *eng.Engine, config *TwoStateCoreConfig, constants *Constants) *twoStateCore {
	c := &twoStateCore{
		app:         app,
		state:       new(CoreState),
		constants:   constants,
		missingChan: make(chan []MissingData, constants.InitialMissingChannelSize),
		syncPrefix:  config.SyncPrefix,
		srcStr:      config.Source.String(),
		srcName:     config.Source,
		vector:      NewStateVector(),
		record:      NewStateVector(),
		logger:      log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(constants.SyncInterestLifeTime),
		},
		formal: config.FormalEncoding,
	}
	c.scheduler = NewScheduler(c.target, constants.Interval, constants.IntervalRandomness)
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
	close(c.missingChan)
	c.logger.Info("Core Shutdown.")
}

func (c *twoStateCore) SetSeqno(seqno uint64) {
	if seqno <= c.srcSeq {
		c.logger.Warn("The Core was updated with a lower seqno.")
		return
	}
	c.srcSeq = seqno
	c.vectorMtx.Lock()
	c.vector.Set(c.srcStr, c.srcName, seqno, false)
	c.vectorMtx.Unlock()
	c.scheduler.Skip()
}

func (c *twoStateCore) Seqno() uint64 {
	return c.srcSeq
}

func (c *twoStateCore) StateVector() StateVector {
	return c.vector
}

func (c *twoStateCore) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	c.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (c *twoStateCore) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	// TODO: VERIFY THE INTEREST
	incomingVector, err := ParseStateVector(interest.Name()[len(interest.Name())-2], c.formal)
	if err != nil {
		c.logger.Warnf("Received unparsable statevector: %+v", err)
		return
	}
	localNewer := c.mergeStateVector(incomingVector)
	if CoreState(atomic.LoadInt32((*int32)(c.state))) == Suppression {
		c.recordStateVector(incomingVector)
		return
	}
	if !localNewer {
		c.scheduler.Reset()
	} else {
		atomic.StoreInt32((*int32)(c.state), int32(Suppression))
		delay := AddRandomness(c.constants.BriefInterval, c.constants.BriefIntervalRandomness)
		if c.scheduler.TimeLeft() > delay {
			c.scheduler.Set(delay)
		}
	}
}

func (c *twoStateCore) target() {
	c.recordMtx.Lock()
	defer c.recordMtx.Unlock()
	localNewer := c.mergeStateVector(c.record)
	if CoreState(atomic.LoadInt32((*int32)(c.state))) == Steady || localNewer {
		c.sendInterest()
	}
	atomic.StoreInt32((*int32)(c.state), int32(Steady))
	c.record = NewStateVector()
}

func (c *twoStateCore) sendInterest() {
	// make the interest
	// TODO: SIGN THE INTEREST WITH AUTHENTICATABLE KEY
	// WARNING: SHA SIGNER PROVIDES NOTHING (signature only includes the appParams) & IS ONLY PLACEHOLDER
	c.vectorMtx.Lock()
	name := append(c.syncPrefix, c.vector.ToComponent(c.formal))
	c.vectorMtx.Unlock()
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

func (c *twoStateCore) mergeStateVector(incomingVector StateVector) bool {
	var (
		missing = make([]MissingData, 0)
		temp    uint64
		isNewer bool
	)
	c.vectorMtx.Lock()
	for pair := incomingVector.Entries().Back(); pair != nil; pair = pair.Prev() {
		temp = c.vector.Get(pair.Kstring)
		if temp < pair.Value {
			missing = append(missing, NewMissingData(pair.Kstring, temp+1, pair.Value))
			c.vector.Set(pair.Kstring, pair.Kname, pair.Value, false)
		} else if pair.Kstring != c.srcStr && temp > pair.Value {
			isNewer = true
		}
	}
	if incomingVector.Len() < c.vector.Len() {
		isNewer = true
	}
	c.vectorMtx.Unlock()
	if len(missing) != 0 {
		c.missingChan <- missing
	}
	return isNewer
}

func (c *twoStateCore) recordStateVector(incomingVector StateVector) {
	c.recordMtx.Lock()
	defer c.recordMtx.Unlock()
	for pair := incomingVector.Entries().Back(); pair != nil; pair = pair.Prev() {
		if c.record.Get(pair.Kstring) < pair.Value {
			c.record.Set(pair.Kstring, pair.Kname, pair.Value, false)
		}
	}
}

func (c *twoStateCore) Chan() chan []MissingData {
	return c.missingChan
}
