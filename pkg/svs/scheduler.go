/*
 Copyright (C) 2022-2030, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 ndn-sync is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 ndn-sync is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. If absent, it can be found within the
 GitHub repository:
          https://github.com/justincpresley/ndn-sync
*/

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

func (s *scheduler) target(execute bool) {
	defer close(s.done)
	if execute {
		s.function()
	}
	temp := AddRandomness(s.interval, s.randomness)
	atomic.StoreUint64(s.cycleTime, uint64(temp))
	atomic.StoreInt64(s.startTime, time.Now().UnixNano())
	s.timer = time.NewTimer(time.Duration(temp) * time.Millisecond)
	for {
		select {
		case <-s.timer.C:
			s.function()
			temp = AddRandomness(s.interval, s.randomness)
			atomic.StoreUint64(s.cycleTime, uint64(temp))
			atomic.StoreInt64(s.startTime, time.Now().UnixNano())
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
				atomic.StoreUint64(s.cycleTime, uint64(temp))
				atomic.StoreInt64(s.startTime, time.Now().UnixNano())
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
				atomic.StoreUint64(s.cycleTime, a.val)
				atomic.StoreInt64(s.startTime, time.Now().UnixNano())
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
	elapsed := time.Now().UnixNano() - atomic.LoadInt64(s.startTime)
	return time.Duration((atomic.LoadUint64(s.cycleTime) * time.Millisecond) - elapsed)
}

func AddRandomness(value uint, randomness float32) uint {
	r := rand.Intn(int(float32(value) * randomness))
	return value + uint(r)
}
