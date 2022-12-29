package svs

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type action struct {
	typ int
	val uint64
}

const (
	actionStop int = iota
	actionSkip
	actionReset
	actionSet
	actionAdd
)

type Scheduler interface {
	Start(bool)
	Stop()
	Skip()
	Reset()
	Set(uint)
	Add(uint)
	TimeLeft() time.Duration
}

type scheduler struct {
	function   func()
	interval   uint
	randomness float32
	actions    chan action
	timer      *time.Timer
	cycleTime  *uint64
	startTime  *int64
	done       chan struct{}
}

func NewScheduler(function func(), interval uint, randomness float32) Scheduler {
	return &scheduler{
		function:   function,
		interval:   interval,
		randomness: randomness,
		actions:    make(chan action, 3),
		cycleTime:  new(uint64),
		startTime:  new(int64),
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

func (s *scheduler) Skip()      { s.actions <- action{typ: actionSkip} }
func (s *scheduler) Reset()     { s.actions <- action{typ: actionReset} }
func (s *scheduler) Set(v uint) { s.actions <- action{typ: actionSet, val: uint64(v)} }
func (s *scheduler) Add(v uint) { s.actions <- action{typ: actionAdd, val: uint64(v)} }

func (s *scheduler) TimeLeft() time.Duration {
	// total interval time - time that has past = time left
	return (time.Duration(atomic.LoadUint64(s.cycleTime)) * time.Millisecond) -
		time.Since(time.Unix(0, atomic.LoadInt64(s.startTime)))
}

func (s *scheduler) target(execute bool) {
	defer close(s.done)
	if execute {
		s.function()
	}
	temp := AddRandomness(s.interval, s.randomness)
	atomic.StoreInt64(s.startTime, time.Now().UnixNano())
	atomic.StoreUint64(s.cycleTime, uint64(temp))
	s.timer = time.NewTimer(time.Duration(temp) * time.Millisecond)
	for {
		select {
		case <-s.timer.C:
			s.function()
			temp = AddRandomness(s.interval, s.randomness)
			atomic.StoreInt64(s.startTime, time.Now().UnixNano())
			atomic.StoreUint64(s.cycleTime, uint64(temp))
			if !s.timer.Stop() {
				select {
				case <-s.timer.C:
				default:
				}
			}
			s.timer.Reset(time.Duration(temp) * time.Millisecond)
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
				atomic.StoreInt64(s.startTime, time.Now().UnixNano())
				atomic.StoreUint64(s.cycleTime, uint64(temp))
				if !s.timer.Stop() {
					select {
					case <-s.timer.C:
					default:
					}
				}
				s.timer.Reset(time.Duration(temp) * time.Millisecond)
			case actionAdd:
				a.val += atomic.LoadUint64(s.cycleTime)
				fallthrough
			case actionSet:
				atomic.StoreInt64(s.startTime, time.Now().UnixNano())
				atomic.StoreUint64(s.cycleTime, a.val)
				if !s.timer.Stop() {
					select {
					case <-s.timer.C:
					default:
					}
				}
				s.timer.Reset(time.Duration(a.val) * time.Millisecond)
			default:
			}
		}
	}
}

func AddRandomness(value uint, randomness float32) uint {
	r := rand.Intn(int(float32(value) * randomness))
	return value + uint(r)
}
