package events

import (
	"fmt"
	"math/rand"

	"github.com/codyconfer/goose/internal/content"
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
			ev := content.Events.LuckyEgg
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
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
			ev := content.Events.GoldenHour
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
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
			ev := content.Events.MarketBoom
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
				economy.Earn(gain),
			)
		},
	},
	{
		Key:     "wandering_goose",
		Trigger: ChanceTrigger{P: chanceP(0.006)},
		CanFire: func(s economy.State) bool { return s.TotalEarned > 120 },
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			ev := content.Events.WanderingGoose
			return outcome(
				notify.Positive(ev.Title, ev.Message),
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
			ev := content.Events.MarketDay
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(bought), economy.FormatNum(spend))),
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
			ev := content.Events.FoxRaid
			return outcome(
				notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(loss))),
				economy.Spend(loss),
			)
		},
	},
	{
		Key:     "margin_call",
		Trigger: MarginTrigger{Chance: 0.35},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			ev := content.Events.MarginCall
			return outcome(
				notify.Negative(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(economy.MarginPenaltyPct()))),
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
			ev := content.Events.FlashCrash
			return outcome(
				notify.Warning(ev.Title, ev.Message),
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
			ev := content.Events.MeltUp
			return outcome(
				notify.Positive(ev.Title, ev.Message),
				economy.ShockPrice(factor),
			)
		},
	},
	{
		Key:     "press_darling",
		Trigger: LevelTrigger{Level: 3},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*120 + 100
			ev := content.Events.PressDarling
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
				economy.GrowCrowd(1.5), economy.Earn(gain),
			)
		},
	},
	{
		Key:     "ipo_rumor",
		Trigger: MarketCapTrigger{Cap: 100_000},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*300 + 500
			ev := content.Events.IPORumor
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
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
			ev := content.Events.SellingFrenzy
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(sold), economy.FormatNum(earned))),
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
				ev := content.Events.CircularInvestment.Gain
				return outcome(
					notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.2), economy.AddConsumers(6),
				)
			}),
			br(0.3, func(s economy.State, r *rand.Rand) Outcome {
				loss := s.Tokens * (0.08 + r.Float64()*0.1)
				ev := content.Events.CircularInvestment.Loss
				return outcome(
					notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(loss))),
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
			ev := content.Events.ROIReckoning
			return outcome(
				notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(loss))),
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
			ev := content.Events.TokenBurn
			return outcome(
				notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(loss))),
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
			ev := content.Events.GPUShortage
			return outcome(
				notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(loss))),
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
				ev := content.Events.VaporwareKeynote.Gain
				return outcome(
					notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.25),
				)
			}),
			br(0.28, func(s economy.State, r *rand.Rand) Outcome {
				ev := content.Events.VaporwareKeynote.Fail
				return outcome(
					notify.Warning(ev.Title, ev.Message),
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
			ev := content.Events.ChipDelay
			return outcome(
				notify.Warning(ev.Title, ev.Message),
				economy.GrowCrowd(0.85),
			)
		},
	},
	{
		Key:     "stargate_groundbreaking",
		Trigger: LevelTrigger{Level: 7},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*200 + 5000
			ev := content.Events.StargateGroundbreaking
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
				economy.GrowCrowd(1.6), economy.Earn(gain),
			)
		},
	},
	{
		Key:     "sovereign_mandate",
		Trigger: LevelTrigger{Level: 9},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*300 + 100000
			ev := content.Events.SovereignMandate
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
				economy.Earn(gain), economy.AddConsumers(40),
			)
		},
	},
	{
		Key:     "ludicrous_valuation",
		Trigger: LevelTrigger{Level: 11},
		Apply: func(s economy.State, r *rand.Rand) Outcome {
			gain := s.TokensPerSecond()*400 + 5_000_000
			ev := content.Events.LudicrousValuation
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
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
			ev := content.Events.OpenWeightsDump
			return outcome(
				notify.Warning(ev.Title, ev.Message),
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
				ev := content.Events.EfficiencyMemo.Gain
				return outcome(
					notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(0.95),
				)
			}),
			br(0.4, func(s economy.State, r *rand.Rand) Outcome {
				ev := content.Events.EfficiencyMemo.Backlash
				return outcome(
					notify.Warning(ev.Title, ev.Message),
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
				ev := content.Events.SuperintelligenceBlog.Gain
				return outcome(
					notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
					economy.Earn(gain), economy.GrowCrowd(1.3),
				)
			}),
			br(0.35, func(s economy.State, r *rand.Rand) Outcome {
				ev := content.Events.SuperintelligenceBlog.Miss
				return outcome(
					notify.Neutral(ev.Title, ev.Message),
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
			ev := content.Events.DataBreach
			return outcome(
				notify.Negative(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(fine))),
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
				ev := content.Events.GridStrain.Surcharge
				return outcome(
					notify.Warning(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(bill))),
					economy.Spend(bill),
				)
			}),
			br(0.4, func(s economy.State, r *rand.Rand) Outcome {
				dur := 6 + r.Float64()*8
				ev := content.Events.GridStrain.Brownout
				return outcome(
					notify.Warning(ev.Title, ev.Message),
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
			ev := content.Events.BagholderCapitulation
			return outcome(
				notify.Warning(ev.Title, ev.Message),
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
			ev := content.Events.HypeCyclePeak
			return outcome(
				notify.Positive(ev.Title, fmt.Sprintf(ev.MessageFmt, economy.FormatNum(gain))),
				economy.Earn(gain), economy.GrowCrowd(1.15),
			)
		},
	},
}
