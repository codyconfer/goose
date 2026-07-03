package game

import (
	"time"

	"github.com/codyconfer/goose/internal/content"
)

var (
	upBeatRate = time.Duration(content.Tuning.UpBeatRateMs) * time.Millisecond

	flashBeats            = content.Tuning.FlashBeats
	notifBeats            = content.Tuning.NotifBeats
	outcomeBeats          = content.Tuning.OutcomeBeats
	characterTimeoutBeats = content.Tuning.CharacterTimeoutBeats
	offlineBeats          = content.Tuning.OfflineBeats

	notifQueueCap = content.Tuning.NotifQueueCap

	pulseDecayRate   = content.Tuning.PulseDecayRate
	buyRateSmoothing = content.Tuning.BuyRateSmoothing
)
