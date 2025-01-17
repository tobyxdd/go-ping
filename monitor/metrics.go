package monitor

// Metrics is a dumb data point computed from a history of Results.
type Metrics struct {
	Results     []Result
	PacketsSent int     // number of packets sent
	PacketsLost int     // number of packets lost
	Best        float32 // best rtt in ms
	Worst       float32 // worst rtt in ms
	Median      float32 // median rtt in ms
	Mean        float32 // mean rtt in ms
	StdDev      float32 // std deviation in ms
}
