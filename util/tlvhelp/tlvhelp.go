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

package tlvhelp

import (
	"encoding/binary"
)

func GetUintByteSize(val uint) int {
	switch {
	case val <= 0xfc:
		return 1
	case val <= 0xffff:
		return 3
	case val <= 0xffffffff:
		return 5
	default:
		return 9
	}
}

func WriteUint(val uint, buf []byte, offset int) int {
	switch {
	case val <= 0xfc:
		buf[offset] = byte(val)
		return 1
	case val <= 0xffff:
		buf[offset] = 0xfd
		binary.BigEndian.PutUint16(buf[offset+1:], uint16(val))
		return 3
	case val <= 0xffffffff:
		buf[offset] = 0xfe
		binary.BigEndian.PutUint32(buf[offset+1:], uint32(val))
		return 5
	default:
		buf[offset] = 0xff
		binary.BigEndian.PutUint64(buf[offset+1:], uint64(val))
		return 9
	}
}

func ParseUint(buf []byte, offset uint) (uint, uint) {
	switch ret := buf[offset]; {
	case ret <= 0xfc:
		return uint(ret), 1
	case ret == 0xfd:
		return uint(binary.BigEndian.Uint16(buf[offset+1 : offset+3])), 3
	case ret == 0xfe:
		return uint(binary.BigEndian.Uint32(buf[offset+1 : offset+5])), 5
	default:
		return uint(binary.BigEndian.Uint64(buf[offset+1 : offset+9])), 9
	}
}