package utilprint

import (
	"time"

	"github.com/hako/durafmt"
)

func Since(d time.Time) *durafmt.Durafmt {
	diff := time.Since(d)
	duration := durafmt.Parse(diff).LimitFirstN(1)
	return duration
}
