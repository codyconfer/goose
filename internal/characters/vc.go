package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	vcChance              = 0.021
	vcEligibleTotalEarned = 1000
)

func VCEligible(s economy.State) bool {
	return s.TotalEarned > vcEligibleTotalEarned
}

var vcNames = []string{
	"Jerry Tan of But Why? Combinator",
	"Jerry Tan of But Why? Combinator",
	"Marc Andregoosen of a16zzz",
	"Chamath Palihonkitiya of a blank-check SPAC",
	"Masa's Golden Egg Factory SPV",
	"Vinod Cooshla, thesis-first",
}

var vcPitches = []string{
	"%s waddles in, quarter-zip vest and all. \"Love the flock. Love the honk. We're not investing in eggs — we're investing in the *goose*. I'll wire %s tokens at a valuation we both agree to never justify. Let's talk terms.\"",
	"%s slides into the nest, already mid-thread. \"Everyone's crying about ROI. Boring. Nobody asked the printing press for a spreadsheet. I'll drop %s tokens now and mark it up 10x next round — that's just good hygiene.\"",
	"%s appears, phone out. \"This is a category-defining flock. Pre-seed, pre-revenue, pre-product, honestly pre-goose. Doesn't matter. Here's %s tokens. We'll figure out what it does after the term sheet.\"",
}

type vcContext struct {
	name   string
	invest float64
}

type vcOptionFactory func(ctx vcContext) Option

var vcOptionFactories = []vcOptionFactory{
	newTakeMoneyOption,
	newCounterOption,
	newDeclineOption,
}

func NewVC(s economy.State, r *rand.Rand) Character {
	scale := s.TokensPerSecond()*120 + 500
	ctx := vcContext{
		name:   choose(r, vcNames),
		invest: scale * (1 + r.Float64()),
	}
	pitch := fmt.Sprintf(choose(r, vcPitches), ctx.name, economy.FormatNum(ctx.invest))

	options := make([]Option, len(vcOptionFactories))
	for i, factory := range vcOptionFactories {
		options[i] = factory(ctx)
	}

	return character(VC, ctx.name, pitch, ctx.invest, options...)
}

func newTakeMoneyOption(ctx vcContext) Option {
	cleanDeal := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🤝 Deal Closed",
				fmt.Sprintf("%s wires %s tokens and tweets that they were 'early to the goose.' A wave of momentum-chasers piles in. Line goes up.", ctx.name, economy.FormatNum(ctx.invest))),
			economy.Earn(ctx.invest), economy.GrowCrowd(1.25), economy.AddConsumers(5),
		)
	}

	dilutive := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Neutral("🧷 Dirty Term Sheet",
				fmt.Sprintf("%s wires the %s tokens — behind a 3x liquidation stack, a board seat, and full-ratchet anti-dilution. You took the money; you also took a leash.", ctx.name, economy.FormatNum(ctx.invest))),
			economy.Earn(ctx.invest), economy.GrowCrowd(0.95),
		)
	}
	liquidation := func(s economy.State, r *rand.Rand) Outcome {
		cmds := []economy.Command{economy.Earn(ctx.invest), economy.Spend(ctx.invest)}
		if p, ok := s.BestProducer(); ok {
			cmds = append(cmds, economy.SeizeBest())
			return outcome(
				notify.Negative("🦈 Liquidation Preference",
					fmt.Sprintf("Page 40 of the term sheet: %s clawed back all %s tokens AND took your %s as a '2x participating preferred.' You have been value-added.", ctx.name, economy.FormatNum(ctx.invest), p.Name)),
				cmds...,
			)
		}
		return outcome(
			notify.Warning("🦈 Liquidation Preference",
				fmt.Sprintf("The wire clears, then reverses. %s 'de-risked their position' and took every one of the %s tokens back. You got a LinkedIn post out of it.", ctx.name, economy.FormatNum(ctx.invest))),
			cmds...,
		)
	}
	return option(
		"Sign the term sheet",
		"Take the check now. Growth solves everything, or so the deck says. Money this easy always has fine print.",
		pick(br(0.55, cleanDeal), br(0.20, dilutive), br(0.25, liquidation)),
	)
}

func newCounterOption(ctx vcContext) Option {
	fomoWin := func(s economy.State, r *rand.Rand) Outcome {
		big := ctx.invest * 2
		return outcome(
			notify.Positive("💼 FOMO Wins",
				fmt.Sprintf("%s panics that a rival fund might get the allocation, doubles to %s tokens, and throws in a Server Rack to 'help you scale into the vision.' Respect.", ctx.name, economy.FormatNum(big))),
			economy.Earn(big), economy.GrantProducer("rack", 1),
		)
	}
	metAtPar := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Neutral("🤏 Split the Difference",
				fmt.Sprintf("%s won't do double, but blinks and matches the original %s tokens 'to keep it moving.' Nobody's thrilled, everybody's funded.", ctx.name, economy.FormatNum(ctx.invest))),
			economy.Earn(ctx.invest),
		)
	}
	walked := func(s economy.State, r *rand.Rand) Outcome {
		loss := s.Tokens * 0.15
		return outcome(
			notify.Negative("🚪 Down-Round Whispers",
				fmt.Sprintf("%s laughs, walks, and posts a subtweet about 'founders who don't understand fundraising markets.' You bleed %s tokens and half the crowd reads it.", ctx.name, economy.FormatNum(loss))),
			economy.Spend(loss), economy.GrowCrowd(0.6),
		)
	}
	return option(
		"Counter at a stupid valuation",
		"Ask for double at a number no one can defend. VCs respect confidence over cash flow — or they ghost you forever.",
		pick(br(0.35, fomoWin), br(0.30, metAtPar), br(0.35, walked)),
	)
}

func newDeclineOption(ctx vcContext) Option {
	respected := flat(outcome(
		notify.Neutral("🪿 Stayed Independent",
			fmt.Sprintf("You ask %s a single question about revenue and they suddenly have another meeting. The flock stays yours — and the locals respect a goose that didn't take the money.", ctx.name)),
		economy.GrowCrowd(1.05),
	))
	impressed := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🌱 Warm Intro Instead",
				fmt.Sprintf("Charmed that you didn't grovel, %s intros you to a few actual customers instead of writing a check. Real demand walks in the door.", ctx.name)),
			economy.GrowCrowd(1.12), economy.AddConsumers(6),
		)
	}
	blacklisted := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Warning("🧊 Frozen Out",
				fmt.Sprintf("%s takes the 'no' personally and mentions on three podcasts that your flock 'isn't fundable.' The tourists get spooked; the believers stay.", ctx.name)),
			economy.GrowCrowd(0.9),
		)
	}
	return option(
		"Ask them to define ROI",
		"Stay independent. Actually ask what the money is for. No dilution — and the flock keeps its cap table clean.",
		pick(br(0.7, respected), br(0.18, impressed), br(0.12, blacklisted)),
	)
}
