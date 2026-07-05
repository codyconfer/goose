package content

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed data/*.json
var dataFiles embed.FS

type Producer struct {
	Key         string  `json:"key"`
	Name        string  `json:"name"`
	Icon        string  `json:"icon"`
	BaseCost    float64 `json:"base_cost"`
	TokenRate   float64 `json:"token_rate"`
	EggRate     float64 `json:"egg_rate"`
	UnlockLevel int     `json:"unlock_level"`
	Desc        string  `json:"desc"`
	Singularity bool    `json:"singularity,omitempty"`
}

type Upgrade struct {
	Key      string  `json:"key"`
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Desc     string  `json:"desc"`
	BaseCost float64 `json:"base_cost"`
	Growth   float64 `json:"growth"`
}

type balanceData struct {
	CostGrowth float64 `json:"cost_growth"`

	BaselineTokens float64 `json:"baseline_tokens"`
	BaselineEggs   float64 `json:"baseline_eggs"`

	BasePrice        float64 `json:"base_price"`
	ConsumerAppetite float64 `json:"consumer_appetite"`
	CrowdHeadroom    float64 `json:"crowd_headroom"`
	CrowdAdjustRate  float64 `json:"crowd_adjust_rate"`

	PriceHoardOnOffer       float64 `json:"price_hoard_on_offer"`
	PriceMinSupply          float64 `json:"price_min_supply"`
	PriceAdjustRate         float64 `json:"price_adjust_rate"`
	PriceVolatility         float64 `json:"price_volatility"`
	PriceTrendDecay         float64 `json:"price_trend_decay"`
	PriceTrendVol           float64 `json:"price_trend_vol"`
	PriceTrendMax           float64 `json:"price_trend_max"`
	PriceShockTrend         float64 `json:"price_shock_trend"`
	PriceTrendTurbulence    float64 `json:"price_trend_turbulence"`
	PriceTrendReversionDrag float64 `json:"price_trend_reversion_drag"`
	PriceCashflowTrend      float64 `json:"price_cashflow_trend"`
	PriceTradeTrend         float64 `json:"price_trade_trend"`
	PriceDerivativeTrend    float64 `json:"price_derivative_trend"`
	PriceFloor              float64 `json:"price_floor"`
	PriceCeil               float64 `json:"price_ceil"`

	CrierBonusPerLevel float64 `json:"crier_bonus_per_level"`
	BlitzBonusPerLevel float64 `json:"blitz_bonus_per_level"`

	TradeFloorRate float64 `json:"trade_floor_rate"`
	TradeEpsilon   float64 `json:"trade_epsilon"`

	SpecUnlockLevel       int       `json:"spec_unlock_level"`
	SpecExpirySeconds     float64   `json:"spec_expiry_seconds"`
	SpecMarginPenalty     float64   `json:"spec_margin_penalty"`
	SpecMaintenanceMargin float64   `json:"spec_maintenance_margin"`
	SpecLeverages         []float64 `json:"spec_leverages"`
	SpecPremiums          []float64 `json:"spec_premiums"`

	DecommissionRefund float64 `json:"decommission_refund"`
	LedgerMax          int     `json:"ledger_max"`
	AgentQueueMax      int     `json:"agent_queue_max"`

	MaxOfflineSeconds int64 `json:"max_offline_seconds"`
}

type tuningData struct {
	UpBeatRateMs int `json:"up_beat_rate_ms"`

	FlashBeats            int `json:"flash_beats"`
	NotifBeats            int `json:"notif_beats"`
	OutcomeBeats          int `json:"outcome_beats"`
	CharacterTimeoutBeats int `json:"character_timeout_beats"`
	OfflineBeats          int `json:"offline_beats"`

	NotifQueueCap int `json:"notif_queue_cap"`

	PulseDecayRate   float64 `json:"pulse_decay_rate"`
	BuyRateSmoothing float64 `json:"buy_rate_smoothing"`
}

type SettingOption struct {
	Label string  `json:"label"`
	Mult  float64 `json:"mult"`
}

type SettingSpec struct {
	Label   string          `json:"label"`
	Desc    string          `json:"desc"`
	Default int             `json:"default"`
	Options []SettingOption `json:"options"`
}

type settingsData struct {
	LevelPace  SettingSpec `json:"level_pace"`
	EventPace  SettingSpec `json:"event_pace"`
	MarketPace SettingSpec `json:"market_pace"`
}

type textData struct {
	App struct {
		Title    string `json:"title"`
		Subtitle string `json:"subtitle"`
		Tagline  string `json:"tagline"`
		Quit     string `json:"quit"`
	} `json:"app"`
	Menu struct {
		NewGame   string `json:"new_game"`
		Continue  string `json:"continue"`
		Exit      string `json:"exit"`
		SaveError string `json:"save_error"`
	} `json:"menu"`
	Status struct {
		Panel       string `json:"panel"`
		RateFmt     string `json:"rate_fmt"`
		LevelFmt    string `json:"level_fmt"`
		ProgressFmt string `json:"progress_fmt"`
		MaxLevel    string `json:"max_level"`
	} `json:"status"`
	Tapper struct {
		OfflineFmt string `json:"offline_fmt"`
	} `json:"tapper"`
	Activity struct {
		Idle string `json:"idle"`
	} `json:"activity"`
	Viz struct {
		Panel string `json:"panel"`
		Empty string `json:"empty"`
	} `json:"viz"`
	Clock struct {
		Panel       string `json:"panel"`
		BinaryPanel string `json:"binary_panel"`
	} `json:"clock"`
	Feed struct {
		Panel string `json:"panel"`
	} `json:"feed"`
	Market struct {
		Panel          string `json:"panel"`
		PriceSteady    string `json:"price_steady"`
		PriceDemand    string `json:"price_demand"`
		PriceGlut      string `json:"price_glut"`
		StockLabel     string `json:"stock_label"`
		MarketCapLabel string `json:"market_cap_label"`
		LayingLabel    string `json:"laying_label"`
		SellingLabel   string `json:"selling_label"`
		ConsumersLabel string `json:"consumers_label"`
		PriceLabel     string `json:"price_label"`
	} `json:"market"`
	Capex struct {
		Panel               string `json:"panel"`
		ProducerBoughtFmt   string `json:"producer_bought_fmt"`
		ProducerDeniedFmt   string `json:"producer_denied_fmt"`
		ProducerSoldFmt     string `json:"producer_sold_fmt"`
		ProducerCantSellFmt string `json:"producer_cant_sell_fmt"`
		LockedTeaserFmt     string `json:"locked_teaser_fmt"`
		UpgradeBoughtFmt    string `json:"upgrade_bought_fmt"`
		UpgradeDeniedFmt    string `json:"upgrade_denied_fmt"`
		UpgradeCantSell     string `json:"upgrade_cant_sell"`
	} `json:"capex"`
	Trade struct {
		DeskTitle                string `json:"desk_title"`
		PursePanel               string `json:"purse_panel"`
		MarketPriceLabel         string `json:"market_price_label"`
		ConsumersPayLabel        string `json:"consumers_pay_label"`
		NewOrderPanel            string `json:"new_order_panel"`
		DirectionLabel           string `json:"direction_label"`
		AmountLabel              string `json:"amount_label"`
		EstimateLabel            string `json:"estimate_label"`
		BuyToggle                string `json:"buy_toggle"`
		SellToggle               string `json:"sell_toggle"`
		SpendFmt                 string `json:"spend_fmt"`
		ProceedsFmt              string `json:"proceeds_fmt"`
		MaxFmt                   string `json:"max_fmt"`
		ClearedFlash             string `json:"cleared_flash"`
		CancelledFlash           string `json:"cancelled_flash"`
		QueuedFmt                string `json:"queued_fmt"`
		NothingToSchedule        string `json:"nothing_to_schedule"`
		VerbBuy                  string `json:"verb_buy"`
		VerbSell                 string `json:"verb_sell"`
		CompletedSellFmt         string `json:"completed_sell_fmt"`
		CompletedBuyFmt          string `json:"completed_buy_fmt"`
		QueuePanel               string `json:"queue_panel"`
		QueueConsumersLabel      string `json:"queue_consumers_label"`
		QueueConsumersSuffix     string `json:"queue_consumers_suffix"`
		QueueQuiet               string `json:"queue_quiet"`
		PriceChartPanel          string `json:"price_chart_panel"`
		PriceChartTitleFmt       string `json:"price_chart_title_fmt"`
		PriceChartGathering      string `json:"price_chart_gathering"`
		TrendFlat                string `json:"trend_flat"`
		TrendUpFmt               string `json:"trend_up_fmt"`
		TrendDownFmt             string `json:"trend_down_fmt"`
		TrendLabel               string `json:"trend_label"`
		TrendStrengthLabel       string `json:"trend_strength_label"`
		TrendBullFmt             string `json:"trend_bull_fmt"`
		TrendBearFmt             string `json:"trend_bear_fmt"`
		TrendSideways            string `json:"trend_sideways"`
		NowPrefix                string `json:"now_prefix"`
		LedgerPanel              string `json:"ledger_panel"`
		LedgerEmpty              string `json:"ledger_empty"`
		LedgerBuyEggsFmt         string `json:"ledger_buy_eggs_fmt"`
		LedgerSellEggsFmt        string `json:"ledger_sell_eggs_fmt"`
		LedgerBuyProducerPrefix  string `json:"ledger_buy_producer_prefix"`
		LedgerSellProducerPrefix string `json:"ledger_sell_producer_prefix"`
		LedgerUpgradePrefix      string `json:"ledger_upgrade_prefix"`
		LedgerOptionOpenFmt      string `json:"ledger_option_open_fmt"`
		LedgerOptionSettleFmt    string `json:"ledger_option_settle_fmt"`
		SpecLockedFmt            string `json:"spec_locked_fmt"`
		FlowPanel                string `json:"flow_panel"`
		FlowLaying               string `json:"flow_laying"`
		FlowSelling              string `json:"flow_selling"`
		FlowDemand               string `json:"flow_demand"`
		SpotSection              string `json:"spot_section"`
		DerivSection             string `json:"deriv_section"`
		AgentsSection            string `json:"agents_section"`
		TapeSection              string `json:"tape_section"`
		LedgerSection            string `json:"ledger_section"`
	} `json:"trade"`
	Spec struct {
		DeskTitle       string `json:"desk_title"`
		PursePanel      string `json:"purse_panel"`
		PriceLabel      string `json:"price_label"`
		ExposureLabel   string `json:"exposure_label"`
		TrendLabel      string `json:"trend_label"`
		TicketPanel     string `json:"ticket_panel"`
		DirectionLabel  string `json:"direction_label"`
		CallToggle      string `json:"call_toggle"`
		PutToggle       string `json:"put_toggle"`
		CallThesis      string `json:"call_thesis"`
		PutThesis       string `json:"put_thesis"`
		PremiumLabel    string `json:"premium_label"`
		LeverageLabel   string `json:"leverage_label"`
		NotionalLabel   string `json:"notional_label"`
		LiqPriceLabel   string `json:"liq_price_label"`
		BufferLabel     string `json:"buffer_label"`
		ExpiryLabel     string `json:"expiry_label"`
		ExpiryFmt       string `json:"expiry_fmt"`
		RiskLabel       string `json:"risk_label"`
		WipeWarnFmt     string `json:"wipe_warn_fmt"`
		PositionsPanel  string `json:"positions_panel"`
		PositionsEmpty  string `json:"positions_empty"`
		PosDescFmt      string `json:"pos_desc_fmt"`
		ExpiresInFmt    string `json:"expires_in_fmt"`
		OpenedFmt       string `json:"opened_fmt"`
		CantAfford      string `json:"cant_afford"`
		ClosedFmt       string `json:"closed_fmt"`
		ClosedAllFlash  string `json:"closed_all_flash"`
		NothingToClose  string `json:"nothing_to_close"`
		SettledWinFmt   string `json:"settled_win_fmt"`
		SettledLossFmt  string `json:"settled_loss_fmt"`
		MarginCallTitle string `json:"margin_call_title"`
		MarginCallFmt   string `json:"margin_call_fmt"`
		CallWord        string `json:"call_word"`
		PutWord         string `json:"put_word"`
		PnlPanel        string `json:"pnl_panel"`
		MixPanel        string `json:"mix_panel"`
		MixEmpty        string `json:"mix_empty"`
		MixCash         string `json:"mix_cash"`
		MixEggs         string `json:"mix_eggs"`
		MixExposure     string `json:"mix_exposure"`
	} `json:"spec"`
	Character struct {
		Prompt   string `json:"prompt"`
		BackHint string `json:"back_hint"`
	} `json:"character"`
	Agents struct {
		DeskTitle string `json:"desk_title"`
		Subtitle  string `json:"subtitle"`
		Panel     string `json:"panel"`
		Empty     string `json:"empty"`
		FiredFmt  string `json:"fired_fmt"`
		RuleFmt   string `json:"rule_fmt"`
		OnWord    string `json:"on_word"`
		OffWord   string `json:"off_word"`

		MetricTrend  string `json:"metric_trend"`
		MetricPrice  string `json:"metric_price"`
		MetricTokens string `json:"metric_tokens"`
		MetricEggs   string `json:"metric_eggs"`
		CmpAbove     string `json:"cmp_above"`
		CmpBelow     string `json:"cmp_below"`
		ActBuyEggs   string `json:"act_buy_eggs"`
		ActSellEggs  string `json:"act_sell_eggs"`
		ActOpenCall  string `json:"act_open_call"`
		ActOpenPut   string `json:"act_open_put"`

		Roster map[string]struct {
			Name  string `json:"name"`
			Blurb string `json:"blurb"`
		} `json:"roster"`
	} `json:"agents"`
}

var (
	Producers  []Producer
	Upgrades   []Upgrade
	Levels     []float64
	TradeSizes []float64
	Balance    balanceData
	Tuning     tuningData
	Text       textData
	Settings   settingsData
)

func init() {
	mustLoad("data/producers.json", &Producers)
	mustLoad("data/upgrades.json", &Upgrades)
	mustLoad("data/levels.json", &Levels)
	mustLoad("data/trade_sizes.json", &TradeSizes)
	mustLoad("data/balance.json", &Balance)
	mustLoad("data/tuning.json", &Tuning)
	mustLoad("data/text.json", &Text)
	mustLoad("data/settings.json", &Settings)
	mustLoad("data/encounters.json", &Narrative)
}

func mustLoad(name string, dst any) {
	raw, err := dataFiles.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("content: reading %s: %v", name, err))
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		panic(fmt.Sprintf("content: parsing %s: %v", name, err))
	}
}
