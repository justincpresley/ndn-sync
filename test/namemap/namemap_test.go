package orderedmap_test

import (
	"strconv"
	"testing"

	nm "github.com/justincpresley/ndn-sync/util/namemap"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestBasicFeatures(t *testing.T) {
	n := 100
	m := nm.New[int](nm.LatestEntriesFirst)

	for i := range n {
		assert.Equal(t, i, m.Len())
		n, _ := enc.NameFromStr(strconv.Itoa(i))
		present := m.Set(strconv.Itoa(i), n, 2*i, nm.MetaV{Old: true})
		assert.Equal(t, i+1, m.Len())
		assert.False(t, present)
	}

	for i := range n {
		value, present := m.Get(strconv.Itoa(i))
		assert.Equal(t, 2*i, value)
		assert.True(t, present)
	}

	for i := range n {
		e := m.GetElement(strconv.Itoa(i))
		assert.NotNil(t, e)
		assert.Equal(t, 2*i, e.Val)
	}
}

func TestCopy(t *testing.T) {
	n := 100
	m := nm.New[int](nm.LatestEntriesFirst)

	for i := range n {
		assert.Equal(t, i, m.Len())
		n, _ := enc.NameFromStr(strconv.Itoa(i))
		present := m.Set(strconv.Itoa(i), n, 2*i, nm.MetaV{Old: true})
		assert.Equal(t, i+1, m.Len())
		assert.False(t, present)
	}

	c := m.Copy()

	for i := range n {
		value, present := c.Get(strconv.Itoa(i))
		assert.Equal(t, 2*i, value)
		assert.True(t, present)
	}

	for i := range n {
		e := c.GetElement(strconv.Itoa(i))
		assert.NotNil(t, e)
		assert.Equal(t, 2*i, e.Val)
	}
}


func TestUpdatingDoesntChangePairsOrder(t *testing.T) {
	m := nm.New[any](nm.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, "bar", nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("wk")
	m.Set("wk", n, 28, nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	m.Set("po", n, 100, nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, "baz", nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	present := m.Set("po", n, 102, nm.MetaV{Old: true})
	assert.True(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "po", "bar"},
		[]any{"bar", 28, 102, "baz"})
}

func TestDifferentOrderings(t *testing.T) {
	m := nm.New[any](nm.Canonical)
	s := "oooo"
	n, _ := enc.NameFromStr(s)
	m.Set(s, n, "bar", nm.MetaV{})
	s = "ooooo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, 28, nm.MetaV{})
	s = "ooo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, 100, nm.MetaV{})
	s = "oo"
	n, _ = enc.NameFromStr(s)
	m.Set(s, n, "baz", nm.MetaV{})
	s = "ooooo"
	n, _ = enc.NameFromStr(s)
	present := m.Set(s, n, 102, nm.MetaV{})
	assert.True(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"oo", "ooo", "oooo", "ooooo"},
		[]any{"baz", 100, "bar", 102})
}

func TestDeletingAndReinsertingChangesPairsOrder(t *testing.T) {
	m := nm.New[any](nm.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, "bar", nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("wk")
	m.Set("wk", n, 28, nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("po")
	m.Set("po", n, 100, nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, "baz", nm.MetaV{Old: true})

	present := m.Remove("po")
	assert.True(t, present)

	n, _ = enc.NameFromStr("po")
	present = m.Set("po", n, 100, nm.MetaV{Old: true})
	assert.False(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "bar", "po"},
		[]any{"bar", 28, "baz", 100})
}

func TestEmptyMapOperations(t *testing.T) {
	m := nm.New[any](nm.LatestEntriesFirst)

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
	m := nm.New[sampleStruct](nm.LatestEntriesFirst)
	n, _ := enc.NameFromStr("foo")
	m.Set("foo", n, sampleStruct{"foo!"}, nm.MetaV{Old: true})
	n, _ = enc.NameFromStr("bar")
	m.Set("bar", n, sampleStruct{"bar!"}, nm.MetaV{Old: true})

	value, present := m.Get("foo")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "foo!", value.value)
	}

	n, _ = enc.NameFromStr("bar")
	present = m.Set("bar", n, sampleStruct{"baz!"}, nm.MetaV{Old: true})
	assert.True(t, present)

	value, present = m.Get("bar")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "baz!", value.value)
	}
}

func assertOrderedPairsEqual[V any](t *testing.T, m *nm.NameMap[V], expectedKeys []string, expectedValues []V) {
	assertOrderedPairsEqualFromBack(t, m, expectedKeys, expectedValues)
	assertOrderedPairsEqualFromFront(t, m, expectedKeys, expectedValues)
}

func assertOrderedPairsEqualFromBack[V any](t *testing.T, m *nm.NameMap[V], expectedKeys []string, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Kstr)
			assert.Equal(t, expectedValues[i], e.Val)
			i--
		}
	}
}

func assertOrderedPairsEqualFromFront[V any](t *testing.T, m *nm.NameMap[V], expectedKeys []string, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Kstr)
			assert.Equal(t, expectedValues[i], e.Val)
			i--
		}
	}
}
