package svs

type Constants struct {
	Interval                   uint    // (ms) milliseconds
	IntervalRandomness         float32 // percentage variance 0.00<=x<=1.00
	BriefInterval              uint    // (ms) milliseconds
	BriefIntervalRandomness    float32 // percentage variance 0.00<=x<=1.00
	DataInterestLifeTime       uint    // (ms) milliseconds
	DataInterestRetries        uint    // 0 = no retry
	DataPacketFressness        uint    // (ms) milliseconds
	SyncInterestLifeTime       uint    // (ms) milliseconds
	MaxConcurrentDataInterests uint    // 0 = inf
	InitialFetchQueueLength    uint    // only helps to mitigate allocation resizing
}

func GetDefaultConstants() *Constants {
	return &Constants{
		Interval:                   30000,
		IntervalRandomness:         0.10,
		BriefInterval:              200,
		BriefIntervalRandomness:    0.50,
		DataInterestLifeTime:       2000,
		DataInterestRetries:        2,
		DataPacketFressness:        5000,
		SyncInterestLifeTime:       1000,
		MaxConcurrentDataInterests: 10,
		InitialFetchQueueLength:    50,
	}
}
