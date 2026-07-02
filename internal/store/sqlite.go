package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/world"
	"github.com/codyconfer/goose/internal/worldgen"

	_ "modernc.org/sqlite"
)

const legacySaveID = 1

const savesSchema = `
CREATE TABLE IF NOT EXISTS saves (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT    NOT NULL COLLATE NOCASE UNIQUE,
    economy    TEXT    NOT NULL,
    events     TEXT    NOT NULL,
    world      TEXT    NOT NULL DEFAULT '',
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS saves_updated_at_idx ON saves(updated_at DESC);`

const legacySchema = `
CREATE TABLE IF NOT EXISTS flock (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    tokens       REAL    NOT NULL,
    total_earned REAL    NOT NULL,
    per_click    REAL    NOT NULL,
    eggs         REAL    NOT NULL,
    peak_eggs    REAL    NOT NULL,
    eggs_laid    REAL    NOT NULL,
    eggs_bought  REAL    NOT NULL,
    consumers    REAL    NOT NULL,
    price_factor REAL    NOT NULL,
    owned        TEXT    NOT NULL,
    upgrades     TEXT    NOT NULL,
    transactions TEXT    NOT NULL,
    price_candles TEXT   NOT NULL,
    price_candle_beats INTEGER NOT NULL,
    ledger       TEXT    NOT NULL,
    fired_events TEXT    NOT NULL,
    last_seen    INTEGER NOT NULL
);`

var legacyMigrations = []string{
	`ALTER TABLE flock ADD COLUMN positions TEXT NOT NULL DEFAULT '[]';`,
	`ALTER TABLE flock ADD COLUMN price_candles TEXT NOT NULL DEFAULT '[]';`,
	`ALTER TABLE flock ADD COLUMN price_candle_beats INTEGER NOT NULL DEFAULT 0;`,
}

type SQLiteStore struct{ Path string }

func DefaultSQLiteStore() SQLiteStore { return SQLiteStore{Path: dbPath()} }

func dbPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_save.db"
	}
	return filepath.Join(home, ".goose", "save.db")
}

func (ss SQLiteStore) open() (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(ss.Path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", ss.Path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(savesSchema); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := ensureWorldColumn(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	if err := migrateLegacyFlock(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func (ss SQLiteStore) Create(name string, s *economy.Machine, ev *events.Machine, wrld *world.State) (SaveInfo, error) {
	db, err := ss.open()
	if err != nil {
		return SaveInfo{}, err
	}
	defer func() { _ = db.Close() }()

	name = CleanName(name)
	if name == "" {
		return SaveInfo{}, errors.New("save name cannot be empty")
	}
	econJSON, eventsJSON, worldJSON, err := marshalSave(s, ev, wrld)
	if err != nil {
		return SaveInfo{}, err
	}
	now := time.Now().Unix()
	res, err := db.Exec(`INSERT INTO saves (name, economy, events, world, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`, name, econJSON, eventsJSON, worldJSON, now, now)
	if err != nil {
		return SaveInfo{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return SaveInfo{}, err
	}
	return summarize(id, name, now, now, s.Get()), nil
}

func (ss SQLiteStore) Read(id int64) (*economy.Machine, *events.Machine, *world.State, error) {
	db, err := ss.open()
	if err != nil {
		return nil, nil, nil, err
	}
	defer func() { _ = db.Close() }()

	var econJSON, eventsJSON, worldJSON string
	err = db.QueryRow(`SELECT economy, events, world FROM saves WHERE id = ?`, id).Scan(&econJSON, &eventsJSON, &worldJSON)
	if err == sql.ErrNoRows {
		return nil, nil, nil, ErrSaveNotFound
	}
	if err != nil {
		return nil, nil, nil, err
	}
	return unmarshalSave(econJSON, eventsJSON, worldJSON)
}

func (ss SQLiteStore) Write(id int64, s *economy.Machine, ev *events.Machine, wrld *world.State) error {
	db, err := ss.open()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	econJSON, eventsJSON, worldJSON, err := marshalSave(s, ev, wrld)
	if err != nil {
		return err
	}
	res, err := db.Exec(`UPDATE saves
        SET economy = ?, events = ?, world = ?, updated_at = ?
        WHERE id = ?`, econJSON, eventsJSON, worldJSON, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func (ss SQLiteStore) Rename(id int64, name string) error {
	db, err := ss.open()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	name = CleanName(name)
	if name == "" {
		return errors.New("save name cannot be empty")
	}
	res, err := db.Exec(`UPDATE saves
        SET name = ?, updated_at = ?
        WHERE id = ?`, name, time.Now().Unix(), id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func (ss SQLiteStore) Delete(id int64) error {
	db, err := ss.open()
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	res, err := db.Exec(`DELETE FROM saves WHERE id = ?`, id)
	if err != nil {
		return err
	}
	return requireAffected(res)
}

func (ss SQLiteStore) List() ([]SaveInfo, error) {
	db, err := ss.open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = db.Close() }()

	rows, err := db.Query(`SELECT id, name, created_at, updated_at, economy
        FROM saves ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []SaveInfo
	for rows.Next() {
		var id, createdAt, updatedAt int64
		var name, econJSON string
		if err := rows.Scan(&id, &name, &createdAt, &updatedAt, &econJSON); err != nil {
			return nil, err
		}
		var state economy.State
		if err := json.Unmarshal([]byte(econJSON), &state); err != nil {
			return nil, err
		}
		out = append(out, summarize(id, name, createdAt, updatedAt, state))
	}
	return out, rows.Err()
}

func (ss SQLiteStore) Exists() bool {
	db, err := ss.open()
	if err != nil {
		return false
	}
	defer func() { _ = db.Close() }()

	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM saves`).Scan(&n); err != nil {
		return false
	}
	return n > 0
}

func migrateLegacyFlock(db *sql.DB) error {
	var legacyCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'flock'`).Scan(&legacyCount); err != nil {
		return err
	}
	if legacyCount == 0 {
		return nil
	}

	var saveCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM saves`).Scan(&saveCount); err != nil {
		return err
	}
	if saveCount > 0 {
		_, err := db.Exec(`DROP TABLE flock`)
		return err
	}

	for _, mig := range legacyMigrations {
		_, _ = db.Exec(mig)
	}

	econ, ev, ok, err := readLegacyFlock(db)
	if err != nil || !ok {
		return err
	}
	econJSON, eventsJSON, worldJSON, err := marshalSave(econ, ev, worldgen.Generate(worldgen.DefaultSeed))
	if err != nil {
		return err
	}
	now := time.Now().Unix()
	updatedAt := econ.Get().LastSeen
	if updatedAt <= 0 {
		updatedAt = now
	}
	_, err = db.Exec(`INSERT INTO saves (name, economy, events, world, created_at, updated_at)
        VALUES (?, ?, ?, ?, ?, ?)`, NextSaveName(nil), econJSON, eventsJSON, worldJSON, updatedAt, updatedAt)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DROP TABLE flock`)
	return err
}

func readLegacyFlock(db *sql.DB) (*economy.Machine, *events.Machine, bool, error) {
	s := economy.NewState()
	ev := events.NewState()
	var owned, upgrades, transactions, positions, priceCandles, ledger, firedEvents string
	row := db.QueryRow(`SELECT tokens, total_earned, per_click, eggs, peak_eggs,
        eggs_laid, eggs_bought, consumers, price_factor,
        owned, upgrades, transactions, positions, price_candles, price_candle_beats,
        ledger, fired_events, last_seen
        FROM flock WHERE id = ?`, legacySaveID)
	err := row.Scan(&s.Tokens, &s.TotalEarned, &s.PerClick, &s.Eggs, &s.PeakEggs,
		&s.EggsLaid, &s.EggsBought, &s.Consumers, &s.PriceFactor,
		&owned, &upgrades, &transactions, &positions, &priceCandles, &s.PriceCandleBeats,
		&ledger, &firedEvents, &s.LastSeen)
	if err == sql.ErrNoRows {
		return nil, nil, false, nil
	}
	if err != nil {
		return nil, nil, false, err
	}

	for _, dec := range []struct {
		raw string
		dst any
	}{
		{owned, &s.Owned},
		{upgrades, &s.UpgradeLevels},
		{transactions, &s.Transactions},
		{positions, &s.Positions},
		{priceCandles, &s.PriceCandles},
		{ledger, &s.Ledger},
		{firedEvents, &ev.Fired},
	} {
		if err := json.Unmarshal([]byte(dec.raw), dec.dst); err != nil {
			return nil, nil, false, err
		}
	}

	return economy.FromState(s), events.FromState(ev), true, nil
}

func marshalSave(s *economy.Machine, ev *events.Machine, wrld *world.State) (string, string, string, error) {
	econJSON, err := json.Marshal(s.Get())
	if err != nil {
		return "", "", "", err
	}
	eventsJSON, err := json.Marshal(ev.Get())
	if err != nil {
		return "", "", "", err
	}
	worldJSON, err := json.Marshal(normalizeSQLiteWorld(wrld))
	if err != nil {
		return "", "", "", err
	}
	return string(econJSON), string(eventsJSON), string(worldJSON), nil
}

func unmarshalSave(econJSON, eventsJSON, worldJSON string) (*economy.Machine, *events.Machine, *world.State, error) {
	s := economy.NewState()
	ev := events.NewState()
	wrld := normalizeSQLiteWorld(nil)
	if err := json.Unmarshal([]byte(econJSON), &s); err != nil {
		return nil, nil, nil, err
	}
	if err := json.Unmarshal([]byte(eventsJSON), &ev); err != nil {
		return nil, nil, nil, err
	}
	if worldJSON != "" {
		if err := json.Unmarshal([]byte(worldJSON), &wrld); err != nil {
			return nil, nil, nil, err
		}
	}
	if len(wrld.Events) == 0 && len(wrld.Characters) == 0 {
		wrld = normalizeSQLiteWorld(nil)
	}
	return economy.FromState(s), events.FromState(ev), &wrld, nil
}

func requireAffected(res sql.Result) error {
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrSaveNotFound
	}
	return nil
}

func ensureWorldColumn(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(saves)`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	var (
		cid      int
		name     string
		typeName string
		notNull  int
		defaultV sql.NullString
		primaryK int
		hasWorld bool
	)
	for rows.Next() {
		if err := rows.Scan(&cid, &name, &typeName, &notNull, &defaultV, &primaryK); err != nil {
			return err
		}
		if name == "world" {
			hasWorld = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasWorld {
		return nil
	}
	_, err = db.Exec(`ALTER TABLE saves ADD COLUMN world TEXT NOT NULL DEFAULT ''`)
	return err
}

func normalizeSQLiteWorld(wrld *world.State) world.State {
	if wrld == nil || (len(wrld.Events) == 0 && len(wrld.Characters) == 0) {
		return *worldgen.Generate(worldgen.DefaultSeed)
	}
	return *wrld
}
