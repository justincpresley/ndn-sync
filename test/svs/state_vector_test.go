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
	sv.Set("/node1", 60, false)
	sv.Set("/node2", 9, false)
	sv.Set("/node1", 62, true)
	sv.Set("/node3", 1, false)
	assert.Equal(t, uint64(62), sv.Get("/node1"))
	assert.Equal(t, uint64(9), sv.Get("/node2"))
	assert.Equal(t, uint64(1), sv.Get("/node3"))
	assert.Equal(t, uint64(72), sv.Total())
	assert.Equal(t, int(3), sv.Len())
}

func TestStateVectorLoop(t *testing.T) {
	sv := svs.NewStateVector()
	nsv := svs.NewStateVector()
	sv.Set("/node1", 2, false)
	sv.Set("/node2", 8, false)
	sv.Set("/node3", 1, false)
	for pair := sv.Entries().Front(); pair != nil; pair = pair.Next() {
		nsv.Set(pair.Key, pair.Value, true)
	}
	assert.Equal(t, sv, nsv)
}

func TestStateVectorFormalEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	sv.Set("/one", 1, true)
	sv.Set("/two", 2, true)
	comp := sv.ToComponent(true)
	nsv, _ := svs.ParseStateVector(comp, true)
	assert.Equal(t, uint64(1), nsv.Get("/one"))
	assert.Equal(t, uint64(2), nsv.Get("/two"))
	assert.Equal(t, uint64(3), nsv.Total())
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorInformalEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	sv.Set("/one", 1, true)
	sv.Set("/two", 2, true)
	comp := sv.ToComponent(false)
	nsv, _ := svs.ParseStateVector(comp, false)
	assert.Equal(t, uint64(1), nsv.Get("/one"))
	assert.Equal(t, uint64(2), nsv.Get("/two"))
	assert.Equal(t, uint64(3), nsv.Total())
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorFormalDecodeStatic(t *testing.T) {
	comp, _ := enc.ComponentFromBytes([]byte{201, 24, 202, 10, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 202, 10, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2})
	sv, _ := svs.ParseStateVector(comp, true)
	assert.Equal(t, uint64(1), sv.Get("/one"))
	assert.Equal(t, uint64(2), sv.Get("/two"))
	assert.Equal(t, uint64(3), sv.Total())
	assert.Equal(t, int(2), sv.Len())
}

func TestStateVectorInformalDecodeStatic(t *testing.T) {
	comp, _ := enc.ComponentFromBytes([]byte{201, 20, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2})
	sv, _ := svs.ParseStateVector(comp, false)
	assert.Equal(t, uint64(1), sv.Get("/one"))
	assert.Equal(t, uint64(2), sv.Get("/two"))
	assert.Equal(t, uint64(3), sv.Total())
	assert.Equal(t, int(2), sv.Len())
}

func TestStateVectorOrdering(t *testing.T) {
	sv1 := svs.NewStateVector()
	sv1.Set("/one", 1, false)
	sv1.Set("/two", 2, false)
	sv2 := svs.NewStateVector()
	sv2.Set("/two", 2, true)
	sv2.Set("/one", 1, true)
	for p1, p2 := sv1.Entries().Front(), sv2.Entries().Front(); p1 != nil; p1, p2 = p1.Next(), p2.Next() {
		assert.Equal(t, p1.Key, p2.Key)
		assert.Equal(t, p1.Value, p2.Value)
	}
	assert.Equal(t, sv1, sv2)
}
