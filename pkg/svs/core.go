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
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
	eng "github.com/zjkmxy/go-ndn/pkg/engine/basic"
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type CoreState int32

const (
	SteadyState      CoreState = 0
	SuppressionState CoreState = 1
)

type CoreConfig struct {
	Source         enc.Name
	SyncPrefix     enc.Name
	UpdateCallback func([]MissingData)
}

type Core interface {
	Listen()
	Activate(bool)
	Shutdown()
	SetSeqno(uint64)
	GetSeqno() uint64
	GetStateVector() StateVector
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
}

func NewCore(app *eng.Engine, config *CoreConfig, constants *Constants) Core {
	return newTwoStateCore(app, config, constants)
}
