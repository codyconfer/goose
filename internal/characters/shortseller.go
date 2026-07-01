package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	shortSellerChance              = 0.012
	shortSellerEligibleTotalEarned = 10000
)

func ShortSellerEligible(s economy.State) bool {
	return s.TotalEarned > shortSellerEligibleTotalEarned
}

var shortSellerNames = []string{
	"Muddy Pond Research",
	"Hindengoose",
	"a pseudonymous account called @eggsposed",
	"Citron-Adjacent Capital",
}

var shortSellerPitches = []string{
	"%s drops a 100-page report titled 'THE GOOSE IS COOKED' and a thread with 4,000 retweets. \"We are short. The eggs are a Ponzi, the consumers are bots, and the goose is, on close inspection, a very confident duck. Price target: zero.\"",
	"%s publishes drone footage of your allegedly-humming data center sitting dark. \"Channel checks suggest the honking is a soundboard. We hold puts. We suggest you hold on to something.\"",
}

type shortSellerContext struct {
	name string
}

func NewShortSeller(s economy.State, r *rand.Rand) Character {
	ctx := shortSellerContext{name: choose(r, shortSellerNames)}
	pitch := fmt.Sprintf(choose(r, shortSellerPitches), ctx.name)

	return character(ShortSeller, ctx.name, pitch, 0,
		newStayTheCourseOption(ctx),
		newRebutOption(ctx),
		newSettleQuietlyOption(ctx),
	)
}

func newStayTheCourseOption(ctx shortSellerContext) Option {
	fizzles := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Neutral("🥱 The Report Fizzles",
				fmt.Sprintf("%s posts the thread on a slow news day and it sinks under a celebrity divorce. You say nothing, lose nothing, and the position quietly bleeds them.", ctx.name)),
			economy.GrowCrowd(0.98),
		)
	}
	traction := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Negative("🩸 The Short Report Lands",
				fmt.Sprintf("%s's thread hits the mainstream feeds and the 'confident duck' line becomes a meme. Prices gap down hard and half the momentum crowd is gone by lunch.", ctx.name)),
			economy.ShockPrice(0.7), economy.GrowCrowd(0.6),
		)
	}
	backfires := func(s economy.State, r *rand.Rand) Outcome {
		gain := s.TokensPerSecond()*30 + 250
		return outcome(
			notify.Positive("🚀 Short Squeeze",
				fmt.Sprintf("The report is sloppy, the faithful mobilize, and a coordinated buy-in torches %s's puts. The squeeze rips prices higher and %s tokens of vengeful FOMO rolls in.", ctx.name, economy.FormatNum(gain))),
			economy.Earn(gain), economy.ShockPrice(1.3), economy.GrowCrowd(1.25),
		)
	}
	return option(
		"Say nothing and stay the course",
		"Refuse to dignify it. Sometimes the report dies quietly; sometimes it goes viral and cracks the price; sometimes the faithful squeeze the shorts into the sun.",
		pick(br(0.35, fizzles), br(0.4, traction), br(0.25, backfires)),
	)
}

func newRebutOption(ctx shortSellerContext) Option {
	convinces := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.08)
		return outcome(
			notify.Positive("📑 Rebuttal Lands",
				fmt.Sprintf("Your %s-token counter-deck is dense, dull, and devastating. The market reads six of the ninety pages, decides you're credible, and bids the eggs back up.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.ShockPrice(1.1), economy.GrowCrowd(1.05),
		)
	}
	draw := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.08)
		return outcome(
			notify.Neutral("🤝 Fought to a Draw",
				fmt.Sprintf("You spend %s tokens rebutting and the whole thing curdles into 'both sides have a point.' No crash, no recovery — just a permanent asterisk on the flock.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.GrowCrowd(0.97),
		)
	}
	streisand := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.08)
		return outcome(
			notify.Negative("📢 Streisand Effect",
				fmt.Sprintf("Your %s-token rebuttal introduces the report to everyone who'd missed it. Now the whole world has read page 40, prices crater worse, and the crowd stampedes.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.ShockPrice(0.75), economy.GrowCrowd(0.55),
		)
	}
	return option(
		"Rebut with a 100-page counter-deck",
		"Pay to fight fire with fire. A crisp rebuttal can win back the tape — or hand the short a bigger audience than it ever had.",
		pick(br(0.4, convinces), br(0.3, draw), br(0.3, streisand)),
	)
}

func newSettleQuietlyOption(ctx shortSellerContext) Option {
	quiet := func(s economy.State, r *rand.Rand) Outcome {
		hush := affordable(s, s.Tokens*0.12)
		return outcome(
			notify.Neutral("🤐 Quietly Settled",
				fmt.Sprintf("A %s-token 'consulting arrangement' later, %s deletes the thread and covers the position. The story dies. You feel a little dirty and a lot solvent.", economy.FormatNum(hush), ctx.name)),
			economy.Spend(hush), economy.GrowCrowd(0.98),
		)
	}
	leaks := func(s economy.State, r *rand.Rand) Outcome {
		hush := affordable(s, s.Tokens*0.12)
		return outcome(
			notify.Negative("💧 The Hush Money Leaked",
				fmt.Sprintf("The %s-token payment to a short-seller ends up in a follow-up report titled 'THEY PAID US TO STOP.' It is somehow worse than the original. Prices and crowd both crater.", economy.FormatNum(hush))),
			economy.Spend(hush), economy.ShockPrice(0.8), economy.GrowCrowd(0.6),
		)
	}
	return option(
		"Settle it quietly",
		"Pay the short-seller to cover and go away. Clean if it stays a secret — a catastrophe if the check ever surfaces.",
		pick(br(0.6, quiet), br(0.4, leaks)),
	)
}
