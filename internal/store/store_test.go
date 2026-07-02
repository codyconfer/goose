package store

import (
	"path/filepath"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/worldgen"
)

func TestFileStoreRoundTrip(t *testing.T) {
	store := FileStore{Path: filepath.Join(t.TempDir(), "save.json")}
	if store.Exists() {
		t.Fatal("store reports existing before any write")
	}
	fresh := economy.NewMachine()
	if fs := fresh.Get(); fs.Owned == nil || fs.UpgradeLevels == nil || fs.PerClick != 1 {
		t.Fatal("fresh state not normalized")
	}

	state := fresh.Get()
	state.Tokens = 1234
	state.Owned["gpu"] = 3
	state.UpgradeLevels[economy.UpgradeClick] = 2
	state.Ledger = []economy.Transaction{{Kind: economy.TxSellEggs, Amount: 12, Tokens: 24}}
	state.PriceCandles = []economy.PriceCandle{{Open: 1, High: 2, Low: 0.5, Close: 1.5}}
	state.PriceCandleBeats = 2
	m := economy.FromState(state)
	evm := events.FromState(events.State{Fired: map[string]bool{"first-egg": true}})
	wrld := worldgen.Generate(42)
	info, err := store.Create("Flock 1", m, evm, wrld)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if info.ID <= 0 || info.Name != "Flock 1" {
		t.Fatalf("bad save info: %+v", info)
	}
	if !store.Exists() {
		t.Fatal("store reports missing after write")
	}

	got, gotEv, gotWorld, err := store.Read(info.ID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if gs := got.Get(); gs.Tokens != 1234 || gs.Owned["gpu"] != 3 || gs.UpgradeLevel(economy.UpgradeClick) != 2 {
		t.Fatalf("round-trip mismatch: %+v", got.Get())
	}
	gs := got.Get()
	if len(gs.Ledger) != 1 || gs.Ledger[0].Tokens != 24 {
		t.Fatalf("ledger not preserved: %+v", gs.Ledger)
	}
	if len(gs.PriceCandles) != 1 || gs.PriceCandles[0].Close != 1.5 || gs.PriceCandleBeats != 2 {
		t.Fatalf("price chart not preserved: candles=%+v beats=%d", gs.PriceCandles, gs.PriceCandleBeats)
	}
	if !gotEv.Get().HasFired("first-egg") {
		t.Fatal("events state not preserved across the file round-trip")
	}
	if gotWorld.Seed != 42 {
		t.Fatalf("world seed=%d, want 42", gotWorld.Seed)
	}

	updated := economy.FromState(func() economy.State { s := got.Get(); s.Tokens = 5678; return s }())
	if err := store.Write(info.ID, updated, gotEv, gotWorld); err != nil {
		t.Fatalf("write update: %v", err)
	}
	again, _, againWorld, err := store.Read(info.ID)
	if err != nil {
		t.Fatalf("read update: %v", err)
	}
	if again.Get().Tokens != 5678 {
		t.Fatalf("update not persisted: got %v", again.Get().Tokens)
	}
	if againWorld.Seed != 42 {
		t.Fatalf("updated world seed=%d, want 42", againWorld.Seed)
	}
}
