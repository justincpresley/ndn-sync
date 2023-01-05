package svs

type Constants struct {
	IntervalRandomness             float32 // percentage variance 0.00<=x<=1.00
	BriefIntervalRandomness        float32 // percentage variance 0.00<=x<=1.00
	Interval                       uint    // (ms) milliseconds
	BriefInterval                  uint    // (ms) milliseconds
	DataInterestLifeTime           uint    // (ms) milliseconds
	DataInterestRetries            uint    // 0 = no retry
	DataPacketFressness            uint    // (ms) milliseconds
	SyncInterestLifeTime           uint    // (ms) milliseconds
	MaxConcurrentDataInterests     int32   // 0 = inf
	InitialFetchQueueLength        uint    // only helps to mitigate allocation resizing
	InitialMissingChannelSize      uint    // only helps to mitigate allocation resizing
	InitialStatusChangeChannelSize uint    // only helps to mitigate allocation resizing
	HeartbeatsToRenew              uint
	HeartbeatsToExpire             uint
	TrackRate                      int64
	HeartbeatRate                  uint
	MonitorInterval                uint
}

func GetDefaultConstants() *Constants {
	return &Constants{
		IntervalRandomness:             0.10,
		BriefIntervalRandomness:        0.50,
		Interval:                       30000,
		BriefInterval:                  200,
		DataInterestLifeTime:           2000,
		DataInterestRetries:            2,
		DataPacketFressness:            5000,
		SyncInterestLifeTime:           1000,
		MaxConcurrentDataInterests:     10,
		InitialFetchQueueLength:        50,
		InitialMissingChannelSize:      5,
		InitialStatusChangeChannelSize: 5,
		HeartbeatsToRenew:              3,
		HeartbeatsToExpire:             3,
		TrackRate:                      50000,
		HeartbeatRate:                  45000,
		MonitorInterval:                10,
	}
}
