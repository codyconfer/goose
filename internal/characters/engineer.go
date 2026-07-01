package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	engineerChance              = 0.016
	engineerEligibleTotalEarned = 4000
)

func EngineerEligible(s economy.State) bool {
	return s.TotalEarned > engineerEligibleTotalEarned
}

var engineerPitches = []string{
	"A tired engineer sets down a laptop caked in stickers. \"I ran the numbers. Profit is revenue minus costs — and the costs are, well, all of them. There's no return here. I can show you the spreadsheet. You are not going to like the spreadsheet.\"",
	"An engineer who actually did the reading clears their throat. \"The eggs don't make Goose Premium. The goose makes the Goose Premium, and the goose loses money on every unit — you're making it up on volume. Somebody had to say it out loud.\"",
	"A quiet skeptic draws an arrow on the whiteboard that eats its own tail. \"This is the business model. You buy compute from the man you sold eggs to, who invested in you. It's circular. I can pop this gently now, or the market pops it later, much harder.\"",
}

func NewEngineer(s economy.State, r *rand.Rand) Character {
	pitch := choose(r, engineerPitches)

	return character(Engineer, "The Engineer", pitch, 0,
		newFaceNumbersOption(),
		newWickerManOption(),
		newHireSkepticOption(),
	)
}

func newFaceNumbersOption() Option {
	cleanReset := func(s economy.State, r *rand.Rand) Outcome {
		froth := s.Tokens * 0.15
		return outcome(
			notify.Neutral("🧼 Clean Reset",
				fmt.Sprintf("The spreadsheet is bad but not fatal. Only %s tokens of vapor evaporate, the honest customers shrug and stay, and the engineer bolts down two real Server Racks on the way out.", economy.FormatNum(froth))),
			economy.Spend(froth), economy.GrowCrowd(0.7), economy.GrantProducer("rack", 2),
		)
	}
	hardReset := func(s economy.State, r *rand.Rand) Outcome {
		froth := s.Tokens * 0.25
		return outcome(
			notify.Warning("📉 Bubble Popped (On Purpose)",
				fmt.Sprintf("You read the spreadsheet out loud. %s tokens of imaginary value vanish and the tourists flee — but the engineer bolts down a real Server Rack and the survivors run a sober, honest flock.", economy.FormatNum(froth))),
			economy.Spend(froth), economy.GrowCrowd(0.4), economy.GrantProducer("rack", 1),
		)
	}
	return option(
		"Face the numbers",
		"Let them pop it now. The speculative froth evaporates and the hype crowd bolts — but what's left is an actual business.",
		pick(br(0.4, cleanReset), br(0.6, hardReset)),
	)
}

func newWickerManOption() Option {
	prophecyHolds := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🔮 The Prophecy Holds",
				"You call them a doomer who 'doesn't get it,' and a fresh megaround the next morning proves you gloriously right — for now. The believers become evangelists."),
			economy.GrowCrowd(1.3), economy.AddConsumers(8),
		)
	}
	crowdCheers := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("🔥 Messenger, Meet Wicker Man",
				"You call them a doomer who 'doesn't get it,' and the believers cheer louder than ever. The great prophecy holds. For now."),
			economy.GrowCrowd(1.12),
		)
	}
	correction := func(s economy.State, r *rand.Rand) Outcome {
		froth := s.Tokens * 0.35
		return outcome(
			notify.Negative("💥 The Correction Came Anyway",
				fmt.Sprintf("The subtweet aged poorly. Gravity exists, the froth burns off %s tokens, and the crowd stampedes for the exits — the pop just hurt more for being denied.", economy.FormatNum(froth))),
			economy.Spend(froth), economy.GrowCrowd(0.3),
		)
	}
	vindicated := func(s economy.State, r *rand.Rand) Outcome {
		froth := s.Tokens * 0.5
		return outcome(
			notify.Negative("⚰️ They Were Right, With Receipts",
				fmt.Sprintf("The engineer's spreadsheet leaks with a timestamp. The denial becomes the story, %s tokens vaporize, and the crowd doesn't just leave — it warns its friends.", economy.FormatNum(froth))),
			economy.Spend(froth), economy.GrowCrowd(0.25), economy.ShockPrice(0.85),
		)
	}
	return option(
		"Put them in the wicker man",
		"Discredit the messenger, keep the party going. Sometimes the timeline agrees with you — for a while.",
		pick(br(0.15, prophecyHolds), br(0.3, crowdCheers), br(0.35, correction), br(0.2, vindicated)),
	)
}

func newHireSkepticOption() Option {
	inHouse := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Neutral("🔧 Reality, In-House",
				fmt.Sprintf("For %s tokens the engineer joins up and quietly builds two Servers that actually work. The froth chasers wander off; a handful of loyal, sensible regulars move in.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.GrantProducer("server", 2), economy.GrowCrowd(0.85), economy.AddConsumers(4),
		)
	}
	tenX := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Positive("🏗️ Turns Out They're a 10x Builder",
				fmt.Sprintf("The skeptic quietly re-architects the whole operation for %s tokens and ships a Data Center that just... works. Word spreads that this goose has real engineers, and grown-up money follows.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.GrantProducer("datacenter", 1), economy.GrowCrowd(1.05), economy.AddConsumers(8),
		)
	}
	quit := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.08)
		return outcome(
			notify.Warning("🚪 Gone in a Week",
				fmt.Sprintf("The skeptic reads the internal roadmap, says 'yeah, no,' and leaves after cashing a %s-token signing bonus. You keep one Server and a memo about morale.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.GrantProducer("server", 1), economy.GrowCrowd(0.9),
		)
	}
	return option(
		"Hire the skeptic",
		"Pay them to build the real thing. You get boring and durable; the hype crowd leaves, the grown-ups stay.",
		pick(br(0.6, inHouse), br(0.2, tenX), br(0.2, quit)),
	)
}
