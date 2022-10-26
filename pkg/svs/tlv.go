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
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

const (
	TypeVector     enc.TLNum = 0xc9
	TypeEntry      enc.TLNum = 0xca
	TypeEntrySeqno enc.TLNum = 0xcc
)
