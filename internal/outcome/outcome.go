package outcome

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

type Outcome struct {
	Notif notify.Notification
	Cmds  []economy.Command
}

type Resolver = func(s economy.State, r *rand.Rand) Outcome

func New(n notify.Notification, cmds ...economy.Command) Outcome {
	return Outcome{Notif: n, Cmds: cmds}
}

type Branch struct {
	Weight  float64
	Resolve Resolver
}

func Br(weight float64, resolve Resolver) Branch {
	return Branch{Weight: weight, Resolve: resolve}
}

func Pick(branches ...Branch) Resolver {
	total := 0.0
	for _, b := range branches {
		if b.Weight > 0 {
			total += b.Weight
		}
	}
	return func(s economy.State, r *rand.Rand) Outcome {
		x := r.Float64() * total
		for _, b := range branches {
			if b.Weight <= 0 {
				continue
			}
			if x < b.Weight {
				return b.Resolve(s, r)
			}
			x -= b.Weight
		}
		return branches[len(branches)-1].Resolve(s, r)
	}
}
