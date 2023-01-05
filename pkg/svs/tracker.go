package svs

import (
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
	entries   map[string]*heart
	constants *Constants
	statChan  chan StatusChange
	selfHrt   *heart
	selfSrc   string
}

func NewTracker(self string, cs *Constants) Tracker {
	s := &tracker{
		entries:   make(map[string]*heart),
		constants: cs,
		statChan:  make(chan StatusChange, cs.InitialStatusChangeChannelSize),
		selfSrc:   self,
		selfHrt:   &heart{status: Expired},
	}
	s.statChan <- NewStatusChange(self, Unseen, Expired)
	return s
}

func (t *tracker) Reset(src string) {
	if src == t.selfSrc {
		t.resetHeart(t.selfSrc, t.selfHrt)
		return
	}
	hrt, exists := t.entries[src]
	if !exists {
		t.statChan <- NewStatusChange(src, Unseen, Expired)
		t.entries[src] = &heart{status: Expired}
		return
	}
	t.resetHeart(src, hrt)
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
		t.statChan <- NewStatusChange(src, Expired, Renewed)
	}
}

func (t *tracker) Detect() {
	currentTime := time.Now()
	var tp int64
	for src, hrt := range t.entries {
		tp = currentTime.Sub(hrt.lastBeat).Milliseconds()
		if tp > t.constants.TrackRate {
			if hrt.status != Renewed {
				hrt.beats = 0
			} else {
				hrt.beats = uint(tp / t.constants.TrackRate)
				if hrt.beats >= t.constants.HeartbeatsToExpire {
					hrt.beats = 0
					hrt.status = Expired
					t.statChan <- NewStatusChange(src, Renewed, Expired)
				}
			}
		}
	}
}

func (t *tracker) UntilBeat() time.Duration {
	return time.Until(t.selfHrt.lastBeat.Add(time.Duration(t.constants.HeartbeatRate) * time.Millisecond))
}

func (t *tracker) Status(src string) Status {
	if src == t.selfSrc {
		return t.selfHrt.status
	}
	hrt, ok := t.entries[src]
	if !ok {
		return Unseen
	}
	return hrt.status
}

func (t *tracker) Chan() chan StatusChange {
	return t.statChan
}
