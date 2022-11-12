/*
 This module is a modified version of the original work.
 Original work can be found at:
          github.com/reddragon/bloomfilter.go

 No license was provided. However, the changes or modifications
 done are described in changes.md under /internal/bloomfilter.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.
*/

package bloomfilter

import (
	"testing"

	bf "github.com/justincpresley/ndn-sync/util/bloomfilter"
	assert "github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	b := bf.NewFilter(3, 100)
	d1, d2 := []byte("Hello"), []byte("Jello")
	b.Add(d1)

	assert.True(t, b.Check(d1))
	assert.False(t, b.Check(d2))
}

func TestEncode(t *testing.T) {
	b := bf.NewFilter(3, 100)
	b.Add([]byte("TEST_ENTRY1"))
	b.Add([]byte("ANOTHER_ENTRY2"))

	buf, err := b.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte{3, 4, 0, 0, 0, 0, 0, 64, 0, 0, 2, 0, 65, 68, 0}, buf)
}

func TestDecode(t *testing.T) {
	b, err := bf.ParseFilter([]byte{3, 4, 0, 0, 0, 0, 0, 64, 0, 0, 2, 0, 65, 68, 0})
	assert.Nil(t, err)
	assert.True(t, b.Check([]byte("TEST_ENTRY1")))
	assert.True(t, b.Check([]byte("ANOTHER_ENTRY2")))
}
