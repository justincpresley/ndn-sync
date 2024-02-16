package svs

import (
	"sync"
	"time"
)

type Tracker interface {
	Reset(string)
	Detect()
	Status(string) Status
	UntilBeat() time.Duration
	Chan() chan StatusChange
}

type heart struct {
	lastBeat time.Time
	status   Status
	beats    uint
}

type tracker struct {
	entries   sync.Map
	constants *Constants
	statChan  chan StatusChange
	selfHrt   *heart
}

func NewTracker(src string, cs *Constants) Tracker {
	t := &tracker{
		constants: cs,
		statChan:  make(chan StatusChange, cs.InitialStatusChangeChannelSize),
	}
	hrt := &heart{status: Expired}
	t.entries.Store(src, hrt)
	t.selfHrt = hrt
	t.statChan <- StatusChange{Node: src, OldStatus: Unseen, NewStatus: Expired}
	return t
}

func (t *tracker) Reset(src string) {
	hrt, ok := t.entries.Load(src)
	if !ok {
		t.entries.Store(src, &heart{status: Expired})
		t.statChan <- StatusChange{Node: src, OldStatus: Unseen, NewStatus: Expired}
		return
	}
	t.resetHeart(src, hrt.(*heart))
}

func (t *tracker) resetHeart(src string, hrt *heart) {
	hrt.lastBeat = time.Now()
	if hrt.status == Renewed {
		hrt.beats = 0
		return
	}
	hrt.beats++
	if hrt.beats >= t.constants.HeartbeatsToRenew {
		hrt.beats = 0
		hrt.status = Renewed
		t.statChan <- StatusChange{Node: src, OldStatus: Expired, NewStatus: Renewed}
	}
}

func (t *tracker) Detect() {
	var (
		currentTime = time.Now()
		tp          time.Duration
		src         string
		hrt         *heart
	)
	t.entries.Range(func(key, value any) bool {
		src = key.(string)
		hrt = value.(*heart)
		if hrt == t.selfHrt {
			return true
		}
		tp = currentTime.Sub(hrt.lastBeat)
		if tp > t.constants.TrackRate {
			if hrt.status != Renewed {
				hrt.beats = 0
			} else {
				hrt.beats = uint(tp / t.constants.TrackRate)
				if hrt.beats >= t.constants.HeartbeatsToExpire {
					hrt.beats = 0
					hrt.status = Expired
					t.statChan <- StatusChange{Node: src, OldStatus: Renewed, NewStatus: Expired}
				}
			}
		}
		return true
	})
}

func (t *tracker) UntilBeat() time.Duration {
	return time.Until(t.selfHrt.lastBeat.Add(t.constants.HeartbeatRate))
}

func (t *tracker) Status(src string) Status {
	hrt, ok := t.entries.Load(src)
	if !ok {
		return Unseen
	}
	return hrt.(*heart).status
}

func (t *tracker) Chan() chan StatusChange {
	return t.statChan
}
