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

type MissingData interface {
	Source() string
	LowSeqno() uint64
	HighSeqno() uint64
}

type missingData struct {
	source    string
	lowSeqno  uint64
	highSeqno uint64
}

func NewMissingData(source string, low uint64, high uint64) MissingData {
	return missingData{source: source, lowSeqno: low, highSeqno: high}
}

func (md missingData) Source() string {
	return md.source
}

func (md missingData) LowSeqno() uint64 {
	return md.lowSeqno
}

func (md missingData) HighSeqno() uint64 {
	return md.highSeqno
}
