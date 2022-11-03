package bitset

import (
	"testing"

	bitset "github.com/justincpresley/ndn-sync/util/bitset"
	assert "github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	b := bitset.New(10)
	b.Set(4, true)
	assert.True(t, b.Test(4))
	assert.False(t, b.Test(1))
	b.Set(1, true)
	b.Set(4, false)
	assert.True(t, b.Test(1))
	assert.False(t, b.Test(4))
	b.Clear()
	assert.False(t, b.Test(1))
}

func TestEncode(t *testing.T) {
	b := bitset.New(100)
	b.Set(4, true)
	b.Set(1, true)
	b.Set(10, true)
	b.Set(17, true)
	b.Set(76, true)
	buf, err := b.Bytes()
	assert.Nil(t, err)
	assert.Equal(t, []byte{4, 18, 4, 2, 0, 0, 0, 0, 0, 0, 16, 0, 0, 0}, buf)
}

func TestDecode(t *testing.T) {
	b, err := bitset.Parse([]byte{4, 18, 4, 2, 0, 0, 0, 0, 0, 0, 16, 0, 0, 0})
	assert.Nil(t, err)
	assert.True(t, b.Test(4))
	assert.True(t, b.Test(17))
	assert.True(t, b.Test(76))
	assert.False(t, b.Test(0))
	assert.False(t, b.Test(11))
	assert.False(t, b.Test(99))
}
