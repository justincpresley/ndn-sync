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
	sv.Set("/node1", 60, false)
	sv.Set("/node2", 9, false)
	sv.Set("/node1", 62, true)
	sv.Set("/node3", 1, false)
	assert.Equal(t, uint(62), sv.Get("/node1"))
	assert.Equal(t, uint(9), sv.Get("/node2"))
	assert.Equal(t, uint(1), sv.Get("/node3"))
	assert.Equal(t, uint(72), sv.Total())
	assert.Equal(t, int(3), sv.Len())
}

func TestStateVectorLoop(t *testing.T) {
	sv := svs.NewStateVector()
	nsv := svs.NewStateVector()
	sv.Set("/node1", 2, false)
	sv.Set("/node2", 8, false)
	sv.Set("/node3", 1, false)
	for pair := sv.Entries().First(); pair != nil; pair = pair.Next() {
		nsv.Set(pair.Key, pair.Value, true)
	}
	assert.Equal(t, sv, nsv)
}

func TestStateVectorEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	sv.Set("one", 1, true)
	sv.Set("two", 2, true)
	comp := sv.ToComponent()
	nsv, _ := svs.ParseStateVector(comp)
	assert.Equal(t, uint(1), nsv.Get("one"))
	assert.Equal(t, uint(2), nsv.Get("two"))
	assert.Equal(t, uint(3), nsv.Total())
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorDecodeStatic(t *testing.T) {
	comp, _ := enc.ComponentFromBytes([]byte{201, 16, 202, 3, 111, 110, 101, 204, 1, 1, 202, 3, 116, 119, 111, 204, 1, 2})
	sv, _ := svs.ParseStateVector(*comp)
	assert.Equal(t, uint(1), sv.Get("one"))
	assert.Equal(t, uint(2), sv.Get("two"))
	assert.Equal(t, uint(3), sv.Total())
	assert.Equal(t, int(2), sv.Len())
}

func TestStateVectorOrdering(t *testing.T) {
	sv1 := svs.NewStateVector()
	sv1.Set("one", 1, false)
	sv1.Set("two", 2, false)
	sv2 := svs.NewStateVector()
	sv2.Set("two", 2, true)
	sv2.Set("one", 1, true)
	for p1, p2 := sv1.Entries().First(), sv2.Entries().First(); p1 != nil; p1, p2 = p1.Next(), p2.Next() {
		assert.Equal(t, p1.Key, p2.Key)
		assert.Equal(t, p1.Value, p2.Value)
	}
	assert.Equal(t, sv1, sv2)
}
