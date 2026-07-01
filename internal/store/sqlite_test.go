package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestSQLiteStoreRoundTrip(t *testing.T) {
	store := SQLiteStore{Path: filepath.Join(t.TempDir(), "save.db")}
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
	state.Transactions = []economy.Transaction{{Kind: economy.TxBuyEggs, Amount: 50}}
	state.Positions = []economy.Position{{Kind: economy.PosCall, Strike: 2.5, Premium: 200, Leverage: 5, Expiry: 30}}
	state.Ledger = []economy.Transaction{{Kind: economy.TxSellEggs, Amount: 12, Tokens: 24}}
	state.PriceCandles = []economy.PriceCandle{{Open: 1, High: 2, Low: 0.5, Close: 1.5}}
	state.PriceCandleBeats = 2
	m := economy.FromState(state)
	evm := events.FromState(events.State{Fired: map[string]bool{"first-egg": true}})
	info, err := store.Create("Flock 1", m, evm)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if info.ID <= 0 || info.Name != "Flock 1" {
		t.Fatalf("bad save info: %+v", info)
	}
	if !store.Exists() {
		t.Fatal("store reports missing after write")
	}

	got, gotEv, err := store.Read(info.ID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	gs := got.Get()
	if gs.Tokens != 1234 || gs.Owned["gpu"] != 3 || gs.UpgradeLevel(economy.UpgradeClick) != 2 {
		t.Fatalf("round-trip mismatch: %+v", gs)
	}
	if len(gs.Transactions) != 1 || gs.Transactions[0].Amount != 50 {
		t.Fatalf("transactions not preserved: %+v", gs.Transactions)
	}
	if len(gs.Positions) != 1 || gs.Positions[0].Kind != economy.PosCall || gs.Positions[0].Leverage != 5 {
		t.Fatalf("positions not preserved: %+v", gs.Positions)
	}
	if len(gs.Ledger) != 1 || gs.Ledger[0].Tokens != 24 {
		t.Fatalf("ledger not preserved: %+v", gs.Ledger)
	}
	if len(gs.PriceCandles) != 1 || gs.PriceCandles[0].Close != 1.5 || gs.PriceCandleBeats != 2 {
		t.Fatalf("price chart not preserved: candles=%+v beats=%d", gs.PriceCandles, gs.PriceCandleBeats)
	}
	if !gotEv.Get().HasFired("first-egg") {
		t.Fatal("fired event not preserved")
	}

	got2 := economy.FromState(func() economy.State { s := got.Get(); s.Tokens = 5678; return s }())
	if err := store.Write(info.ID, got2, gotEv); err != nil {
		t.Fatalf("second write: %v", err)
	}
	again, _, err := store.Read(info.ID)
	if err != nil {
		t.Fatalf("read after update: %v", err)
	}
	if again.Get().Tokens != 5678 {
		t.Fatalf("update not persisted: got %v", again.Get().Tokens)
	}
}

func TestSQLiteStoreManagesMultipleSaves(t *testing.T) {
	store := SQLiteStore{Path: filepath.Join(t.TempDir(), "save.db")}

	alphaState := economy.NewState()
	alphaState.Tokens = 100
	alpha, err := store.Create("Alpha", economy.FromState(alphaState), events.NewMachine())
	if err != nil {
		t.Fatalf("create alpha: %v", err)
	}

	betaState := economy.NewState()
	betaState.Tokens = 200
	betaState.Eggs = 10
	beta, err := store.Create("Beta", economy.FromState(betaState), events.NewMachine())
	if err != nil {
		t.Fatalf("create beta: %v", err)
	}

	saves, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(saves) != 2 {
		t.Fatalf("save count=%d, want 2", len(saves))
	}
	if savesByName(saves)["Alpha"].Tokens != 100 || savesByName(saves)["Beta"].Eggs != 10 {
		t.Fatalf("save summaries not preserved: %+v", saves)
	}

	if err := store.Rename(alpha.ID, "Gamma"); err != nil {
		t.Fatalf("rename: %v", err)
	}
	saves, err = store.List()
	if err != nil {
		t.Fatalf("list after rename: %v", err)
	}
	if _, ok := savesByName(saves)["Gamma"]; !ok {
		t.Fatalf("renamed save missing: %+v", saves)
	}

	if err := store.Delete(beta.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, err := store.Read(beta.ID); !errors.Is(err, ErrSaveNotFound) {
		t.Fatalf("read deleted save err=%v, want ErrSaveNotFound", err)
	}
	saves, err = store.List()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(saves) != 1 || saves[0].ID != alpha.ID {
		t.Fatalf("remaining saves=%+v, want only alpha", saves)
	}
}

func TestSQLiteStoreImportsLegacyFlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "save.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open legacy db: %v", err)
	}
	if _, err := db.Exec(legacySchema); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	_, err = db.Exec(`INSERT INTO flock (
        id, tokens, total_earned, per_click, eggs, peak_eggs, eggs_laid,
        eggs_bought, consumers, price_factor, owned, upgrades, transactions,
        price_candles, price_candle_beats, ledger, fired_events, last_seen
    ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		legacySaveID, 42.0, 42.0, 1.0, 3.0, 3.0, 3.0,
		0.0, 0.0, 1.0, `{"gpu":2}`, `{}`, `[]`,
		`[]`, 0, `[]`, `{"first-egg":true}`, int64(1234))
	if err != nil {
		t.Fatalf("insert legacy save: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close legacy db: %v", err)
	}

	store := SQLiteStore{Path: path}
	saves, err := store.List()
	if err != nil {
		t.Fatalf("list migrated saves: %v", err)
	}
	if len(saves) != 1 {
		t.Fatalf("migrated save count=%d, want 1", len(saves))
	}
	if saves[0].Name != "Flock 1" || saves[0].Tokens != 42 {
		t.Fatalf("bad migrated summary: %+v", saves[0])
	}
	got, gotEv, err := store.Read(saves[0].ID)
	if err != nil {
		t.Fatalf("read migrated save: %v", err)
	}
	if got.Get().Owned["gpu"] != 2 || !gotEv.Get().HasFired("first-egg") {
		t.Fatalf("legacy state not preserved: econ=%+v events=%+v", got.Get(), gotEv.Get())
	}
}

func savesByName(saves []SaveInfo) map[string]SaveInfo {
	out := make(map[string]SaveInfo, len(saves))
	for _, save := range saves {
		out[save.Name] = save
	}
	return out
}
