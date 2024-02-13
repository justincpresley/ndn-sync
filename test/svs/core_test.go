package svs_test

import (
	"testing"

	svs "github.com/justincpresley/ndn-sync/pkg/svs"
	assert "github.com/stretchr/testify/assert"
	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

func TestCoreInitialState(t *testing.T) {
	syncPrefix, _ := enc.NameFromStr("/svs")
	config := &svs.TwoStateCoreConfig{
		SyncPrefix: syncPrefix,
	}
	core := svs.NewCore(nil, config, svs.GetDefaultConstants())
	assert.Equal(t, svs.NewStateVector(), core.StateVector())
}
