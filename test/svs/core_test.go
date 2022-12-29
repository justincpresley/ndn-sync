package svs

import (
	"testing"

	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestCoreInitialState(t *testing.T) {
	syncPrefix, _ := enc.NameFromStr("/svs")
	nid, _ := enc.NameFromStr("/nodename")
	config := &svs.CoreConfig{
		Source:         nid,
		SyncPrefix:     syncPrefix,
	}
	core := svs.NewCore(nil, config, svs.GetDefaultConstants())
	assert.Equal(t, uint64(0), core.GetSeqno())
	assert.Equal(t, svs.NewStateVector(), core.GetStateVector())
}
