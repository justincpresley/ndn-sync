package orderedmap_test

import (
	"strconv"
	"testing"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestBasicFeatures(t *testing.T) {
	n := 100
	m := om.New[int](om.LatestEntriesFirst)

	for i := 0; i < n; i++ {
		assert.Equal(t, i, m.Len())
		n, _ := enc.NameFromStr(strconv.Itoa(i))
		present := m.Set(strconv.Itoa(i), n, 2*i, om.MetaV{Old: true})
		assert.Equal(t, i+1, m.Len())
		assert.False(t, present)
	}

	for i := 0; i < n; i++ {
		value, present := m.Get(strconv.Itoa(i))
		assert.Equal(t, 2*i, value)
		assert.True(t, present)
	}

	for i := 0; i < n; i++ {
		e := m.GetElement(strconv.Itoa(i))
		assert.NotNil(t, e)
		assert.Equal(t, 2*i, e.Val)
	}
}

func TestUpdatingDoesntChangePairsOrder(t *testing.T) {
	m := om.New[any](om.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, "bar", om.MetaV{Old: true})
	n, _ = enc.NameFromStr("wk")
	m.Set("wk", n, 28, om.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	m.Set("po", n, 100, om.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, "baz", om.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	present := m.Set("po", n, 102, om.MetaV{Old: true})
	assert.True(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "po", "bar"},
		[]any{"bar", 28, 102, "baz"})
}

func TestDifferentOrderings(t *testing.T) {
	m := om.New[any](om.Canonical)
	s := "oooo"
	n, _ := enc.NameFromStr(s)
	m.Set(s, n, "bar", om.MetaV{})
	s = "ooooo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, 28, om.MetaV{})
	s = "ooo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, 100, om.MetaV{})
	s = "oo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, "baz", om.MetaV{})
	s = "ooooo"
	n, _ = enc.NameFromStr(s)
	present := m.Set(s, n, 102, om.MetaV{})
	assert.True(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"oo", "ooo", "oooo", "ooooo"},
		[]any{"baz", 100, "bar", 102})
}

func TestDeletingAndReinsertingChangesPairsOrder(t *testing.T) {
	m := om.New[any](om.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, "bar", om.MetaV{Old: true})
	n, _ = enc.NameFromStr("wk")
	m.Set("wk", n, 28, om.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	m.Set("po", n, 100, om.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, "baz", om.MetaV{Old: true})

	present := m.Remove("po")
	assert.True(t, present)

	n, _ = enc.NameFromStr("po")
	present = m.Set("po", n, 100, om.MetaV{Old: true})
	assert.False(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "bar", "po"},
		[]any{"bar", 28, "baz", 100})
}

func TestEmptyMapOperations(t *testing.T) {
	m := om.New[any](om.LatestEntriesFirst)

	val, present := m.Get("foo")
	assert.Nil(t, val)
	assert.False(t, present)

	present = m.Remove("bar")
	assert.False(t, present)

	assert.Equal(t, 0, m.Len())
	assert.Nil(t, m.Front())
	assert.Nil(t, m.Back())
}

type sampleStruct struct {
	value string
}

func TestPackUnpackStructs(t *testing.T) {
	m := om.New[sampleStruct](om.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, sampleStruct{"foo!"}, om.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, sampleStruct{"bar!"}, om.MetaV{Old: true})

	value, present := m.Get("foo")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "foo!", value.value)
	}

	n, _ = enc.NameFromStr("bar")
	present = m.Set("bar", n, sampleStruct{"baz!"}, om.MetaV{Old: true})
	assert.True(t, present)

	value, present = m.Get("bar")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "baz!", value.value)
	}
}

func assertOrderedPairsEqual[V any](t *testing.T, m *om.OrderedMap[V], expectedKeys []string, expectedValues []V) {
	assertOrderedPairsEqualFromBack(t, m, expectedKeys, expectedValues)
	assertOrderedPairsEqualFromFront(t, m, expectedKeys, expectedValues)
}

func assertOrderedPairsEqualFromBack[V any](t *testing.T, m *om.OrderedMap[V], expectedKeys []string, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Kstr)
			assert.Equal(t, expectedValues[i], e.Val)
			i--
		}
	}
}

func assertOrderedPairsEqualFromFront[V any](t *testing.T, m *om.OrderedMap[V], expectedKeys []string, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Kstr)
			assert.Equal(t, expectedValues[i], e.Val)
			i--
		}
	}
}
