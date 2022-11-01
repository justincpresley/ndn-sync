/*
 This module is a modified version of the original work.
 Original work can be found at:
          https://github.com/reddragon/bloomfilter.go

 No license was provided. However, the changes or modifications
 done are described in changes.md under /internal/bloomfilter.
 I do not claim ownership or creation of this module. All
 credit should be given to the original author.
*/

package bloomfilter

import (
	"testing"

	bf "github.com/justincpresley/ndn-sync/internal/bloomfilter"
	assert "github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	b := bf.NewFilter(3, 100)
	d1, d2 := []byte("Hello"), []byte("Jello")
	b.Add(d1)

	assert.True(t, b.Check(d1))
	assert.False(t, b.Check(d2))
}
