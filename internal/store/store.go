package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

var ErrSaveNotFound = errors.New("save not found")

type SaveInfo struct {
	ID        int64
	Name      string
	CreatedAt int64
	UpdatedAt int64
	Tokens    float64
	Eggs      float64
	Level     int
}

type Store interface {
	Create(string, *economy.Machine, *events.Machine) (SaveInfo, error)
	Read(int64) (*economy.Machine, *events.Machine, error)
	Write(int64, *economy.Machine, *events.Machine) error
	Rename(int64, string) error
	Delete(int64) error
	List() ([]SaveInfo, error)
	Exists() bool
}

func DefaultStore() Store { return DefaultSQLiteStore() }

func Load(id int64) (*economy.Machine, *events.Machine, float64, error) {
	econ, ev, err := DefaultStore().Read(id)
	if err != nil {
		return nil, nil, 0, err
	}
	var offline float64
	if last := econ.LastSeen(); last > 0 {
		offline = econ.ApplyOffline(time.Now().Unix() - last)
	}
	ev.Reconcile(econ.Get())
	return econ, ev, offline, nil
}

func CreateSave(name string, s *economy.Machine, ev *events.Machine) (SaveInfo, error) {
	name = CleanName(name)
	if name == "" {
		return SaveInfo{}, errors.New("save name cannot be empty")
	}
	s.SetLastSeen(time.Now().Unix())
	return DefaultStore().Create(name, s, ev)
}

func Save(id int64, s *economy.Machine, ev *events.Machine) error {
	if id <= 0 {
		return ErrSaveNotFound
	}
	s.SetLastSeen(time.Now().Unix())
	return DefaultStore().Write(id, s, ev)
}

func SaveExists() bool { return DefaultStore().Exists() }

func ListSaves() ([]SaveInfo, error) { return DefaultStore().List() }

func RenameSave(id int64, name string) error {
	name = CleanName(name)
	if name == "" {
		return errors.New("save name cannot be empty")
	}
	return DefaultStore().Rename(id, name)
}

func DeleteSave(id int64) error { return DefaultStore().Delete(id) }

func CleanName(name string) string { return strings.TrimSpace(name) }

func NextSaveName(saves []SaveInfo) string {
	taken := make(map[string]bool, len(saves))
	for _, s := range saves {
		taken[strings.ToLower(s.Name)] = true
	}
	for i := 1; ; i++ {
		name := fmt.Sprintf("Flock %d", i)
		if !taken[strings.ToLower(name)] {
			return name
		}
	}
}

func summarize(id int64, name string, createdAt, updatedAt int64, s economy.State) SaveInfo {
	economy.Normalize(&s)
	return SaveInfo{
		ID:        id,
		Name:      name,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Tokens:    s.Tokens,
		Eggs:      s.Eggs,
		Level:     s.Level(),
	}
}
