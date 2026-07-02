package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/world"
)

type FileStore struct{ Path string }

type fileSave struct {
	Economy economy.State `json:"economy"`
	Events  events.State  `json:"events"`
	World   world.State   `json:"world,omitempty"`
}

type fileSaveSlot struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	CreatedAt int64         `json:"created_at"`
	UpdatedAt int64         `json:"updated_at"`
	Economy   economy.State `json:"economy"`
	Events    events.State  `json:"events"`
	World     world.State   `json:"world,omitempty"`
}

type fileSaveCollection struct {
	NextID int64          `json:"next_id"`
	Saves  []fileSaveSlot `json:"saves"`
}

func DefaultFileStore() FileStore { return FileStore{Path: savePath()} }

func savePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_save.json"
	}
	return filepath.Join(home, ".goose", "save.json")
}

func (fs FileStore) Create(name string, s *economy.Machine, ev *events.Machine, wrld *world.State) (SaveInfo, error) {
	coll, err := fs.readCollection()
	if err != nil {
		return SaveInfo{}, err
	}
	name = CleanName(name)
	if name == "" {
		return SaveInfo{}, errors.New("save name cannot be empty")
	}
	if coll.hasName(name, 0) {
		return SaveInfo{}, errors.New("save name already exists")
	}
	now := time.Now().Unix()
	id := coll.NextID
	if id <= 0 {
		id = 1
	}
	slot := fileSaveSlot{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		Economy:   s.Get(),
		Events:    ev.Get(),
		World:     normalizeWorld(wrld),
	}
	coll.NextID = id + 1
	coll.Saves = append(coll.Saves, slot)
	if err := fs.writeCollection(coll); err != nil {
		return SaveInfo{}, err
	}
	return summarize(slot.ID, slot.Name, slot.CreatedAt, slot.UpdatedAt, slot.Economy), nil
}

func (fs FileStore) Read(id int64) (*economy.Machine, *events.Machine, *world.State, error) {
	coll, err := fs.readCollection()
	if err != nil {
		return nil, nil, nil, err
	}
	for _, slot := range coll.Saves {
		if slot.ID == id {
			wrld := slot.World
			if len(wrld.Events) == 0 && len(wrld.Characters) == 0 {
				wrld = *world.Generate(world.DefaultSeed)
			}
			return economy.FromState(slot.Economy), events.FromState(slot.Events), &wrld, nil
		}
	}
	return nil, nil, nil, ErrSaveNotFound
}

func (fs FileStore) Write(id int64, s *economy.Machine, ev *events.Machine, wrld *world.State) error {
	coll, err := fs.readCollection()
	if err != nil {
		return err
	}
	for i := range coll.Saves {
		if coll.Saves[i].ID == id {
			coll.Saves[i].Economy = s.Get()
			coll.Saves[i].Events = ev.Get()
			coll.Saves[i].World = normalizeWorld(wrld)
			coll.Saves[i].UpdatedAt = time.Now().Unix()
			return fs.writeCollection(coll)
		}
	}
	return ErrSaveNotFound
}

func (fs FileStore) Rename(id int64, name string) error {
	coll, err := fs.readCollection()
	if err != nil {
		return err
	}
	name = CleanName(name)
	if name == "" {
		return errors.New("save name cannot be empty")
	}
	if coll.hasName(name, id) {
		return errors.New("save name already exists")
	}
	for i := range coll.Saves {
		if coll.Saves[i].ID == id {
			coll.Saves[i].Name = name
			coll.Saves[i].UpdatedAt = time.Now().Unix()
			return fs.writeCollection(coll)
		}
	}
	return ErrSaveNotFound
}

func (fs FileStore) Delete(id int64) error {
	coll, err := fs.readCollection()
	if err != nil {
		return err
	}
	for i := range coll.Saves {
		if coll.Saves[i].ID == id {
			coll.Saves = append(coll.Saves[:i], coll.Saves[i+1:]...)
			return fs.writeCollection(coll)
		}
	}
	return ErrSaveNotFound
}

func (fs FileStore) List() ([]SaveInfo, error) {
	coll, err := fs.readCollection()
	if err != nil {
		return nil, err
	}
	out := make([]SaveInfo, 0, len(coll.Saves))
	for _, slot := range coll.Saves {
		out = append(out, summarize(slot.ID, slot.Name, slot.CreatedAt, slot.UpdatedAt, slot.Economy))
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].UpdatedAt == out[j].UpdatedAt {
			return out[i].ID > out[j].ID
		}
		return out[i].UpdatedAt > out[j].UpdatedAt
	})
	return out, nil
}

func (fs FileStore) Exists() bool {
	coll, err := fs.readCollection()
	return err == nil && len(coll.Saves) > 0
}

func (fs FileStore) readCollection() (fileSaveCollection, error) {
	data, err := os.ReadFile(fs.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return fileSaveCollection{NextID: 1}, nil
		}
		return fileSaveCollection{}, err
	}
	var coll fileSaveCollection
	if err := json.Unmarshal(data, &coll); err == nil && (len(coll.Saves) > 0 || coll.NextID > 0) {
		coll.normalize()
		return coll, nil
	}

	legacy := fileSave{Economy: economy.NewState(), Events: events.NewState()}
	if err := json.Unmarshal(data, &legacy); err != nil {
		return fileSaveCollection{}, err
	}
	now := time.Now().Unix()
	if legacy.Economy.LastSeen > 0 {
		now = legacy.Economy.LastSeen
	}
	coll = fileSaveCollection{
		NextID: 2,
		Saves: []fileSaveSlot{{
			ID:        1,
			Name:      NextSaveName(nil),
			CreatedAt: now,
			UpdatedAt: now,
			Economy:   legacy.Economy,
			Events:    legacy.Events,
			World:     normalizeWorld(&legacy.World),
		}},
	}
	coll.normalize()
	return coll, nil
}

func (fs FileStore) writeCollection(coll fileSaveCollection) error {
	if err := os.MkdirAll(filepath.Dir(fs.Path), 0o755); err != nil {
		return err
	}
	coll.normalize()
	data, err := json.MarshalIndent(coll, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(fs.Path, data, 0o644)
}

func (coll *fileSaveCollection) normalize() {
	if coll.NextID <= 0 {
		coll.NextID = 1
	}
	var maxID int64
	for i := range coll.Saves {
		if coll.Saves[i].ID > maxID {
			maxID = coll.Saves[i].ID
		}
	}
	if coll.NextID <= maxID {
		coll.NextID = maxID + 1
	}
}

func (coll fileSaveCollection) hasName(name string, exceptID int64) bool {
	name = strings.ToLower(CleanName(name))
	for _, slot := range coll.Saves {
		if slot.ID != exceptID && strings.ToLower(slot.Name) == name {
			return true
		}
	}
	return false
}

func normalizeWorld(wrld *world.State) world.State {
	if wrld == nil || (len(wrld.Events) == 0 && len(wrld.Characters) == 0) {
		return *world.Generate(world.DefaultSeed)
	}
	return *wrld
}
