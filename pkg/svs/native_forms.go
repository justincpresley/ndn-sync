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

type NativeSync interface {
	Listen()
	Activate(bool)
	Shutdown()
	FetchData(string, uint64)
	PublishData([]byte)
	FeedInterest(ndn.Interest, enc.Wire, enc.Wire, ndn.ReplyFunc, time.Time)
	GetCore() Core
}

type NativeConfig struct {
	Source       enc.Name
	GroupPrefix  enc.Name
	NamingScheme NamingScheme
	StoragePath  string
	DataCallback func(source string, seqno uint64, data ndn.Data)
	// low-level only
	UpdateCallback func(sync NativeSync, missing []MissingData)
}

func NewNativeSync(app *eng.Engine, config *NativeConfig, constants *Constants) NativeSync {
	return newNativeSync(app, config, constants)
}

func GetBasicNativeConfig(source enc.Name, group enc.Name, callback func(source string, seqno uint64, data ndn.Data)) *NativeConfig {
	return &NativeConfig{
		Source:       source,
		GroupPrefix:  group,
		NamingScheme: HostOrientedNaming,
		StoragePath:  "./" + source.String() + "_bolt.db",
		DataCallback: callback,
	}
}
