package game

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type clockConfig struct {
	Timezone string `json:"timezone"`
	Label    string `json:"label"`
}

var uiClock = clockConfig{Timezone: "America/New_York", Label: "NYC"}

var clockZoneChoices = []clockConfig{
	{Timezone: "America/New_York", Label: "NYC"},
	{Timezone: "America/Chicago", Label: "Chicago"},
	{Timezone: "America/Denver", Label: "Denver"},
	{Timezone: "America/Los_Angeles", Label: "LA"},
	{Timezone: "Europe/London", Label: "London"},
	{Timezone: "Europe/Berlin", Label: "Berlin"},
	{Timezone: "Asia/Dubai", Label: "Dubai"},
	{Timezone: "Asia/Kolkata", Label: "Mumbai"},
	{Timezone: "Asia/Singapore", Label: "Singapore"},
	{Timezone: "Asia/Hong_Kong", Label: "Hong Kong"},
	{Timezone: "Asia/Shanghai", Label: "Shanghai"},
	{Timezone: "Asia/Tokyo", Label: "Tokyo"},
	{Timezone: "Australia/Sydney", Label: "Sydney"},
	{Timezone: "UTC", Label: "UTC"},
}

func currentClockIndex() int {
	for i, c := range clockZoneChoices {
		if c.Timezone == uiClock.Timezone {
			return i
		}
	}
	return 0
}

func setClockZone(idx int) {
	if idx < 0 || idx >= len(clockZoneChoices) {
		return
	}
	uiClock = clockZoneChoices[idx]
	_ = saveClockConfig()
}

func clockConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "goose_clock.json"
	}
	return filepath.Join(home, ".goose", "clock.json")
}

func loadClockConfig() {
	uiClock = clockConfig{Timezone: "America/New_York", Label: "NYC"}
	data, err := os.ReadFile(clockConfigPath())
	if err != nil {
		return
	}
	var cfg clockConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return
	}
	if cfg.Timezone == "" {
		return
	}
	if _, err := time.LoadLocation(cfg.Timezone); err != nil {
		return
	}
	uiClock = cfg
}

func saveClockConfig() error {
	path := clockConfigPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(uiClock, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func clockZone() (string, *time.Location) {
	loc, err := time.LoadLocation(uiClock.Timezone)
	if err != nil {
		loc, _ = time.LoadLocation("America/New_York")
	}
	label := strings.TrimSpace(uiClock.Label)
	if label == "" {
		label = zoneLeaf(uiClock.Timezone)
	}
	return label, loc
}

func zoneLeaf(tz string) string {
	leaf := tz
	if i := strings.LastIndex(tz, "/"); i >= 0 {
		leaf = tz[i+1:]
	}
	return strings.ReplaceAll(leaf, "_", " ")
}
