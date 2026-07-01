package economy

import (
	"testing"

	"github.com/codyconfer/goose/internal/content"
)

func TestDefaultSettingsPreserveBaseBalance(t *testing.T) {
	def := DefaultSettings()
	if got := def.LevelMult(); got != 1 {
		t.Errorf("default level mult=%v, want 1", got)
	}
	if got := def.EventMult(); got != 1 {
		t.Errorf("default event mult=%v, want 1", got)
	}
	if got := def.MarketMult(); got != 1 {
		t.Errorf("default market mult=%v, want 1", got)
	}
}

func TestSettingsWithersRoundTripIndexes(t *testing.T) {
	s := DefaultSettings().WithLevel(0).WithEvent(4).WithMarket(5)
	if s.LevelIdx() != 0 {
		t.Errorf("level idx=%d, want 0", s.LevelIdx())
	}
	if s.EventIdx() != 4 {
		t.Errorf("event idx=%d, want 4", s.EventIdx())
	}
	if s.MarketIdx() != 5 {
		t.Errorf("market idx=%d, want 5", s.MarketIdx())
	}
}

func TestLevelPaceStretchesThresholds(t *testing.T) {

	fast := NewState()
	fast.Settings = DefaultSettings().WithLevel(0)
	fast.Eggs = LevelThresholds[1]
	if fast.Level() < 2 {
		t.Fatalf("fast pace at base threshold should be >= level 2, got %d", fast.Level())
	}

	eternal := NewState()
	last := len(content.Settings.LevelPace.Options) - 1
	eternal.Settings = DefaultSettings().WithLevel(last)
	eternal.Eggs = LevelThresholds[1]
	if eternal.Level() != 1 {
		t.Fatalf("eternal pace at base threshold should still be level 1, got %d", eternal.Level())
	}
}

func TestMarketPaceScalesProduction(t *testing.T) {
	base := NewState()
	base.Owned["gpu"] = 1

	slow := base.clone()
	slow.Settings = DefaultSettings().WithMarket(0)
	fast := base.clone()
	fast.Settings = DefaultSettings().WithMarket(base.Settings.MarketIdx() + 3)

	if !(fast.TokensPerSecond() > slow.TokensPerSecond()) {
		t.Fatalf("faster market pace should out-produce slower: fast=%v slow=%v",
			fast.TokensPerSecond(), slow.TokensPerSecond())
	}
}
