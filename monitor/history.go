package monitor

import (
	"math"
	"sort"
	"sync"
	"time"
)

// Result stores the information about a single ping, in particular
// the round-trip time or whether the packet was lost.
type Result struct {
	RTT  time.Duration
	Lost bool
}

// History represents the ping history for a single node/device.
type History struct {
	capacity int
	results  []Result
	sync.RWMutex
}

// NewHistory creates a new History object with a specific capacity
func NewHistory(capacity int) History {
	return History{
		capacity: capacity,
		results:  make([]Result, 0, capacity+1),
	}
}

// AddResult saves a ping result into the internal history.
func (h *History) AddResult(rtt time.Duration, err error) {
	h.Lock()

	h.results = append(h.results, Result{RTT: rtt, Lost: err != nil})
	if len(h.results) > h.capacity {
		h.results = h.results[1:]
	}

	h.Unlock()
}

func (h *History) clear() {
}

// ComputeAndClear aggregates the result history into a single data point and clears the result set.
func (h *History) ComputeAndClear() *Metrics {
	h.Lock()
	result := h.compute()
	h.clear()
	h.Unlock()
	return result
}

// Compute aggregates the result history into a single data point.
func (h *History) Compute() *Metrics {
	h.RLock()
	defer h.RUnlock()
	return h.compute()
}

func (h *History) compute() *Metrics {
	numFailure := 0
	numTotal := len(h.results)
	µsPerMs := 1.0 / float64(time.Millisecond)

	if numTotal == 0 {
		return nil
	}

	data := make([]float64, 0, numTotal)
	var best, worst, mean, stddev, total, sumSquares float64
	var extremeFound bool

	for i := 0; i < numTotal; i++ {
		curr := &h.results[i]
		if curr.Lost {
			numFailure++
		} else {
			rtt := float64(curr.RTT) * µsPerMs
			data = append(data, rtt)

			if !extremeFound || rtt < best {
				best = rtt
			}
			if !extremeFound || rtt > worst {
				worst = rtt
			}

			extremeFound = true
			total += rtt
		}
	}

	size := float64(numTotal - numFailure)
	mean = total / size
	for _, rtt := range data {
		sumSquares += math.Pow(rtt-mean, 2)
	}
	stddev = math.Sqrt(sumSquares / size)

	median := math.NaN()
	if l := len(data); l > 0 {
		sort.Float64Slice(data).Sort()
		if l%2 == 0 {
			median = (data[l/2-1] + data[l/2]) / 2
		} else {
			median = data[l/2]
		}
	}

	return &Metrics{
		Results:     h.results,
		PacketsSent: numTotal,
		PacketsLost: numFailure,
		Best:        float32(best),
		Worst:       float32(worst),
		Median:      float32(median),
		Mean:        float32(mean),
		StdDev:      float32(stddev),
	}
}
