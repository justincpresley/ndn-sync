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
	Start(bool)
	Stop()
	Skip()
	Reset()
	Set(time.Duration)
	TimeLeft() time.Duration
}

type scheduler struct {
	function   func()
	interval   time.Duration
	randomness float32
	actions    chan action
	timer      *time.Timer
	cycleTime  int64
	startTime  int64
	mtx        sync.Mutex
	done       chan struct{}
}

func NewScheduler(function func(), interval time.Duration, randomness float32) Scheduler {
	return &scheduler{
		function:   function,
		interval:   interval,
		randomness: randomness,
		actions:    make(chan action, 3),
	}
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
	temp := AddRandomness(s.interval, s.randomness)
	s.mtx.Lock()
	s.startTime = time.Now().UnixNano()
	s.cycleTime = int64(temp)
	s.mtx.Unlock()
	s.timer = time.NewTimer(temp)
	for {
		select {
		case <-s.timer.C:
			s.function()
			temp = AddRandomness(s.interval, s.randomness)
			s.mtx.Lock()
			s.startTime = time.Now().UnixNano()
			s.cycleTime = int64(temp)
			s.mtx.Unlock()
			if !s.timer.Stop() {
				select {
				case <-s.timer.C:
				default:
				}
			}
			s.timer.Reset(temp)
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
				temp = AddRandomness(s.interval, s.randomness)
				s.mtx.Lock()
				s.startTime = time.Now().UnixNano()
				s.cycleTime = int64(temp)
				s.mtx.Unlock()
				if !s.timer.Stop() {
					select {
					case <-s.timer.C:
					default:
					}
				}
				s.timer.Reset(temp)
			case actionSet:
				s.mtx.Lock()
				s.startTime = time.Now().UnixNano()
				s.cycleTime = a.val
				s.mtx.Unlock()
				if !s.timer.Stop() {
					select {
					case <-s.timer.C:
					default:
					}
				}
				s.timer.Reset(time.Duration(a.val))
			default:
			}
		}
	}
}

func AddRandomness(value time.Duration, randomness float32) time.Duration {
	r := rand.Intn(int(float32(value) * randomness))
	return value + time.Duration(r)
}
