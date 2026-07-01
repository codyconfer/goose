package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	analystChance              = 0.015
	analystEligibleTotalEarned = 6000
)

func AnalystEligible(s economy.State) bool {
	return s.TotalEarned > analystEligibleTotalEarned
}

var analystNames = []string{
	"Chip Dunkford, Managing Director of Egg Equity Research",
	"Tabitha Quill, sell-side and proud",
	"Dr. Preston Vane, PhD in Narrative",
	"the ax on the goose sector",
}

var analystPitches = []string{
	"%s clears his throat and flips to a slide with a single enormous number. \"I'm initiating coverage. My price target on your eggs is, and I want to be conservative here, absurd. Give me the exclusive and I'll make the tape agree with me by Tuesday.\"",
	"%s taps a laser pointer at a hockey-stick chart with no y-axis. \"Look, I don't need the eggs to make money. I need the *story* to make sense to a portfolio manager in a hurry. Feed me a narrative and I'll feed you a re-rating.\"",
}

type analystContext struct {
	name string
}

func NewAnalyst(s economy.State, r *rand.Rand) Character {
	ctx := analystContext{name: choose(r, analystNames)}
	pitch := fmt.Sprintf(choose(r, analystPitches), ctx.name)

	return character(Analyst, ctx.name, pitch, 0,
		newFeedNarrativeOption(ctx),
		newBuyCoverageOption(ctx),
		newReadTheFilingOption(ctx),
	)
}

func newFeedNarrativeOption(ctx analystContext) Option {
	strongBuy := func(s economy.State, r *rand.Rand) Outcome {
		gain := s.TokensPerSecond()*45 + 400
		return outcome(
			notify.Positive("📈 Initiated at Strong Buy",
				fmt.Sprintf("%s publishes a note that reads like fan fiction and the tape obeys. Egg prices re-rate, %s tokens of momentum money piles in, and nobody checks the footnotes.", ctx.name, economy.FormatNum(gain))),
			economy.Earn(gain), economy.ShockPrice(1.25), economy.GrowCrowd(1.2),
		)
	}
	hold := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Neutral("😐 Initiated at Hold",
				fmt.Sprintf("%s hedges every sentence and lands on 'Hold, constructive long-term.' Nothing moves. You gave up the exclusive for a shrug in PDF form.", ctx.name)),
			economy.GrowCrowd(0.99),
		)
	}
	sell := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Negative("📉 Slapped With a Sell",
				fmt.Sprintf("You gave %s too much access and they actually did the model. The Sell note quotes your own numbers back at you; prices crack and the tourists sprint for the door.", ctx.name)),
			economy.ShockPrice(0.8), economy.GrowCrowd(0.7),
		)
	}
	return option(
		"Feed them the narrative",
		"Give the analyst the exclusive and total access. A good rating moves the whole tape — but if they run the model, they can just as easily crater it.",
		pick(br(0.45, strongBuy), br(0.3, hold), br(0.25, sell)),
	)
}

func newBuyCoverageOption(ctx analystContext) Option {
	pump := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.1)
		return outcome(
			notify.Positive("💵 Sponsored Coverage",
				fmt.Sprintf("A %s-token 'research retainer' buys a glowing initiation with a target nobody will remember being wrong. Prices pop and the crowd swells before the disclaimer loads.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.ShockPrice(1.15), economy.GrowCrowd(1.15),
		)
	}
	scandal := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.1)
		return outcome(
			notify.Warning("🕵️ 'Paid For By' In 4pt Font",
				fmt.Sprintf("Someone screenshots the disclosure. The %s-token retainer becomes the story, the rating gets retracted, and the crowd that piled in on it piles right back out.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.ShockPrice(0.92), economy.GrowCrowd(0.8),
		)
	}
	return option(
		"Buy an ad in their newsletter",
		"Quietly retain the analyst for 'coverage.' A reliable pump — right up until someone reads the disclosure in the footer.",
		pick(br(0.6, pump), br(0.4, scandal)),
	)
}

func newReadTheFilingOption(ctx analystContext) Option {
	respected := flat(outcome(
		notify.Neutral("📚 Told Them to Read the 10-K",
			fmt.Sprintf("You hand %s the actual filing and decline to write their narrative for them. They leave insulted; the handful of investors who read filings quietly respect the flock.", ctx.name)),
		economy.GrowCrowd(1.04),
	))
	spiteful := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Warning("😤 Snubbed and Salty",
				fmt.Sprintf("%s takes the brush-off personally and puts out a lukewarm 'valuation stretched' note. Nothing catastrophic — just enough chill to thin the tourists.", ctx.name)),
			economy.ShockPrice(0.96), economy.GrowCrowd(0.92),
		)
	}
	return option(
		"Tell them to read the 10-K",
		"Refuse to script the research. No pump, no dilution of the truth — just a flock that lets the numbers speak.",
		pick(br(0.78, respected), br(0.22, spiteful)),
	)
}
