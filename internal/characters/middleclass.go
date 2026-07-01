package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	middleClassChance              = 0.016
	middleClassEligibleTotalEarned = 3000
)

func MiddleClassEligible(s economy.State) bool {
	return s.TotalEarned > middleClassEligibleTotalEarned
}

var middleClassPitches = []string{
	"A delegation of regulars shuffles in, clutching mason jars of savings. \"We used to be able to afford your eggs. Now we can't afford the eggs OR the rent — and somehow our savings are... over there now.\" They gesture vaguely upward. \"What happened?\"",
	"A retiree, a teacher, and a line cook stand at the pond's edge. \"Every egg costs more, our wages don't, and the difference keeps ending up in a fund with your name near the top. We just want to understand the mechanism.\"",
}

func NewMiddleClass(s economy.State, r *rand.Rand) Character {
	pitch := choose(r, middleClassPitches)

	return character(MiddleClass, "The Squeezed Middle", pitch, 0,
		newSecuritizeOption(),
		newBailThemOutOption(),
		newBlameAlgorithmOption(),
	)
}

func newSecuritizeOption() Option {
	transferFn := func(s economy.State) float64 { return s.TokensPerSecond()*90 + s.Tokens*0.15 + 200 }
	windfall := func(s economy.State, r *rand.Rand) Outcome {
		transfer := transferFn(s)
		return outcome(
			notify.Positive("🏦 Wealth, Transferred",
				fmt.Sprintf("You bundle their pensions into an egg-backed security and keep the spread. +%s tokens. They go home poorer, you go home richer, and the system works exactly as designed — for you.", economy.FormatNum(transfer))),
			economy.Earn(transfer), economy.GrowCrowd(0.35),
		)
	}
	blowsUp := func(s economy.State, r *rand.Rand) Outcome {
		transfer := transferFn(s) * 0.6
		clawback := s.Tokens * 0.2
		return outcome(
			notify.Warning("🧨 Subprime Egg Crisis",
				fmt.Sprintf("The egg-backed securities were rated AAA by a friend of yours. When they default you still net %s tokens up front — but a %s-token bailout and the ensuing rage torch the crowd.", economy.FormatNum(transfer), economy.FormatNum(clawback))),
			economy.Earn(transfer), economy.Spend(clawback), economy.GrowCrowd(0.3),
		)
	}
	return option(
		"Securitize their savings",
		"Bundle their nest eggs into a product and sell it back to them. Enormous windfall for you; the crowd is crushed — assuming the tranche doesn't detonate.",
		pick(br(0.7, windfall), br(0.3, blowsUp)),
	)
}

func newBailThemOutOption() Option {
	loyalty := func(s economy.State, r *rand.Rand) Outcome {
		cost := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Positive("🤲 You Ate the Cost",
				fmt.Sprintf("You spend %s tokens making eggs affordable again. The middle class can breathe, they bring their neighbors, and nobody forgets who did it when times were lean.", economy.FormatNum(cost))),
			economy.Spend(cost), economy.GrowCrowd(1.3), economy.AddConsumers(10),
		)
	}
	movement := func(s economy.State, r *rand.Rand) Outcome {
		cost := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Positive("📣 It Became a Movement",
				fmt.Sprintf("Your %s-token price cut becomes a folk story. 'The goose that gives a damn' trends, and a genuine crowd — not tourists, believers — floods the pond.", economy.FormatNum(cost))),
			economy.Spend(cost), economy.GrowCrowd(1.6), economy.AddConsumers(20),
		)
	}
	moneyPit := func(s economy.State, r *rand.Rand) Outcome {
		cost := affordable(s, s.Tokens*0.2)
		return outcome(
			notify.Warning("🕳️ A Bottomless Money Pit",
				fmt.Sprintf("Good intentions meet bad unit economics. The subsidy costs %s tokens and every discount just brings more people who need discounts. Kind, expensive, and quietly unsustainable.", economy.FormatNum(cost))),
			economy.Spend(cost), economy.GrowCrowd(1.1), economy.AddConsumers(6),
		)
	}
	return option(
		"Actually help them",
		"Eat the cost, cut prices, prop up the regulars. You forgo the extraction and earn something rarer: loyalty.",
		pick(br(0.55, loyalty), br(0.2, movement), br(0.25, moneyPit)),
	)
}

func newBlameAlgorithmOption() Option {
	marketForces := flat(outcome(
		notify.Warning("🤷 Market Forces",
			"You shrug and say the algorithm decides prices now. The savings evaporate upward into a fund you don't even hold shares in. You gain nothing; the crowd thins; the top wins by default."),
		economy.GrowCrowd(0.7),
	))
	unnoticed := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Neutral("😶 Nobody Notices, For Now",
				"You blame 'the algorithm,' and — busy, tired, broke — the crowd mostly lets it slide this time. The extraction continues, quietly, above your pay grade."),
			economy.GrowCrowd(0.9),
		)
	}
	scandal := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Negative("🔥 The Op-Ed Went Viral",
				"'The Goose That Ate the Middle Class' hits the front page with your quote in the headline. The savings still evaporate upward — but now your name is stapled to it, and the crowd bolts."),
			economy.GrowCrowd(0.5),
		)
	}
	return option(
		"Blame the algorithm",
		"Cite 'market forces' and move on. Their wealth still evaporates — it just flows past you, straight to the top.",
		pick(br(0.5, marketForces), br(0.25, unnoticed), br(0.25, scandal)),
	)
}
