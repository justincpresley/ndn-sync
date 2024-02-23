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
	n, _ := enc.NameFromStr("/node1")
	sv.Set("/node1", n, 60, false)
	n, _ = enc.NameFromStr("/node2")
	sv.Set("/node2", n, 9, false)
	n, _ = enc.NameFromStr("/node1")
	sv.Set("/node1", n, 62, true)
	n, _ = enc.NameFromStr("/node3")
	sv.Set("/node3", n, 1, false)
	assert.Equal(t, uint64(62), sv.Get("/node1"))
	assert.Equal(t, uint64(9), sv.Get("/node2"))
	assert.Equal(t, uint64(1), sv.Get("/node3"))
	assert.Equal(t, uint64(72), sv.Sum())
	assert.Equal(t, int(3), sv.Len())
}

func TestStateVectorLoop(t *testing.T) {
	sv := svs.NewStateVector()
	nsv := svs.NewStateVector()
	n, _ := enc.NameFromStr("/node1")
	sv.Set("/node1", n, 2, false)
	n, _ = enc.NameFromStr("/node2")
	sv.Set("/node2", n, 8, false)
	n, _ = enc.NameFromStr("/node3")
	sv.Set("/node3", n, 1, false)
	for pair := sv.Entries().Front(); pair != nil; pair = pair.Next() {
		nsv.Set(pair.Kstr, pair.Kname, pair.Val, true)
	}
	assert.Equal(t, sv, nsv)
}

func TestStateVectorFormalEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	n, _ := enc.NameFromStr("/one")
	sv.Set("/one", n, 1, true)
	n, _ = enc.NameFromStr("/two")
	sv.Set("/two", n, 2, true)
	wire := sv.Encode(true)
	nsv, _ := svs.ParseStateVector(enc.NewWireReader(wire), true)
	assert.Equal(t, uint64(1), nsv.Get("/one"))
	assert.Equal(t, uint64(2), nsv.Get("/two"))
	assert.Equal(t, uint64(3), nsv.Sum())
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorInformalEncodeDecode(t *testing.T) {
	sv := svs.NewStateVector()
	n, _ := enc.NameFromStr("/one")
	sv.Set("/one", n, 1, true)
	n, _ = enc.NameFromStr("/two")
	sv.Set("/two", n, 2, true)
	wire := sv.Encode(false)
	nsv, _ := svs.ParseStateVector(enc.NewWireReader(wire), false)
	assert.Equal(t, uint64(1), nsv.Get("/one"))
	assert.Equal(t, uint64(2), nsv.Get("/two"))
	assert.Equal(t, uint64(3), nsv.Sum())
	assert.Equal(t, int(2), nsv.Len())
	assert.Equal(t, sv, nsv)
}

func TestStateVectorFormalDecodeStatic(t *testing.T) {
	wire := enc.Wire{[]byte{201, 24, 202, 10, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 202, 10, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2}}
	sv, _ := svs.ParseStateVector(enc.NewWireReader(wire), true)
	assert.Equal(t, uint64(1), sv.Get("/one"))
	assert.Equal(t, uint64(2), sv.Get("/two"))
	assert.Equal(t, uint64(3), sv.Sum())
	assert.Equal(t, int(2), sv.Len())
}

func TestStateVectorInformalDecodeStatic(t *testing.T) {
	wire := enc.Wire{[]byte{201, 20, 7, 5, 8, 3, 111, 110, 101, 204, 1, 1, 7, 5, 8, 3, 116, 119, 111, 204, 1, 2}}
	sv, _ := svs.ParseStateVector(enc.NewWireReader(wire), false)
	assert.Equal(t, uint64(1), sv.Get("/one"))
	assert.Equal(t, uint64(2), sv.Get("/two"))
	assert.Equal(t, uint64(3), sv.Sum())
	assert.Equal(t, int(2), sv.Len())
}

func TestStateVectorOrdering(t *testing.T) {
	sv1 := svs.NewStateVector()
	n, _ := enc.NameFromStr("/one")
	sv1.Set("/one", n, 1, false)
	n, _ = enc.NameFromStr("/two")
	sv1.Set("/two", n, 2, false)
	sv2 := svs.NewStateVector()
	n, _ = enc.NameFromStr("/two")
	sv2.Set("/two", n, 2, true)
	n, _ = enc.NameFromStr("/one")
	sv2.Set("/one", n, 1, true)
	for p1, p2 := sv1.Entries().Front(), sv2.Entries().Front(); p1 != nil; p1, p2 = p1.Next(), p2.Next() {
		assert.Equal(t, p1.Kstr, p2.Kstr)
		assert.Equal(t, p1.Kname, p2.Kname)
		assert.Equal(t, p1.Val, p2.Val)
	}
	assert.Equal(t, sv1, sv2)
}
