package svs_test

import (
	"testing"

	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestCoreInitialState(t *testing.T) {
	syncPrefix, _ := enc.NameFromStr("/svs")
	nid, _ := enc.NameFromStr("/nodename")
	config := &svs.TwoStateCoreConfig{
		Source:         nid,
		SyncPrefix:     syncPrefix,
		FormalEncoding: false,
	}
	core := svs.NewCore(nil, config, svs.GetDefaultConstants())
	assert.Equal(t, uint64(0), core.Seqno())
	assert.Equal(t, svs.NewStateVector(), core.StateVector())
}
