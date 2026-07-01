package characters

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

const (
	wookChance              = 0.018
	wookEligibleTotalEarned = 400
)

func WookEligible(s economy.State) bool {
	return s.TotalEarned > wookEligibleTotalEarned
}

var wookPitches = []string{
	"A barefoot wook drifts in smelling of palo santo and cold brew. \"Bro. Bro. What if the eggs were *on-chain*? Every egg is an NFT, the goose is a DAO, we tokenize the honk itself. I need %s tokens and, like, total creative control.\"",
	"A guy with a melting smartwatch corners the flock. \"Forget selling eggs. We sell the *idea* of eggs. Pre-sell Goose Premium the goose hasn't shipped yet — infinite Goose Premium, it's basically a stablecoin. Give me %s tokens, we'll be non-custodial about it.\"",
	"A visionary uncaps a dying marker. \"Eggs are legacy. We pivot the goose to AGI. The goose *becomes* the model, the eggs become inference, we raise at a quadrillion. It's %s tokens to start believing. Don't look at the process, just trust the process.\"",
	"Someone in five-fingered shoes leans in. \"Data centers in the desert. Solar. We mine tokens by day, ship Goose Premium by night, and the whole thing is a company town called Goose Valley. First step is you giving me %s tokens and signing this napkin.\"",
}

type wookContext struct {
	ask float64
}

func NewWook(s economy.State, r *rand.Rand) Character {
	ctx := wookContext{
		ask: (s.TokensPerSecond()*40 + 250) * (1 + r.Float64()),
	}
	pitch := fmt.Sprintf(choose(r, wookPitches), economy.FormatNum(ctx.ask))

	return character(Wook, "The Wook", pitch, ctx.ask,
		newLetHimCookOption(ctx),
		newHumorHimOption(ctx),
		newBanFromPondOption(ctx),
	)
}

func newLetHimCookOption(ctx wookContext) Option {
	moonshot := func(s economy.State, r *rand.Rand) Outcome {
		win := ctx.ask * (5 + r.Float64()*5)
		return outcome(
			notify.Positive("🚀 The Maniac Was Right",
				fmt.Sprintf("The honk-token doesn't just pump, it *ascends*. Two sovereign funds and a late-night host declare it civilization-defining and %s tokens land before anyone reads the whitepaper.", economy.FormatNum(win))),
			economy.Earn(win), economy.GrowCrowd(1.6), economy.AddConsumers(15),
		)
	}
	worked := func(s economy.State, r *rand.Rand) Outcome {
		win := ctx.ask * (2 + r.Float64()*2)
		return outcome(
			notify.Positive("🌈 It Somehow Worked",
				fmt.Sprintf("Against all reason the honk-token pumps, a podcast calls it 'the future,' and the flock nets %s tokens before anyone asks what it does.", economy.FormatNum(win))),
			economy.Earn(win), economy.GrowCrowd(1.3), economy.AddConsumers(8),
		)
	}
	derailed := func(s economy.State, r *rand.Rand) Outcome {
		burn := s.Tokens * (0.2 + r.Float64()*0.25)
		return outcome(
			notify.Negative("🌀 Fully Derailed",
				fmt.Sprintf("Three weeks later the goose is a DAO with no product, %s tokens are gone, and the customers wandered off during the second all-hands about 'the mission.'", economy.FormatNum(burn))),
			economy.Spend(burn), economy.GrowCrowd(0.65),
		)
	}
	rugged := func(s economy.State, r *rand.Rand) Outcome {
		burn := s.Tokens * (0.45 + r.Float64()*0.25)
		return outcome(
			notify.Negative("🪤 Rugged By Your Own Wook",
				fmt.Sprintf("The wook mints the honk-token, drains the liquidity, and DMs you a sunset from a jurisdiction with no extradition. %s tokens gone; the crowd learns the word 'exit scam.'", economy.FormatNum(burn))),
			economy.Spend(burn), economy.GrowCrowd(0.45),
		)
	}
	return option(
		"Let him cook",
		"Chase the vision with both wings. Almost always burns the treasury — but every so often the maniac is right, and once in a while he's right about everything.",
		pick(br(0.05, moonshot), br(0.15, worked), br(0.55, derailed), br(0.25, rugged)),
	)
}

func newHumorHimOption(ctx wookContext) Option {
	contained := func(s economy.State, r *rand.Rand) Outcome {
		tab := s.Tokens * 0.02
		return outcome(
			notify.Neutral("☕ Politely Contained",
				fmt.Sprintf("You 'circle back' the wook into the void for the price of a %s-token oat latte. The flock loses an afternoon but keeps its eggs.", economy.FormatNum(tab))),
			economy.Spend(tab), economy.GrowCrowd(0.97),
		)
	}
	goodIdea := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Positive("💡 One Good Idea, By Accident",
				"Somewhere in the ninety-minute monologue the wook says one genuinely smart thing about distribution. You quietly ship it and pocket the upside while he's still talking."),
			economy.GrowCrowd(1.08), economy.AddConsumers(3),
		)
	}
	wontLeave := func(s economy.State, r *rand.Rand) Outcome {
		tab := s.Tokens * 0.06
		return outcome(
			notify.Warning("🛋️ He Moved Into the Nest",
				fmt.Sprintf("The wook mistakes the free coffee for a co-founder offer and squats in the break room for a week. %s tokens of snacks and vibes, gone.", economy.FormatNum(tab))),
			economy.Spend(tab), economy.GrowCrowd(0.95),
		)
	}
	return option(
		"Nod and expense a coffee",
		"Hear him out, buy him a drink, commit to nothing. Cheap tuition in not-listening-to-wooks — usually.",
		pick(br(0.7, contained), br(0.15, goodIdea), br(0.15, wontLeave)),
	)
}

func newBanFromPondOption(ctx wookContext) Option {
	focus := flat(outcome(
		notify.Positive("🚫 Focus Restored",
			"You escort the visionary back to the parking lot. The flock returns to the radical strategy of shipping Goose Premium and selling it, and word gets around that this goose is a serious operation."),
		economy.GrowCrowd(1.08),
	))
	dramatic := func(s economy.State, r *rand.Rand) Outcome {
		return outcome(
			notify.Warning("📸 He Livestreamed the Ban",
				"The wook films himself getting walked out and captions it 'they weren't ready for the vision.' It goes mildly viral; a few edgelords boycott, but the grown-ups nod approvingly."),
			economy.GrowCrowd(1.02),
		)
	}
	return option(
		"Ban him from the pond",
		"No pivots, no roadmaps, no vibes. Just geese shipping Goose Premium. The disciplined flock earns quiet respect.",
		pick(br(0.8, focus), br(0.2, dramatic)),
	)
}
