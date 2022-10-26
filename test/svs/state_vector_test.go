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
	"testing"

	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestStateVectorBasics(t *testing.T) {
	sv := svs.NewStateVector()
	assert.Equal(t, uint(0), sv.Get("/node1"))
	sv.Set("/node1", 60)
	sv.Set("/node2", 9)
	sv.Set("/node1", 62)
	sv.Set("/node3", 1)
	assert.Equal(t, uint(62), sv.Get("/node1"))
	assert.Equal(t, uint(9), sv.Get("/node2"))
	assert.Equal(t, uint(1), sv.Get("/node3"))
	assert.Equal(t, int(3), sv.Len())
}

func TestStateVectorLoop(t *testing.T) {
	sv := svs.NewStateVector()
	nsv := svs.NewStateVector()
	sv.Set("/node1", 2)
	sv.Set("/node2", 8)
	sv.Set("/node3", 1)
	for key, ele := range sv.Entries() {
		nsv.Set(key, ele)
	}
	assert.Equal(t, sv, nsv)
}

func TestStateVectorEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	sv.Set("/one", 1)
	sv.Set("/two", 2)
	comp := sv.ToComponent()
	nsv, _ := svs.ParseStateVector(comp)
	assert.Equal(t, uint(1), nsv.Get("/one"))
	assert.Equal(t, uint(2), nsv.Get("/two"))
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorDecodeStatic(t *testing.T) {
	comp, _ := enc.ComponentFromBytes([]byte{201, 24, 202, 10, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 202, 10, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2})
	sv, _ := svs.ParseStateVector(*comp)
	assert.Equal(t, uint(1), sv.Get("/one"))
	assert.Equal(t, uint(2), sv.Get("/two"))
	assert.Equal(t, int(2), sv.Len())
}
