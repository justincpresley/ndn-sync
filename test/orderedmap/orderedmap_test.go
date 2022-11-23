/*
 This module is a modified version of the original work.
 Original work can be found at:
          github.com/elliotchance/orderedmap

 The license provided (copy_of_license.md under /internal/orderedmap)
 covers this file. In addition, the changes or modifications done
 are described in changes.md under /internal/orderedmap.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.
*/

package orderedmap_test

import (
	"testing"

	om "github.com/justincpresley/ndn-sync/util/orderedmap"
	assert "github.com/stretchr/testify/assert"
)

func TestBasicFeatures(t *testing.T) {
	n := 100
	m := om.New[int, int]()

	for i := 0; i < n; i++ {
		assert.Equal(t, i, m.Len())
		present := m.Set(i, 2*i, true)
		assert.Equal(t, i+1, m.Len())
		assert.False(t, present)
	}

	for i := 0; i < n; i++ {
		value, present := m.Get(i)
		assert.Equal(t, 2*i, value)
		assert.True(t, present)
	}

	for i := 0; i < n; i++ {
		e := m.GetElement(i)
		assert.NotNil(t, e)
		assert.Equal(t, 2*i, e.Value)
	}
}

func TestUpdatingDoesntChangePairsOrder(t *testing.T) {
	m := om.New[string, any]()
	m.Set("foo", "bar", true)
	m.Set("wk", 28, true)
	m.Set("po", 100, true)
	m.Set("bar", "baz", true)

	present := m.Set("po", 102, true)
	assert.True(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "po", "bar"},
		[]any{"bar", 28, 102, "baz"})
}

func TestDeletingAndReinsertingChangesPairsOrder(t *testing.T) {
	m := om.New[string, any]()
	m.Set("foo", "bar", true)
	m.Set("wk", 28, true)
	m.Set("po", 100, true)
	m.Set("bar", "baz", true)

	// delete a element
	present := m.Remove("po")
	assert.True(t, present)

	// re-insert the same element
	present = m.Set("po", 100, true)
	assert.False(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "bar", "po"},
		[]any{"bar", 28, "baz", 100})
}

func TestEmptyMapOperations(t *testing.T) {
	m := om.New[string, any]()

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
	m := om.New[string, sampleStruct]()
	m.Set("foo", sampleStruct{"foo!"}, true)
	m.Set("bar", sampleStruct{"bar!"}, true)

	value, present := m.Get("foo")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "foo!", value.value)
	}

	present = m.Set("bar", sampleStruct{"baz!"}, true)
	assert.True(t, present)

	value, present = m.Get("bar")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "baz!", value.value)
	}
}

func assertOrderedPairsEqual[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	assertOrderedPairsEqualFromBack(t, m, expectedKeys, expectedValues)
	assertOrderedPairsEqualFromFront(t, m, expectedKeys, expectedValues)
}

func assertOrderedPairsEqualFromBack[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Key)
			assert.Equal(t, expectedValues[i], e.Value)
			i--
		}
	}
}

func assertOrderedPairsEqualFromFront[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for e := m.Back(); e != nil; e = e.Prev() {
			assert.Equal(t, expectedKeys[i], e.Key)
			assert.Equal(t, expectedValues[i], e.Value)
			i--
		}
	}
}
