package svs

import (
	"time"
)

type Constants struct {
	IntervalRandomness         float32 // percentage variance 0.00<=x<=1.00
	BriefIntervalRandomness    float32 // percentage variance 0.00<=x<=1.00
	Interval                   time.Duration
	BriefInterval              time.Duration
	DataInterestLifeTime       time.Duration
	DataInterestRetries        uint // 0 = no retry
	DataPacketFreshness        time.Duration
	SyncInterestLifeTime       time.Duration
	MaxConcurrentDataInterests int32 // 0 = inf
	InitialFetchQueueLength    uint  // only helps to mitigate allocation resizing
	InitialMissingChannelSize  uint  // only helps to mitigate allocation resizing
}

func GetDefaultConstants() *Constants {
	return &Constants{
		IntervalRandomness:         0.10,
		BriefIntervalRandomness:    0.50,
		Interval:                   30000 * time.Millisecond,
		BriefInterval:              200 * time.Millisecond,
		DataInterestLifeTime:       2000 * time.Millisecond,
		DataInterestRetries:        2,
		DataPacketFreshness:        5000 * time.Millisecond,
		SyncInterestLifeTime:       1000 * time.Millisecond,
		MaxConcurrentDataInterests: 10,
		InitialFetchQueueLength:    50,
		InitialMissingChannelSize:  5,
	}
}
