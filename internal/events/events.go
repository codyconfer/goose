package events

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
)

type Event struct {
	Key     string
	Trigger Trigger
	CanFire func(s economy.State) bool
	Apply   func(s economy.State, r *rand.Rand) Outcome
}

const chanceBaseScale = 0.35

func chanceP(p float64) float64 { return p * chanceBaseScale }

var Events = []Event{
	{
		Key:     "lucky_egg",
		Trigger: ChanceTrigger{P: chanceP(0.024)},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.PerClick * float64(3+r.Intn(5))
			if gain < 5 {
				gain = 5
			}
			return outcome(
				notify.Positive("🍀 Lucky Egg",
					fmt.Sprintf("You find a rare golden egg and flip it to a collector who 'really gets the vision' for %s bonus tokens.", economy.FormatNum(gain))),
				economy.Earn(gain),
			)
		},
	},
	{
		Key:     "golden_hour",
		Trigger: ChanceTrigger{P: chanceP(0.009)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 50 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			base := s.TokensPerSecond()*45 + s.PerClick*20
			if base < 100 {
				base = 100
			}
			gain := base * (1 + r.Float64())
			return outcome(
				notify.Positive("🌟 Golden Hour",
					fmt.Sprintf("The whole flock catches the same standup energy at once — +%s tokens before the vibe wears off.", economy.FormatNum(gain))),
				economy.Earn(gain),
			)
		},
	},
	{
		Key:     "market_boom",
		Trigger: ChanceTrigger{P: chanceP(0.009)},
		CanFire: func(s economy.State) bool { return s.TokensPerSecond() > 0 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond() * float64(60+r.Intn(60))
			return outcome(
				notify.Positive("📈 Market Boom",
					fmt.Sprintf("Demand surges on no news whatsoever — buyers hurl %s tokens at you purely to avoid missing out.", economy.FormatNum(gain))),
				economy.Earn(gain),
			)
		},
	},
	{
		Key:     "wandering_goose",
		Trigger: ChanceTrigger{P: chanceP(0.006)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 120 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			return outcome(
				notify.Positive("🦢 Wandering Goose",
					"A goose churns out of a collapsed AI startup and joins your flock for free, résumé still warm and eyes slightly haunted."),
				economy.GrantProducer("server", 1),
			)
		},
	},
	{
		Key:     "market_day",
		Trigger: ChanceTrigger{P: chanceP(0.0108)},
		CanFire: func(s economy.State) bool { return s.Tokens > 20 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			price := economy.BasePrice * (0.3 + r.Float64()*0.3)
			spend := s.Tokens * (0.2 + r.Float64()*0.3)
			bought := spend / price
			return outcome(
				notify.Positive("🛒 Market Day",
					fmt.Sprintf("A rival data center liquidates its inventory — eggs flood the stalls dirt-cheap and you snap up %s of them for %s tokens!", economy.FormatNum(bought), economy.FormatNum(spend))),
				economy.Trade(economy.TxBuyEggs, bought, price),
			)
		},
	},
	{
		Key:     "fox_raid",
		Trigger: ChanceTrigger{P: chanceP(0.009)},
		CanFire: func(s economy.State) bool { return s.Tokens > 50 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			loss := s.Tokens * (0.08 + r.Float64()*0.12)
			return outcome(
				notify.Warning("🦊 Fox Raid",
					fmt.Sprintf("A sly fox — or possibly an acquihire, hard to tell — makes off with %s tokens from the stash.", economy.FormatNum(loss))),
				economy.Spend(loss),
			)
		},
	},
	{
		Key:     "margin_call",
		Trigger: MarginTrigger{Chance: 0.35},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			return outcome(
				notify.Negative("🏦 Margin Call",
					fmt.Sprintf("A twitchy prop desk marks your book to market and pulls the plug. Leverage giveth; the margin clerk taketh %s%% of it back.", economy.FormatNum(economy.MarginPenaltyPct()))),
				economy.MarginCall(economy.SpecMarginPenalty),
			)
		},
	},
	{
		Key:     "flash_crash",
		Trigger: ChanceTrigger{P: chanceP(0.01)},
		CanFire: func(s economy.State) bool { return s.PriceFactor > 0.9 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			factor := 0.55 + r.Float64()*0.2
			return outcome(
				notify.Warning("📉 Flash Crash",
					"A single skeptical thread goes viral and the whole egg complex gaps down before anyone can find the sell button. Calls get vaporized; the puts look prophetic."),
				economy.ShockPrice(factor),
			)
		},
	},
	{
		Key:     "melt_up",
		Trigger: ChanceTrigger{P: chanceP(0.01)},
		CanFire: func(s economy.State) bool { return s.PriceFactor < 1.6 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			factor := 1.25 + r.Float64()*0.25
			return outcome(
				notify.Positive("🚀 Melt-Up",
					"An analyst 'updates his model' and eggs rip vertical on zero new information. Momentum funds pile in; the shorts are quietly getting a phone call."),
				economy.ShockPrice(factor),
			)
		},
	},
	{
		Key:     "press_darling",
		Trigger: LevelTrigger{Level: 3},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*120 + 100
			return outcome(
				notify.Positive("📰 Local Darlings",
					fmt.Sprintf("A breathless blog post calls your flock 'the goose that will eat search' — fresh buyers crowd in and you pocket %s tokens.", economy.FormatNum(gain))),
				economy.GrowCrowd(1.5), economy.Earn(gain),
			)
		},
	},
	{
		Key:     "ipo_rumor",
		Trigger: MarketCapTrigger{Cap: 100_000},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*300 + 500
			return outcome(
				notify.Positive("🏦 IPO Rumors",
					fmt.Sprintf("Bankers catch wind of your egg empire and start whispering 'S-1' at each other — speculative buying nets you %s tokens.", economy.FormatNum(gain))),
				economy.Earn(gain),
			)
		},
	},
	{
		Key:     "selling_frenzy",
		Trigger: EggPriceTrigger{High: economy.BasePrice * 1.8, Chance: 0.05},
		CanFire: func(s economy.State) bool { return s.Eggs > 10 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			sold := s.Eggs * (0.1 + r.Float64()*0.2)
			price := s.SellPrice()
			earned := sold * price
			return outcome(
				notify.Positive("🤑 Selling Frenzy",
					fmt.Sprintf("Eggs are trading at a premium nobody can justify — you unload %s of them for %s tokens before the music stops.", economy.FormatNum(sold), economy.FormatNum(earned))),
				economy.Trade(economy.TxSellEggs, sold, price),
			)
		},
	},
	{
		Key:     "circular_investment",
		Trigger: ChanceTrigger{P: chanceP(0.008)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 5000 },
		Apply: pick(
			br(0.7, func(s economy.State, r *rand.Rand) Outcome {
				gain := s.TokensPerSecond()*150 + 400
				return outcome(
					notify.Positive("🔄 Circular Investment",
						fmt.Sprintf("Two data-center barons announce they're investing in each other and, somehow, in you. The market cap balloons and %s tokens spill over. Nobody reads the footnotes.", economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.2), economy.AddConsumers(6),
				)
			}),
			br(0.3, func(s economy.State, r *rand.Rand) Outcome {
				loss := s.Tokens * (0.08 + r.Float64()*0.1)
				return outcome(
					notify.Warning("🔗 The Circle Broke",
						fmt.Sprintf("One baron misses a payment and the whole daisy-chain unwinds at once. The 'revenue' was everyone's money going in a circle; %s tokens of it keeps going, right out the door.", economy.FormatNum(loss))),
					economy.Spend(loss), economy.GrowCrowd(0.85), economy.ShockPrice(0.9),
				)
			}),
		),
	},
	{
		Key:     "roi_reckoning",
		Trigger: ChanceTrigger{P: chanceP(0.007)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 3000 && s.Tokens > 100 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			loss := s.Tokens * (0.1 + r.Float64()*0.15)
			return outcome(
				notify.Warning("🧮 ROI Reckoning",
					fmt.Sprintf("A CFO somewhere finally asks, 'wait — what's the return on all these eggs?' Budgets freeze, %s tokens get clawed back, and the momentum crowd sobers up fast.", economy.FormatNum(loss))),
				economy.Spend(loss), economy.GrowCrowd(0.8),
			)
		},
	},
	{
		Key:     "token_burn",
		Trigger: ChanceTrigger{P: chanceP(0.01)},
		CanFire: func(s economy.State) bool { return s.Tokens > 200 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			loss := s.Tokens * (0.05 + r.Float64()*0.1)
			return outcome(
				notify.Warning("🔥 Token Burn",
					fmt.Sprintf("An intern left the goose running overnight on 'max reasoning.' You wake up to a %s-token bill and a slide deck that should have been an email.", economy.FormatNum(loss))),
				economy.Spend(loss),
			)
		},
	},
	{
		Key:     "gpu_shortage",
		Trigger: ChanceTrigger{P: chanceP(0.009)},
		CanFire: func(s economy.State) bool { return s.Tokens > 1000 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			loss := s.Tokens * (0.06 + r.Float64()*0.1)
			return outcome(
				notify.Warning("🪫 GPU Shortage",
					fmt.Sprintf("Every card on Earth is allocated to someone with a bigger check. You pay %s tokens in scalper markup just to keep the geese warm.", economy.FormatNum(loss))),
				economy.Spend(loss),
			)
		},
	},
	{
		Key:     "vaporware_keynote",
		Trigger: ChanceTrigger{P: chanceP(0.008)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 2000 },
		Apply: pick(
			br(0.72, func(s economy.State, r *rand.Rand) Outcome {
				gain := s.TokensPerSecond()*90 + 250
				return outcome(
					notify.Positive("🎤 Vaporware Keynote",
						fmt.Sprintf("You demo an egg that doesn't exist yet to rapturous applause. Pre-orders and hype net %s tokens before anyone asks for a ship date.", economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.25),
				)
			}),
			br(0.28, func(s economy.State, r *rand.Rand) Outcome {
				return outcome(
					notify.Warning("🎬 The Demo Failed On Stage",
						"The pre-rendered egg freezes, then crashes, then displays a stack trace to a live audience. The clip loops all night with a laugh track. The believers wince; the tourists leave."),
					economy.GrowCrowd(0.82),
				)
			}),
		),
	},
	{
		Key:     "chip_delay",
		Trigger: ChanceTrigger{P: chanceP(0.007)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 5000 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			return outcome(
				notify.Warning("🍽️ Dinner-Plate Chip Slips to 2027",
					"Your custom silicon — a motherboard the size of a dinner plate — quietly slips another year. The believers hold; the tourists drift off."),
				economy.GrowCrowd(0.85),
			)
		},
	},
	{
		Key:     "stargate_groundbreaking",
		Trigger: LevelTrigger{Level: 7},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*200 + 5000
			return outcome(
				notify.Positive("🌌 Stargate Groundbreaking",
					fmt.Sprintf("You break ground on a half-trillion-token megacluster next to a head of state and a gold shovel. The market swoons and %s tokens rain down. Financing? Later.", economy.FormatNum(gain))),
				economy.GrowCrowd(1.6), economy.Earn(gain),
			)
		},
	},
	{
		Key:     "sovereign_mandate",
		Trigger: LevelTrigger{Level: 9},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*300 + 100000
			return outcome(
				notify.Positive("🗽 Sovereign AI Mandate",
					fmt.Sprintf("A nation-state declares your flock critical infrastructure and signs a blank check worth %s tokens. The eggs are now classified.", economy.FormatNum(gain))),
				economy.Earn(gain), economy.AddConsumers(40),
			)
		},
	},
	{
		Key:     "ludicrous_valuation",
		Trigger: LevelTrigger{Level: 11},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*400 + 5_000_000
			return outcome(
				notify.Positive("🏭 Quadrillion-Token Valuation",
					fmt.Sprintf("You unveil the Golden Egg Factory inside the goose. Analysts stop analyzing and simply believe. %s tokens materialize. Own the goose, not the eggs.", economy.FormatNum(gain))),
				economy.Earn(gain), economy.GrowCrowd(2.0),
			)
		},
	},
	{
		Key:     "open_weights_dump",
		Trigger: ChanceTrigger{P: chanceP(0.008)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 4000 && s.PriceFactor > 0.8 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			factor := 0.72 + r.Float64()*0.15
			return outcome(
				notify.Warning("📂 Rival Open-Sources Their Eggs",
					"A lab you'd never heard of dumps a comparable egg on the internet for free 'to accelerate humanity.' Your moat evaporates overnight and the whole complex re-prices lower."),
				economy.ShockPrice(factor), economy.GrowCrowd(0.9),
			)
		},
	},
	{
		Key:     "efficiency_memo",
		Trigger: ChanceTrigger{P: chanceP(0.009)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 3000 && s.Tokens > 100 },
		Apply: pick(
			br(0.6, func(s economy.State, r *rand.Rand) Outcome {
				gain := s.TokensPerSecond()*50 + 200
				return outcome(
					notify.Positive("✂️ Unlocking Efficiencies",
						fmt.Sprintf("You 'right-size the flock to focus on the mission' and the market rewards the discipline: +%s tokens on the news. The remaining geese work weekends now.", economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(0.95),
				)
			}),
			br(0.4, func(s economy.State, r *rand.Rand) Outcome {
				return outcome(
					notify.Warning("😠 The Layoff Backlash",
						"The 'efficiency' memo leaks with the CEO's yacht in the background. The cut geese talk to reporters, morale craters, and the crowd sours on the whole vibe."),
					economy.GrowCrowd(0.8),
				)
			}),
		),
	},
	{
		Key:     "superintelligence_blog",
		Trigger: ChanceTrigger{P: chanceP(0.008)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 6000 },
		Apply: pick(
			br(0.65, func(s economy.State, r *rand.Rand) Outcome {
				gain := s.TokensPerSecond()*80 + 400
				return outcome(
					notify.Positive("🧠 'The Gentle Singularity'",
						fmt.Sprintf("You publish a 3,000-word blog promising the eggs will soon cure disease, fix the climate, and possibly love you back. Belief spikes and %s tokens of true believers arrive.", economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.3),
				)
			}),
			br(0.35, func(s economy.State, r *rand.Rand) Outcome {
				return outcome(
					notify.Neutral("🥱 Nobody Read the Blog",
						"You publish the manifesto about superintelligence. It gets twelve likes and a reply asking when the eggs will actually ship. The moment passes."),
					economy.GrowCrowd(0.99),
				)
			}),
		),
	},
	{
		Key:     "data_breach",
		Trigger: ChanceTrigger{P: chanceP(0.007)},
		CanFire: func(s economy.State) bool { return s.Tokens > 500 && s.Consumers > 5 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			fine := s.Tokens * (0.07 + r.Float64()*0.08)
			return outcome(
				notify.Negative("🔓 Egg Data Breach",
					fmt.Sprintf("Every customer's egg-buying history leaks, tastefully, onto a hacking forum. %s tokens in 'incident response' and free credit monitoring, and a chunk of the crowd never comes back.", economy.FormatNum(fine))),
				economy.Spend(fine), economy.GrowCrowd(0.85),
			)
		},
	},
	{
		Key:     "grid_strain",
		Trigger: ChanceTrigger{P: chanceP(0.008)},
		CanFire: func(s economy.State) bool { return s.Tokens > 2000 && s.Count("datacenter") > 0 },
		Apply: pick(
			br(0.6, func(s economy.State, r *rand.Rand) Outcome {
				bill := s.Tokens * (0.05 + r.Float64()*0.08)
				return outcome(
					notify.Warning("⚡ Grid Strain Surcharge",
						fmt.Sprintf("Your data centers now draw more power than a mid-size town, and the utility has noticed. A %s-token peak-demand surcharge lands, plus a stern letter about 'the community.'", economy.FormatNum(bill))),
					economy.Spend(bill),
				)
			}),
			br(0.4, func(s economy.State, r *rand.Rand) Outcome {
				dur := 6 + r.Float64()*8
				return outcome(
					notify.Warning("🔌 Rolling Brownout",
						"The regional grid buckles under your compute and the utility throttles you off-peak. The geese sit in the dark for a bit while the town keeps its lights on."),
					economy.Freeze(dur, "the power grid tapped out under your data centers"),
				)
			}),
		),
	},
	{
		Key:     "bagholder_capitulation",
		Trigger: EggPriceTrigger{Low: economy.BasePrice * 0.55, Chance: 0.06},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 3000 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			return outcome(
				notify.Warning("🫠 Bagholder Capitulation",
					"Prices have been ugly long enough that the last diamond-handed believers finally give up and post their loss porn. The capitulation clears the froth but thins the crowd."),
				economy.GrowCrowd(0.8),
			)
		},
	},
	{
		Key:     "hype_cycle_peak",
		Trigger: EggPriceTrigger{High: economy.BasePrice * 1.9, Chance: 0.05},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 4000 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*60 + 300
			return outcome(
				notify.Positive("🎢 Peak of Inflated Expectations",
					fmt.Sprintf("Every magazine cover is a goose. Your barber has egg exposure. It's obviously the top and everyone's buying anyway — you ride the mania for %s tokens.", economy.FormatNum(gain))),
				economy.Earn(gain), economy.GrowCrowd(1.15),
			)
		},
	},
}
