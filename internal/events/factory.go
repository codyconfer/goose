package events

import (
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
	out "github.com/codyconfer/goose/internal/outcome"
)

type Outcome = out.Outcome

type apply = out.Resolver

type branch = out.Branch

func outcome(n notify.Notification, cmds ...economy.Command) Outcome {
	return out.New(n, cmds...)
}

func br(weight float64, a apply) branch { return out.Br(weight, a) }

func pick(branches ...branch) apply { return out.Pick(branches...) }
