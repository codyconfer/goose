package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	politicianChance              = 0.014
	politicianEligibleTotalEarned = 2000
	politicianBaseFreeze          = 25.0
)

func PoliticianEligible(s economy.State) bool {
	return s.TotalEarned > politicianEligibleTotalEarned
}

var politicianNames = []string{
	"Senator Cornelius Wattlesworth",
	"Councilwoman Prudence Gaggle",
	"Rep. Buck Mallard (Subcommittee on Poultry Futures)",
	"Regulator-General Vex",
	"Mayor Dwight Pinfeather",
}

var politicianReasons = []string{
	"the eggs are suspected of being unregistered securities",
	"a subcommittee formed this morning needs to understand what a goose is",
	"the flock was never licensed to honk after 5pm",
	"the eggs might be addictive to children",
	"a national-security review of your foreign egg buyers",
	"the goose failed to file Form GOOSE-1099 in triplicate",
	"a senator watched a documentary about geese and has Concerns",
	"the data center is somehow zoned residential",
	"the eggs are, and I quote, 'too round'",
}

type politicianContext struct {
	name   string
	reason string
}

func NewPolitician(s economy.State, r *rand.Rand) Character {
	ctx := politicianContext{
		name:   choose(r, politicianNames),
		reason: choose(r, politicianReasons),
	}

	pitch := fmt.Sprintf(
		"%s arrives with a clipboard and three news cameras. \"On behalf of a subcommittee I invented this morning, this operation is SUSPENDED pending review. The stated concern: %s. We'll be in touch. Or we won't.\"",
		ctx.name, ctx.reason,
	)

	return character(Politician, ctx.name, pitch, 0,
		newComplyOption(ctx),
		newLawyerUpOption(ctx),
		newGreaseWheelsOption(ctx),
	)
}

func newComplyOption(ctx politicianContext) Option {
	quick := func(s economy.State, r *rand.Rand) Outcome {
		dur := 8 + r.Float64()*7
		return outcome(
			notify.Neutral("🏛️ Cleared Quickly",
				fmt.Sprintf("%s gets bored fast — the report is mostly pictures. Business halts ~%.0fs, the cameras leave, and the public loves a cooperative goose.", ctx.name, dur)),
			economy.Freeze(dur, ctx.reason), economy.GrowCrowd(1.08),
		)
	}
	standard := func(s economy.State, r *rand.Rand) Outcome {
		dur := politicianBaseFreeze + r.Float64()*15
		return outcome(
			notify.Warning("🏛️ Under Review",
				fmt.Sprintf("You padlock the pond and cooperate. Business halts for ~%.0fs while %s reads a report they won't finish. The public, at least, roots for the little goose.", dur, ctx.name)),
			economy.Freeze(dur, ctx.reason), economy.GrowCrowd(1.03),
		)
	}
	dragsOn := func(s economy.State, r *rand.Rand) Outcome {
		dur := politicianBaseFreeze + 25 + r.Float64()*20
		return outcome(
			notify.Warning("🐌 The Review Drags On",
				fmt.Sprintf("The subcommittee discovers per-diems and schedules 'further hearings.' The pond stays padlocked ~%.0fs and even the sympathetic crowd starts checking other ponds.", dur)),
			economy.Freeze(dur, ctx.reason), economy.GrowCrowd(0.95),
		)
	}
	return option(
		"Comply with the review",
		"Shut the doors, cooperate fully, wait it out. No fine — but the flock ships no Goose Premium until the subcommittee gets bored.",
		pick(br(0.35, quick), br(0.45, standard), br(0.2, dragsOn)),
	)
}

func newLawyerUpOption(ctx politicianContext) Option {
	countersuit := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.12)
		award := s.TokensPerSecond()*60 + 500
		return outcome(
			notify.Positive("⚖️ Countersuit Lands",
				fmt.Sprintf("Your lawyers don't just win — they get the harassment claim to stick. After %s tokens in fees you net a %s-token settlement and the crowd cheers the underdog.", economy.FormatNum(fee), economy.FormatNum(award))),
			economy.Spend(fee), economy.Earn(award), economy.Freeze(2, ctx.reason), economy.GrowCrowd(1.15),
		)
	}
	dismissed := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.12)
		return outcome(
			notify.Neutral("⚖️ Case Dismissed",
				fmt.Sprintf("Your lawyers point out the subcommittee has no jurisdiction over waterfowl. %s tokens in fees, a token 4s of 'processing,' and you're back in business.", economy.FormatNum(fee))),
			economy.Spend(fee), economy.Freeze(4, ctx.reason),
		)
	}
	angered := func(s economy.State, r *rand.Rand) Outcome {
		fee := affordable(s, s.Tokens*0.12)
		dur := politicianBaseFreeze + 20 + r.Float64()*20
		return outcome(
			notify.Negative("⚖️ You Angered the Committee",
				fmt.Sprintf("Fighting it was 'not a good look.' You eat %s tokens in fees AND a made-to-order ~%.0fs shutdown. Should've brought a muffin basket.", economy.FormatNum(fee), dur)),
			economy.Spend(fee), economy.Freeze(dur, ctx.reason),
		)
	}
	return option(
		"Lawyer up",
		"Pay real legal fees to fight the order. Win and you barely miss a beat; lose and the committee makes an example of you.",
		pick(br(0.15, countersuit), br(0.4, dismissed), br(0.45, angered)),
	)
}

func newGreaseWheelsOption(ctx politicianContext) Option {
	greased := func(s economy.State, r *rand.Rand) Outcome {
		donation := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Neutral("🤝 Wheels Greased",
				fmt.Sprintf("A %s-token 'contribution' later, the concern is expedited into oblivion and you never miss a beat. A few regulars notice how that worked and think less of you.", economy.FormatNum(donation))),
			economy.Spend(donation), economy.GrowCrowd(0.92),
		)
	}
	leaked := func(s economy.State, r *rand.Rand) Outcome {
		donation := affordable(s, s.Tokens*0.15)
		return outcome(
			notify.Warning("📰 It Leaked to a Reporter",
				fmt.Sprintf("The review vanishes, but so does the paper trail's secrecy. %s tokens gone, no shutdown — and a mid-size scandal that peels off the crowd.", economy.FormatNum(donation))),
			economy.Spend(donation), economy.GrowCrowd(0.75),
		)
	}
	sting := func(s economy.State, r *rand.Rand) Outcome {
		donation := affordable(s, s.Tokens*0.15)
		dur := politicianBaseFreeze + 25 + r.Float64()*20
		return outcome(
			notify.Negative("🚨 It Was a Sting",
				fmt.Sprintf("The 'senator' was wearing a wire. You lose %s tokens, eat a ~%.0fs shutdown, and the crowd flees the perp-walk footage. Democracy!", economy.FormatNum(donation), dur)),
			economy.Spend(donation), economy.Freeze(dur, ctx.reason), economy.GrowCrowd(0.6),
		)
	}
	return option(
		"Make a 'campaign contribution'",
		"Grease the wheels. Usually the review quietly evaporates — unless it leaks, or turns out to be a sting, in which case, oof.",
		pick(br(0.55, greased), br(0.2, leaked), br(0.25, sting)),
	)
}
