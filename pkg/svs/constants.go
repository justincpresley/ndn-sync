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

type Constants struct {
	IntervalRandomness         float32 // percentage variance 0.00<=x<=1.00
	BriefIntervalRandomness    float32 // percentage variance 0.00<=x<=1.00
	Interval                   uint    // (ms) milliseconds
	BriefInterval              uint    // (ms) milliseconds
	DataInterestLifeTime       uint    // (ms) milliseconds
	DataInterestRetries        uint    // 0 = no retry
	DataPacketFressness        uint    // (ms) milliseconds
	SyncInterestLifeTime       uint    // (ms) milliseconds
	MaxConcurrentDataInterests int32   // 0 = inf
	InitialFetchQueueLength    uint    // only helps to mitigate allocation resizing
	InitialMissingChannelSize  uint    // only helps to mitigate allocation resizing
}

func GetDefaultConstants() *Constants {
	return &Constants{
		IntervalRandomness:         0.10,
		BriefIntervalRandomness:    0.50,
		Interval:                   30000,
		BriefInterval:              200,
		DataInterestLifeTime:       2000,
		DataInterestRetries:        2,
		DataPacketFressness:        5000,
		SyncInterestLifeTime:       1000,
		MaxConcurrentDataInterests: 10,
		InitialFetchQueueLength:    50,
		InitialMissingChannelSize:  5,
	}
}
