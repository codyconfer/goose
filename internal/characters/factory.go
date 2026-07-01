package characters

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
	out "github.com/codyconfer/goose/internal/outcome"
)

type Outcome = out.Outcome

type resolver = out.Resolver

type branch = out.Branch

func outcome(n notify.Notification, cmds ...economy.Command) Outcome {
	return out.New(n, cmds...)
}

func option(label, desc string, resolve resolver) Option {
	return Option{Label: label, Desc: desc, Resolve: resolve}
}

func character(t Type, name, pitch string, stakes float64, opts ...Option) Character {
	return Character{Type: t, Name: name, Pitch: pitch, Stakes: stakes, Options: opts}
}

func br(weight float64, resolve resolver) branch { return out.Br(weight, resolve) }

func pick(branches ...branch) resolver { return out.Pick(branches...) }

func chance(p float64, win, lose resolver) resolver {
	return pick(br(p, win), br(1-p, lose))
}

func flat(o Outcome) resolver {
	return func(economy.State, *rand.Rand) Outcome { return o }
}

func choose(r *rand.Rand, xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	return xs[r.Intn(len(xs))]
}

func affordable(s economy.State, cost float64) float64 {
	if cost > s.Tokens {
		return s.Tokens
	}
	if cost < 0 {
		return 0
	}
	return cost
}
