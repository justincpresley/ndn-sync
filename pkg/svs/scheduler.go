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
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type Scheduler struct {
	function   func()
	interval   uint
	randomness float32
	quit       chan struct{}
	cycle      chan time.Duration
	startTime  time.Time
	cycleTime  time.Duration
	timer      *time.Timer
}

func NewScheduler(function func(), interval uint, randomness float32) *Scheduler {
	return &Scheduler{
		function:   function,
		interval:   interval,
		randomness: randomness,
		quit:       make(chan struct{}),
		cycle:      make(chan time.Duration),
	}
}

func (s *Scheduler) Start(executeNow bool) {
	go s.target(executeNow)
}

func (s *Scheduler) target(executeNow bool) {
	var temp time.Duration
	if executeNow {
		s.function()
	}
	s.cycleTime = time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
	s.startTime = time.Now()
	s.timer = time.NewTimer(s.cycleTime)
	for {
		select {
		case <-s.timer.C:
			s.function()
			s.cycleTime = time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
			s.startTime = time.Now()
			s.timer.Reset(s.cycleTime)
		case temp = <-s.cycle:
			s.cycleTime = temp
			s.startTime = time.Now()
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

func (s *Scheduler) Stop() {
	s.quit <- struct{}{}
}

func (s *Scheduler) Skip() {
	s.cycle <- time.Duration(0)
}

func (s *Scheduler) Reset() {
	s.cycle <- time.Duration(AddRandomness(s.interval, s.randomness)) * time.Millisecond
}

func (s *Scheduler) Set(value uint) {
	s.cycle <- time.Duration(value) * time.Millisecond
}

func (s *Scheduler) Add(value uint) {
	s.cycle <- (time.Duration(value) * time.Millisecond) + s.cycleTime
}

func (s *Scheduler) TimeLeft() time.Duration {
	return time.Until(s.startTime.Add(s.cycleTime))
}

func AddRandomness(value uint, randomness float32) uint {
	i := uint(float32(value) * randomness)
	min := int(value - i)
	max := int(value + i)
	return uint(rand.Intn(max-min) + min)
}
