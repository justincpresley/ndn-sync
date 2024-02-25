package svs

import (
	"math/rand/v2"
	"sync"
	"time"
)

type action struct {
	typ int
	val time.Duration
}

const (
	actionStop int = iota
	actionSkip
	actionReset
	actionSet
)

type Scheduler interface {
	ApplyBounds(time.Duration, time.Duration)
	Start(bool)
	Stop()
	Skip()
	Reset()
	Set(time.Duration)
	TimeLeft() time.Duration
}

type scheduler struct {
	function    func()
	actions     chan action
	timer       *time.Timer
	minInterval time.Duration
	maxInterval time.Duration
	cycleLength time.Duration
	startTime   time.Time
	mtx         sync.RWMutex
	done        chan struct{}
}

func NewScheduler(function func()) Scheduler {
	return &scheduler{
		function: function,
		actions:  make(chan action, 3),
	}
}

// Must be called before Start()
func (s *scheduler) ApplyBounds(min, max time.Duration) {
	s.minInterval = min
	s.maxInterval = max
}

func (s *scheduler) Start(execute bool) {
	s.done = make(chan struct{})
	go s.target(execute)
}

func (s *scheduler) Stop() {
	s.actions <- action{typ: actionStop}
	<-s.done
}

func (s *scheduler) Skip()               { s.actions <- action{typ: actionSkip} }
func (s *scheduler) Reset()              { s.actions <- action{typ: actionReset} }
func (s *scheduler) Set(v time.Duration) { s.actions <- action{typ: actionSet, val: v} }

func (s *scheduler) TimeLeft() time.Duration {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return time.Duration(s.cycleLength) - time.Since(s.startTime)
}

func (s *scheduler) target(execute bool) {
	defer close(s.done)
	if execute {
		s.function()
	}
	s.newTimer()
	for {
		select {
		case <-s.timer.C:
			s.function()
			s.resetTimer(BoundedRand(s.minInterval, s.maxInterval))
		case a := <-s.actions:
			switch a.typ {
			case actionStop:
				if !s.timer.Stop() {
					select {
					case <-s.timer.C:
					default:
					}
				}
				return
			case actionSkip:
				s.function()
				fallthrough
			case actionReset:
				s.resetTimer(BoundedRand(s.minInterval, s.maxInterval))
			case actionSet:
				s.resetTimer(a.val)
			default:
			}
		}
	}
}

func (s *scheduler) newTimer() {
	r := BoundedRand(s.minInterval, s.maxInterval)
	s.mtx.Lock()
	s.startTime = time.Now()
	s.cycleLength = r
	s.mtx.Unlock()
	s.timer = time.NewTimer(r)
}

func (s *scheduler) resetTimer(val time.Duration) {
	s.mtx.Lock()
	s.startTime = time.Now()
	s.cycleLength = val
	s.mtx.Unlock()
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.timer.Reset(val)
}

func BoundedRand(min, max time.Duration) time.Duration {
	return min + rand.N(max-min+1)
}

func JitterToBounds(base time.Duration, jitter float64) (time.Duration, time.Duration) {
	v := time.Duration(float64(base) * jitter)
	return base - v, base + v
}
