package domain

import (
	"fmt"
	"time"
)

func (w Warning) LevelString() string {
	return fmt.Sprintf("%s rainfall=%dmm", w.Level, w.RainfallMM)
}

func (w Warning) Covers(at time.Time) bool {
	return !at.Before(w.EffectiveFrom) && at.Before(w.EffectiveUntil)
}
