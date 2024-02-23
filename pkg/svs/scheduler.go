package svs

import (
	"math/rand"
	"sync"
	"time"
)

type action struct {
	typ int
	val int64
}

const (
	actionStop int = iota
	actionSkip
	actionReset
	actionSet
)

type Scheduler interface {
	ApplyBounds(int64, int64)
	Start(bool)
	Stop()
	Skip()
	Reset()
	Set(time.Duration)
	TimeLeft() time.Duration
}

type scheduler struct {
	function    func()
	minInterval int64
	maxInterval int64
	actions     chan action
	timer       *time.Timer
	cycleTime   int64
	startTime   int64
	mtx         sync.Mutex
	done        chan struct{}
}

func NewScheduler(function func()) Scheduler {
	return &scheduler{
		function: function,
		actions:  make(chan action, 3),
	}
}

// Must be called before Start()
func (s *scheduler) ApplyBounds(min, max int64) {
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
func (s *scheduler) Set(v time.Duration) { s.actions <- action{typ: actionSet, val: int64(v)} }

func (s *scheduler) TimeLeft() time.Duration {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	// total interval time - time that has past = time left
	return time.Duration(s.cycleTime) -
		time.Since(time.Unix(0, s.startTime))
}

func (s *scheduler) target(execute bool) {
	defer close(s.done)
	if execute {
		s.function()
	}
	temp := BoundedRand(s.minInterval, s.maxInterval)
	s.mtx.Lock()
	s.startTime = time.Now().UnixNano()
	s.cycleTime = temp
	s.mtx.Unlock()
	s.timer = time.NewTimer(time.Duration(temp))
	for {
		select {
		case <-s.timer.C:
			s.function()
			temp = BoundedRand(s.minInterval, s.maxInterval)
			s.resetTimer(temp)
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
				temp = BoundedRand(s.minInterval, s.maxInterval)
				s.resetTimer(temp)
			case actionSet:
				s.resetTimer(a.val)
			default:
			}
		}
	}
}

func (s *scheduler) resetTimer(val int64) {
	s.mtx.Lock()
	s.startTime = time.Now().UnixNano()
	s.cycleTime = val
	s.mtx.Unlock()
	if !s.timer.Stop() {
		select {
		case <-s.timer.C:
		default:
		}
	}
	s.timer.Reset(time.Duration(val))
}

func BoundedRand(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func JitterToBounds(base time.Duration, jitter float32) (int64, int64) {
	m := base.Milliseconds()
	v := int64(float32(m) * jitter)
	return (m - v) * 1000000, (m + v) * 1000000
}
