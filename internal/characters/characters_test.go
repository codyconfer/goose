package characters

import (
	"math"
	"math/rand"
	"testing"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/world"
)

func richState() economy.State {
	s := economy.NewState()
	s.Tokens = 50_000
	s.TotalEarned = 50_000
	s.Owned["server"] = 3
	s.Consumers = 10
	return s
}

func TestGeneratedCharactersBuildAndResolve(t *testing.T) {
	wrld := world.Generate(11)
	s := richState()
	for _, spec := range wrld.Characters {
		if !spec.Eligible(s) {
			continue
		}
		ch := spec.Build(s)
		if ch.Headline == "" || ch.Pitch == "" || len(ch.Options) == 0 {
			t.Fatalf("malformed character %q: %+v", spec.Key, ch)
		}
		for i, opt := range ch.Options {
			out := opt.Resolve(s, rand.New(rand.NewSource(int64(i+1))))
			if out.Notif.Title == "" || out.Notif.Message == "" {
				t.Fatalf("character %q option %d returned empty notification", spec.Key, i)
			}
			m := economy.FromState(s)
			m.ApplyWindfall(out.Notif.Title, out.Cmds)
			if got := m.Get().Tokens; math.IsNaN(got) || got < 0 {
				t.Fatalf("character %q option %d produced invalid tokens %v", spec.Key, i, got)
			}
		}
	}
}
