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
	LowSeqno() uint
	HighSeqno() uint
}

type missingData struct {
	source    string
	lowSeqno  uint
	highSeqno uint
}

func NewMissingData(source string, low uint, high uint) MissingData {
	return missingData{source: source, lowSeqno: low, highSeqno: high}
}

func (md missingData) Source() string {
	return md.source
}

func (md missingData) LowSeqno() uint {
	return md.lowSeqno
}

func (md missingData) HighSeqno() uint {
	return md.highSeqno
}
