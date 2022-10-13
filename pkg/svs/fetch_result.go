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

import (
	ndn "github.com/zjkmxy/go-ndn/pkg/ndn"
)

type FetchResult interface {
	Source() string
	Seqno() uint
	Data() ndn.Data
}

type fetchResult struct {
	source string
	seqno  uint
	data   ndn.Data
}

func NewFetchResult(source string, seqno uint, data ndn.Data) FetchResult {
	return fetchResult{source: source, seqno: seqno, data: data}
}

func (fr fetchResult) Source() string {
	return fr.source
}

func (fr fetchResult) Seqno() uint {
	return fr.seqno
}

func (fr fetchResult) Data() ndn.Data {
	return fr.data
}
