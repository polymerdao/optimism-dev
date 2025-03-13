package event

import (
	"os"
	"strconv"

	"golang.org/x/time/rate"
)

type ExecutorOpts struct {
	Capacity int // If there is a local buffer capacity
}

type EmitterOpts struct {
	Limiting  bool
	Rate      rate.Limit
	Burst     int
	OnLimited func()
}

// RegisterOpts represents the set of parameters to configure a
// new deriver/emitter with that is registered with an event System.
// These options may be reused for multiple registrations.
type RegisterOpts struct {
	Executor ExecutorOpts
	Emitter  EmitterOpts
}

// 200 events may be buffered per deriver before back-pressure has to kick in
const eventsBuffer = 200

// 10,000 events per second is plenty.
// If we are going through more events, the driver needs to breathe, and warn the user of a potential issue.
const eventsLimit = rate.Limit(10_000)

// 500 events of burst: the maximum amount of events to eat up
// past the rate limit before the rate limit becomes applicable.
const eventsBurst = 500

func DefaultRegisterOpts() *RegisterOpts {
	emitRate := eventsLimit
	if str, ok := os.LookupEnv("OP_NODE_EVENT_RATE_LIMIT"); ok {
		if e, err := strconv.ParseUint(str, 10, 64); err == nil {
			emitRate = rate.Limit(e)
		}
	}

	burst := eventsBurst
	if str, ok := os.LookupEnv("OP_NODE_EVENT_BURST_LIMIT"); ok {
		if e, err := strconv.ParseInt(str, 10, 64); err == nil {
			burst = int(e)
		}
	}

	return &RegisterOpts{
		Executor: ExecutorOpts{
			Capacity: eventsBuffer,
		},
		Emitter: EmitterOpts{
			Limiting:  true,
			Rate:      emitRate,
			Burst:     burst,
			OnLimited: nil,
		},
	}
}
