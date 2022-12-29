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
	missingChan chan *[]MissingData
	syncPrefix  enc.Name
	sourceStr   string
	sourceSeq   uint64
	vector      StateVector
	record      StateVector
	scheduler   Scheduler
	logger      *log.Entry
	intCfg      *ndn.InterestConfig
	vectorMtx   sync.Mutex
	recordMtx   sync.Mutex
	isListening bool
	isActive    bool
}

func newTwoStateCore(app *eng.Engine, config *CoreConfig, constants *Constants) *twoStateCore {
	c := &twoStateCore{
		app:         app,
		state:       new(CoreState),
		constants:   constants,
		missingChan: make(chan *[]MissingData, constants.InitialMissingChannelSize),
		syncPrefix:  config.SyncPrefix,
		sourceStr:   config.Source.String(),
		vector:      NewStateVector(),
		record:      NewStateVector(),
		logger:      log.WithField("module", "svs"),
		intCfg: &ndn.InterestConfig{
			MustBeFresh: true,
			CanBePrefix: true,
			Lifetime:    utl.IdPtr(time.Duration(constants.SyncInterestLifeTime) * time.Millisecond),
		},
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
	if seqno <= c.sourceSeq {
		c.logger.Warn("The Core was updated with a lower seqno.")
		return
	}
	c.sourceSeq = seqno
	c.vectorMtx.Lock()
	c.vector.Set(c.sourceStr, seqno)
	c.vectorMtx.Unlock()
	c.scheduler.Skip()
}

func (c *twoStateCore) GetSeqno() uint64 {
	return c.sourceSeq
}

func (c *twoStateCore) GetStateVector() StateVector {
	return c.vector
}

func (c *twoStateCore) FeedInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	c.onInterest(interest, rawInterest, sigCovered, reply, deadline)
}

func (c *twoStateCore) onInterest(interest ndn.Interest, rawInterest enc.Wire, sigCovered enc.Wire, reply ndn.ReplyFunc, deadline time.Time) {
	// TODO: VERIFY THE INTEREST
	incomingVector, err := ParseStateVector(interest.Name()[len(interest.Name())-2])
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
		if uint(c.scheduler.TimeLeft().Milliseconds()) > delay {
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
	name := append(c.syncPrefix, c.vector.ToComponent())
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
		missing []MissingData = make([]MissingData, 0)
		temp    uint64
		isNewer bool
	)
	c.vectorMtx.Lock()
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
	c.vectorMtx.Unlock()
	if len(missing) != 0 {
		c.missingChan <- &missing
	}
	return isNewer
}

func (c *twoStateCore) recordStateVector(incomingVector StateVector) {
	c.recordMtx.Lock()
	defer c.recordMtx.Unlock()
  for nid, seqno := range incomingVector.Entries() {
		if c.record.Get(nid) < seqno {
			c.record.Set(nid, seqno)
		}
	}
}

func (c *twoStateCore) MissingChan() chan *[]MissingData {
	return c.missingChan
}
