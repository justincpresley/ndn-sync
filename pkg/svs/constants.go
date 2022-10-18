/*
 Copyright (C) 2022-2025, The ndn-sync Go Library Authors

 This file is part of ndn-sync: An NDN Go Library for Sync Protocols.

 This library is free software; you can redistribute it and/or
 modify it under the terms of the GNU Lesser General Public
 License as published by the Free Software Foundation; either
 version 2.1 of the License, or any later version.

 This library is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
 See the GNU Lesser General Public License for more details.

 A copy of the GNU Lesser General Public License is provided by this
 library under LICENSE.md. To see more details about the authors and
 contributors, please see AUTHORS.md. If absent, Both of which can be
 found within the GitHub repository:
          https://github.com/justincpresley/ndn-sync
*/

package svs

type Constants struct {
	Interval                   uint    // (ms) milliseconds
	IntervalRandomness         float32 // percentage variance 0.00<=x<=1.00
	BriefInterval              uint    // (ms) milliseconds
	BriefIntervalRandomness    float32 // percentage variance 0.00<=x<=1.00
	DataInterestLifeTime       uint    // (ms) milliseconds
	DataPacketFressness        uint    // (ms) milliseconds
	SyncInterestLifeTime       uint    // (ms) milliseconds
	MaxConcurrentDataInterests uint    // 0 = inf
}

func GetDefaultConstants() *Constants {
	return &Constants{
		Interval:                   30000,
		IntervalRandomness:         0.10,
		BriefInterval:              200,
		BriefIntervalRandomness:    0.50,
		DataInterestLifeTime:       2000,
		DataPacketFressness:        5000,
		SyncInterestLifeTime:       1000,
		MaxConcurrentDataInterests: 10,
	}
}
