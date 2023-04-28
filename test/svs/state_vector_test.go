package svs_test

import (
	"testing"

	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestStateVectorBasics(t *testing.T) {
	sv := svs.NewStateVector()
	assert.Equal(t, uint64(0), sv.Get("/node1"))
	sv.Set("/node1", 60)
	sv.Set("/node2", 9)
	sv.Set("/node1", 62)
	sv.Set("/node3", 1)
	assert.Equal(t, uint64(62), sv.Get("/node1"))
	assert.Equal(t, uint64(9), sv.Get("/node2"))
	assert.Equal(t, uint64(1), sv.Get("/node3"))
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
	assert.Equal(t, uint64(1), nsv.Get("/one"))
	assert.Equal(t, uint64(2), nsv.Get("/two"))
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorDecodeStatic(t *testing.T) {
	comp, _ := enc.ComponentFromBytes([]byte{201, 24, 202, 10, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 202, 10, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2})
	sv, _ := svs.ParseStateVector(comp)
	assert.Equal(t, uint64(1), sv.Get("/one"))
	assert.Equal(t, uint64(2), sv.Get("/two"))
	assert.Equal(t, int(2), sv.Len())
}
