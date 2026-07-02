package characters

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	out "github.com/codyconfer/goose/internal/outcome"
	"github.com/codyconfer/goose/internal/world"
)

type Outcome = out.Outcome

type Option = world.ResolvedOption

type Character = world.ResolvedCharacter

func Roll(catalog []world.Character, s economy.State, r *rand.Rand) (Character, bool) {
	mult := s.Settings.EventMult()
	for _, i := range r.Perm(len(catalog)) {
		spec := catalog[i]
		if !spec.Eligible(s) {
			continue
		}
		if r.Float64() < spec.Chance*mult {
			return spec.Build(s), true
		}
	}
	return Character{}, false
}
