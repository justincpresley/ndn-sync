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

type FetchItem interface {
	Source() string
	Seqno() uint
}

type fetchItem struct {
	source string
	seqno  uint
}

func NewFetchItem(source string, seqno uint) FetchItem {
	return fetchItem{source: source, seqno: seqno}
}

func (fi fetchItem) Source() string {
	return fi.source
}

func (fi fetchItem) Seqno() uint {
	return fi.seqno
}
