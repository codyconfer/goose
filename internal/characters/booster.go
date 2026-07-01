package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	boosterChance              = 0.015
	boosterEligibleTotalEarned = 8000
)

func BoosterEligible(s economy.State) bool {
	return s.TotalEarned > boosterEligibleTotalEarned
}

var boosterPitches = []string{
	"Jet Set Huang strides in, brand-new leather jacket creaking with every step. \"The more eggs you buy, the more eggs you sell. I did the math — it's a straight line, up and to the right.\" He nods at a Data Center he'll sell you, funded by the %s tokens he's about to invest in you, funded by the eggs you'll buy from him. \"Virtuous circle. Everybody wins. Sign here.\"",
	"Jet Set Huang produces a slide with one arrow eating its own tail. \"We announce that I invest %s tokens in the flock, the flock buys compute from me, I book that as revenue, my market cap goes up, so I invest more. It's not circular, it's a *flywheel*. The more you flywheel, the more you flywheel.\"",
}

type boosterContext struct {
	invest float64
}

func NewBooster(s economy.State, r *rand.Rand) Character {
	ctx := boosterContext{
		invest: (s.TokensPerSecond()*180 + 2000) * (1 + r.Float64()),
	}
	pitch := fmt.Sprintf(choose(r, boosterPitches), economy.FormatNum(ctx.invest))

	return character(Booster, "Jet Set Huang", pitch, ctx.invest,
		newFlywheelOption(ctx),
		newBuyComputeOption(ctx),
		newWheresRevenueOption(ctx),
	)
}

func newFlywheelOption(ctx boosterContext) Option {
	meltUp := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🌐 Sector-Wide Melt-Up",
				fmt.Sprintf("Huang announces the deal from a stadium stage and the entire egg complex goes vertical. %s tokens, a stampede of buyers, and egg prices rip. Everyone agrees, loudly, not to mention the word 'circular.'", economy.FormatNum(ctx.invest))),
			economy.Earn(ctx.invest), economy.GrowCrowd(1.4), economy.AddConsumers(12), economy.ShockPrice(1.3),
		)
	}
	holds := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🚀 The Flywheel Spins",
				fmt.Sprintf("The deal closes quietly and the numbers go up on schedule. %s tokens plus a wave of new buyers flood in. For now the wheel spins and nobody stops to ask who's pushing.", economy.FormatNum(ctx.invest))),
			economy.Earn(ctx.invest), economy.GrowCrowd(1.25), economy.AddConsumers(6),
		)
	}
	stalled := func(s economy.State, r *rand.Rand) Outcome {
		fee := s.Tokens * 0.1
		return outcome(
			notify.Warning("🛞 The Wheel Sticks",
				fmt.Sprintf("A journalist maps the circle and buyers hesitate. The deal half-closes; you eat %s tokens in structuring fees for a flywheel that's mostly just a wheel now.", economy.FormatNum(fee))),
			economy.Earn(ctx.invest*0.5), economy.Spend(fee), economy.GrowCrowd(0.9),
		)
	}
	ponzi := func(s economy.State, r *rand.Rand) Outcome {
		pop := s.Tokens * (0.25 + r.Float64()*0.2)
		return outcome(
			notify.Negative("💥 The Flywheel Was a Ponzi",
				fmt.Sprintf("An analyst draws the arrow eating its own tail on live TV. The deal unwinds, %s tokens evaporate, egg prices gap down, and the momentum crowd exits as fast as it arrived. The jacket, at least, still looks great.", economy.FormatNum(pop))),
			economy.Earn(ctx.invest*0.3), economy.Spend(pop), economy.GrowCrowd(0.5), economy.ShockPrice(0.8),
		)
	}
	return option(
		"Join the flywheel",
		"Sign the circular deal. The market rips today. Whether it's a flywheel or a Ponzi depends entirely on who stops spinning first.",
		pick(br(0.15, meltUp), br(0.4, holds), br(0.2, stalled), br(0.25, ponzi)),
	)
}

func newBuyComputeOption(ctx boosterContext) Option {

	base := func(markupFactor float64, tone notify.Tone, title, msgFmt string, extra ...economy.Command) resolver {
		return func(s economy.State, r *rand.Rand) Outcome {
			markup := affordable(s, ctx.invest*markupFactor)
			cmds := append([]economy.Command{economy.Spend(markup)}, extra...)
			return outcome(notify.Note(tone, title, fmt.Sprintf(msgFmt, economy.FormatNum(markup))), cmds...)
		}
	}
	fair := base(1.5, notify.ToneNeutral, "🚜 Compute Acquired",
		"You overpay %s tokens for a Data Center, and Huang high-fives you for 'securing allocation.' It does actually ship Goose Premium, so that's something.",
		economy.GrantProducer("datacenter", 1), economy.AddConsumers(4))
	lemon := base(1.5, notify.ToneWarning, "🍋 Allocation of Lemons",
		"You wire %s tokens and take delivery of a Data Center that's half decommissioned crypto-mining rigs. It honks, occasionally. Huang has already sold three more to your competitors.",
		economy.GrantProducer("server", 1))
	steal := base(1.2, notify.TonePositive, "🎯 Actually a Steal",
		"Huang, mid-fundraise and desperate for a logo, dumps a Data Center on you at cost — %s tokens — plus a Server Rack to sweeten it. For once the markup went the other way.",
		economy.GrantProducer("datacenter", 1), economy.GrantProducer("rack", 1), economy.AddConsumers(4))
	return option(
		"Just buy a data center off him",
		"Pay cash, at a markup, for compute you may not need. Real capex, real honking — real questions from the CFO later.",
		pick(br(0.6, fair), br(0.25, lemon), br(0.15, steal)),
	)
}

func newWheresRevenueOption(ctx boosterContext) Option {
	footnotes := flat(outcome(
		notify.Positive("🧾 Read the Footnotes",
			"You ask Huang to point at the revenue on his own slide. He adjusts the jacket, mumbles 'ecosystem,' and leaves. Regulars who hate getting rug-pulled quietly start shopping with you instead."),
		economy.GrowCrowd(1.06),
	))
	pumpedAnyway := func(s economy.State, r *rand.Rand) Outcome {
		gain := s.TokensPerSecond()*40 + 300
		return outcome(
			notify.Positive("📣 He Shouted You Out Anyway",
				fmt.Sprintf("On his way out Huang name-drops the flock on stage as 'the disciplined ones.' The contrarian crowd loves it and %s tokens of tasteful, sober FOMO trickle in.", economy.FormatNum(gain))),
			economy.Earn(gain), economy.GrowCrowd(1.08),
		)
	}
	return option(
		"Ask where the revenue is",
		"Decline the pump. No sugar high, no hangover. A sober goose that reads the footnotes earns a loyal crowd.",
		pick(br(0.75, footnotes), br(0.25, pumpedAnyway)),
	)
}
