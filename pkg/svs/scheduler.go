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
	"sync"
	"time"
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
	quit       chan struct{}
	cycle      chan time.Duration
	startTime  time.Time
	cycleTime  time.Duration
	cycleMutex sync.RWMutex
	timer      *time.Timer
}

func NewScheduler(function func(), interval uint, randomness float32) Scheduler {
	return &scheduler{
		function:   function,
		interval:   interval,
		randomness: randomness,
		quit:       make(chan struct{}),
		cycle:      make(chan time.Duration),
	}
}

func (s *scheduler) Start(executeNow bool) {
	go s.target(executeNow)
}

func (s *scheduler) target(executeNow bool) {
	var temp time.Duration
	if executeNow {
		s.function()
	}
	s.cycleMutex.Lock()
	s.cycleTime = time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
	s.startTime = time.Now()
	s.timer = time.NewTimer(s.cycleTime)
	s.cycleMutex.Unlock()
	for {
		select {
		case <-s.timer.C:
			s.function()
			s.cycleMutex.Lock()
			s.cycleTime = time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
			s.startTime = time.Now()
			s.timer.Reset(s.cycleTime)
			s.cycleMutex.Unlock()
		case temp = <-s.cycle:
			s.cycleMutex.Lock()
			s.cycleTime = temp
			s.startTime = time.Now()
			s.cycleMutex.Unlock()
			if !s.timer.Stop() {
				select {
				case <-s.timer.C:
				default:
				}
			}
			s.timer.Reset(temp)
		case <-s.quit:
			if !s.timer.Stop() {
				select {
				case <-s.timer.C:
				default:
				}
			}
			return
		}
	}
}

func (s *scheduler) Stop() {
	s.quit <- struct{}{}
}

func (s *scheduler) Skip() {
	s.cycle <- time.Duration(0)
}

func (s *scheduler) Reset() {
	s.cycle <- time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
}

func (s *scheduler) Set(value uint) {
	s.cycle <- time.Duration(value) * time.Millisecond
}

func (s *scheduler) Add(value uint) {
	s.cycleMutex.RLock()
	defer s.cycleMutex.RUnlock()
	s.cycle <- (time.Duration(value) * time.Millisecond) + s.cycleTime
}

func (s *scheduler) TimeLeft() time.Duration {
	s.cycleMutex.RLock()
	defer s.cycleMutex.RUnlock()
	return time.Until(s.startTime.Add(s.cycleTime))
}

func AddRandomness(value uint, randomness float32) uint {
	i := uint(float32(value) * randomness)
	min := int(value - i)
	max := int(value + i)
	return uint(rand.Intn(max-min) + min)
}
