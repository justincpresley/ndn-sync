/*
 This module is a modified version of the original work.
 Original work can be found at:
          https://github.com/wk8/go-ordered-map

 The license provided (copy_of_license.md under /internal/orderedmap)
 covers this file. In addition, the changes or modifications done
 are described in changes.md under /internal/orderedmap.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.
*/

package orderedmap

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"testing"

	om "github.com/justincpresley/ndn-sync/internal/orderedmap"
	assert "github.com/stretchr/testify/assert"
)

func TestBasicFeatures(t *testing.T) {
	n := 100
	m := om.New[int, int]()

	// set(i, 2 * i)
	for i := 0; i < n; i++ {
		assertLenEqual(t, m, i)
		oldValue, present := m.Set(i, 2*i, true)
		assertLenEqual(t, m, i+1)

		assert.Equal(t, 0, oldValue)
		assert.False(t, present)
	}

	// get what we just set
	for i := 0; i < n; i++ {
		value, present := m.Get(i)

		assert.Equal(t, 2*i, value)
		assert.True(t, present)
	}

	// get pairs of what we just set
	for i := 0; i < n; i++ {
		pair := m.GetPair(i)

		assert.NotNil(t, pair)
		assert.Equal(t, 2*i, pair.Value)
	}

	// forward iteration
	i := 0
	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 2*i, pair.Value)
		i++
	}
	// backward iteration
	i = n - 1
	for pair := m.Newest(); pair != nil; pair = pair.Prev() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 2*i, pair.Value)
		i--
	}

	// forward iteration starting from known key
	i = 42
	for pair := m.GetPair(i); pair != nil; pair = pair.Next() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 2*i, pair.Value)
		i++
	}

	// double values for pairs with even keys
	for j := 0; j < n/2; j++ {
		i = 2 * j
		oldValue, present := m.Set(i, 4*i, true)

		assert.Equal(t, 2*i, oldValue)
		assert.True(t, present)
	}
	// and delete pairs with odd keys
	for j := 0; j < n/2; j++ {
		i = 2*j + 1
		assertLenEqual(t, m, n-j)
		value, present := m.Delete(i)
		assertLenEqual(t, m, n-j-1)

		assert.Equal(t, 2*i, value)
		assert.True(t, present)

		// deleting again shouldn't change anything
		value, present = m.Delete(i)
		assertLenEqual(t, m, n-j-1)
		assert.Equal(t, 0, value)
		assert.False(t, present)
	}

	// get the whole range
	for j := 0; j < n/2; j++ {
		i = 2 * j
		value, present := m.Get(i)
		assert.Equal(t, 4*i, value)
		assert.True(t, present)

		i = 2*j + 1
		value, present = m.Get(i)
		assert.Equal(t, 0, value)
		assert.False(t, present)
	}

	// check iterations again
	i = 0
	for pair := m.Oldest(); pair != nil; pair = pair.Next() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i += 2
	}
	i = 2 * ((n - 1) / 2)
	for pair := m.Newest(); pair != nil; pair = pair.Prev() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i -= 2
	}

	// check iterations with aliases
	i = 0
	for pair := m.Front(); pair != nil; pair = pair.Next() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i += 2
	}
	i = 2 * ((n - 1) / 2)
	for pair := m.Back(); pair != nil; pair = pair.Prev() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i -= 2
	}
	i = 0
	for pair := m.First(); pair != nil; pair = pair.Next() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i += 2
	}
	i = 2 * ((n - 1) / 2)
	for pair := m.Last(); pair != nil; pair = pair.Prev() {
		assert.Equal(t, i, pair.Key)
		assert.Equal(t, 4*i, pair.Value)
		i -= 2
	}

	// check cloning
	cm := m.Clone()
	assert.NotSame(t, cm, m)
	for np, cp := m.Front(), cm.Front(); np != nil; np, cp = np.Next(), cp.Next() {
		assert.Equal(t, np.Key, cp.Key)
		assert.Equal(t, np.Value, cp.Value)
	}
	// check cloning alias
	cm = m.Copy()
	assert.NotSame(t, cm, m)
	for np, cp := m.Front(), cm.Front(); np != nil; np, cp = np.Next(), cp.Next() {
		assert.Equal(t, np.Key, cp.Key)
		assert.Equal(t, np.Value, cp.Value)
	}

	// check reversing
	rm := m.Clone()
	rm.Reverse()
	for np, rp := m.Front(), rm.Back(); np != nil; np, rp = np.Next(), rp.Prev() {
		assert.Equal(t, np.Key, rp.Key)
		assert.Equal(t, np.Value, rp.Value)
	}
	rm.Reverse()
	for np, rp := m.Front(), rm.Front(); np != nil; np, rp = np.Next(), rp.Next() {
		assert.Equal(t, np.Key, rp.Key)
		assert.Equal(t, np.Value, rp.Value)
	}

	// check sizing
	assert.Equal(t, m.Size(), m.Len())
}

func TestUpdatingDoesntChangePairsOrder(t *testing.T) {
	m := om.New[string, any]()
	m.Set("foo", "bar", true)
	m.Set("wk", 28, true)
	m.Set("po", 100, true)
	m.Set("bar", "baz", true)

	oldValue, present := m.Set("po", 102, true)
	assert.Equal(t, 100, oldValue)
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

	// delete a pair
	oldValue, present := m.Delete("po")
	assert.Equal(t, 100, oldValue)
	assert.True(t, present)

	// re-insert the same pair
	oldValue, present = m.Set("po", 100, true)
	assert.Nil(t, oldValue)
	assert.False(t, present)

	assertOrderedPairsEqual(t, m,
		[]string{"foo", "wk", "bar", "po"},
		[]any{"bar", 28, "baz", 100})
}

func TestEmptyMapOperations(t *testing.T) {
	m := om.New[string, any]()

	oldValue, present := m.Get("foo")
	assert.Nil(t, oldValue)
	assert.False(t, present)

	oldValue, present = m.Delete("bar")
	assert.Nil(t, oldValue)
	assert.False(t, present)

	assertLenEqual(t, m, 0)

	assert.Nil(t, m.Oldest())
	assert.Nil(t, m.Newest())
}

type dummyTestStruct struct {
	value string
}

func TestPackUnpackStructs(t *testing.T) {
	m := om.New[string, dummyTestStruct]()
	m.Set("foo", dummyTestStruct{"foo!"}, true)
	m.Set("bar", dummyTestStruct{"bar!"}, true)

	value, present := m.Get("foo")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "foo!", value.value)
	}

	value, present = m.Set("bar", dummyTestStruct{"baz!"}, true)
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "bar!", value.value)
	}

	value, present = m.Get("bar")
	assert.True(t, present)
	if assert.NotNil(t, value) {
		assert.Equal(t, "baz!", value.value)
	}
}

// shamelessly stolen from https://github.com/python/cpython/blob/e19a91e45fd54a56e39c2d12e6aaf4757030507f/Lib/test/test_ordered_dict.py#L55-L61
func TestShuffle(t *testing.T) {
	ranLen := 100

	for _, n := range []int{0, 10, 20, 100, 1000, 10000} {
		t.Run(fmt.Sprintf("shuffle test with %d items", n), func(t *testing.T) {
			m := om.New[string, string]()

			keys := make([]string, n)
			values := make([]string, n)

			for i := 0; i < n; i++ {
				// we prefix with the number to ensure that we don't get any duplicates
				keys[i] = fmt.Sprintf("%d_%s", i, randomHexString(t, ranLen))
				values[i] = randomHexString(t, ranLen)

				value, present := m.Set(keys[i], values[i], true)
				assert.Equal(t, "", value)
				assert.False(t, present)
			}

			assertOrderedPairsEqual(t, m, keys, values)
		})
	}
}

func TestMove(t *testing.T) {
	m := om.New[int, any]()
	m.Set(1, "bar", true)
	m.Set(3, 100, true)
	m.Set(2, 28, true)
	m.Set(4, "baz", true)
	m.Set(5, "28", true)
	m.Set(6, "100", true)
	m.Set(7, "baz", true)
	m.Set(8, "baz", true)

	err := m.MoveAfter(2, 3)
	assert.Nil(t, err)
	assertOrderedPairsEqual(t, m,
		[]int{1, 3, 2, 4, 5, 6, 7, 8},
		[]any{"bar", 100, 28, "baz", "28", "100", "baz", "baz"})

	err = m.MoveBefore(6, 4)
	assert.Nil(t, err)
	assertOrderedPairsEqual(t, m,
		[]int{1, 3, 2, 6, 4, 5, 7, 8},
		[]any{"bar", 100, 28, "100", "baz", "28", "baz", "baz"})

	err = m.MoveToBack(3)
	assert.Nil(t, err)
	assertOrderedPairsEqual(t, m,
		[]int{1, 2, 6, 4, 5, 7, 8, 3},
		[]any{"bar", 28, "100", "baz", "28", "baz", "baz", 100})

	err = m.MoveToFront(5)
	assert.Nil(t, err)
	assertOrderedPairsEqual(t, m,
		[]int{5, 1, 2, 6, 4, 7, 8, 3},
		[]any{"28", "bar", 28, "100", "baz", "baz", "baz", 100})

	err = m.MoveToFront(100)
	assert.NotEqual(t, err, nil)
}

/* Test helpers */

func assertOrderedPairsEqual[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	assertOrderedPairsEqualFromNewest(t, m, expectedKeys, expectedValues)
	assertOrderedPairsEqualFromOldest(t, m, expectedKeys, expectedValues)
}

func assertOrderedPairsEqualFromNewest[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for pair := m.Newest(); pair != nil; pair = pair.Prev() {
			assert.Equal(t, expectedKeys[i], pair.Key)
			assert.Equal(t, expectedValues[i], pair.Value)
			i--
		}
	}
}

func assertOrderedPairsEqualFromOldest[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedKeys []K, expectedValues []V) {
	if assert.Equal(t, len(expectedKeys), len(expectedValues)) && assert.Equal(t, len(expectedKeys), m.Len()) {
		i := m.Len() - 1
		for pair := m.Newest(); pair != nil; pair = pair.Prev() {
			assert.Equal(t, expectedKeys[i], pair.Key)
			assert.Equal(t, expectedValues[i], pair.Value)
			i--
		}
	}
}

func assertLenEqual[K comparable, V any](t *testing.T, m *om.OrderedMap[K, V], expectedLen int) {
	assert.Equal(t, expectedLen, m.Len())
}

func randomHexString(t *testing.T, length int) string {
	b := length / 2
	randBytes := make([]byte, b)

	if n, err := rand.Read(randBytes); err != nil || n != b {
		if err == nil {
			err = fmt.Errorf("only got %v random bytes, expected %v", n, b)
		}
		t.Fatal(err)
	}

	return hex.EncodeToString(randBytes)
}
