package svs

import (
	"time"

	enc "github.com/zjkmxy/go-ndn/pkg/encoding"
)

type Constants struct {
	SyncInterval                   time.Duration
	SuppressionInterval            time.Duration
	SyncIntervalJitter             float32 // percentage variance 0.00<=x<=1.00
	SuppressionIntervalJitter      float32 // percentage variance 0.00<=x<=1.00
	DataInterestLifeTime           time.Duration
	DataInterestRetries            uint // 0 = no retry
	DataPacketFreshness            time.Duration
	SyncInterestLifeTime           time.Duration
	DataComponent                  enc.Component
	SyncComponent                  enc.Component
	MaxConcurrentDataInterests     int32 // 0 = inf
	InitialFetchQueueSize          uint  // only helps to mitigate allocation resizing
	InitialMissingChannelSize      uint  // only helps to mitigate allocation resizing
	InitialStatusChangeChannelSize uint  // only helps to mitigate allocation resizing
	HeartbeatsToRenew              uint
	HeartbeatsToExpire             uint
	TrackRate                      time.Duration
	HeartbeatRate                  time.Duration
	MonitorInterval                time.Duration
}

func GetDefaultConstants() *Constants {
	return &Constants{
		SyncInterval:              30000 * time.Millisecond,
		SuppressionInterval:       200 * time.Millisecond,
		SyncIntervalJitter:        0.10,
		SuppressionIntervalJitter: 0.50,
		DataInterestLifeTime:      2000 * time.Millisecond,
		DataInterestRetries:       2,
		DataPacketFreshness:       5000 * time.Millisecond,
		SyncInterestLifeTime:      1000 * time.Millisecond,
		DataComponent: enc.Component{
			Typ: enc.TypeGenericNameComponent,
			Val: []byte{100, 97, 116, 97},
		},
		SyncComponent: enc.Component{
			Typ: enc.TypeGenericNameComponent,
			Val: []byte{115, 121, 110, 99},
		},
		MaxConcurrentDataInterests:     10,
		InitialFetchQueueSize:          50,
		InitialMissingChannelSize:      5,
		InitialStatusChangeChannelSize: 5,
		HeartbeatsToRenew:              3,
		HeartbeatsToExpire:             3,
		TrackRate:                      50000 * time.Millisecond,
		HeartbeatRate:                  45000 * time.Millisecond,
		MonitorInterval:                10 * time.Millisecond,
	}
}
