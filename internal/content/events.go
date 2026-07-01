package content

type eventText struct {
	Title   string `json:"title"`
	Message string `json:"message"`
}

type eventFmtText struct {
	Title      string `json:"title"`
	MessageFmt string `json:"message_fmt"`
}

type eventsData struct {
	LuckyEgg               eventFmtText `json:"lucky_egg"`
	GoldenHour             eventFmtText `json:"golden_hour"`
	MarketBoom             eventFmtText `json:"market_boom"`
	MarketDay              eventFmtText `json:"market_day"`
	FoxRaid                eventFmtText `json:"fox_raid"`
	MarginCall             eventFmtText `json:"margin_call"`
	FlashCrash             eventText    `json:"flash_crash"`
	MeltUp                 eventText    `json:"melt_up"`
	PressDarling           eventFmtText `json:"press_darling"`
	IPORumor               eventFmtText `json:"ipo_rumor"`
	SellingFrenzy          eventFmtText `json:"selling_frenzy"`
	ROIReckoning           eventFmtText `json:"roi_reckoning"`
	TokenBurn              eventFmtText `json:"token_burn"`
	GPUShortage            eventFmtText `json:"gpu_shortage"`
	ChipDelay              eventText    `json:"chip_delay"`
	StargateGroundbreaking eventFmtText `json:"stargate_groundbreaking"`
	SovereignMandate       eventFmtText `json:"sovereign_mandate"`
	LudicrousValuation     eventFmtText `json:"ludicrous_valuation"`
	OpenWeightsDump        eventText    `json:"open_weights_dump"`
	DataBreach             eventFmtText `json:"data_breach"`
	BagholderCapitulation  eventText    `json:"bagholder_capitulation"`
	HypeCyclePeak          eventFmtText `json:"hype_cycle_peak"`
	WanderingGoose         eventText    `json:"wandering_goose"`

	CircularInvestment struct {
		Gain eventFmtText `json:"gain"`
		Loss eventFmtText `json:"loss"`
	} `json:"circular_investment"`

	VaporwareKeynote struct {
		Gain eventFmtText `json:"gain"`
		Fail eventText    `json:"fail"`
	} `json:"vaporware_keynote"`

	EfficiencyMemo struct {
		Gain     eventFmtText `json:"gain"`
		Backlash eventText    `json:"backlash"`
	} `json:"efficiency_memo"`

	SuperintelligenceBlog struct {
		Gain eventFmtText `json:"gain"`
		Miss eventText    `json:"miss"`
	} `json:"superintelligence_blog"`

	GridStrain struct {
		Surcharge eventFmtText `json:"surcharge"`
		Brownout  eventText    `json:"brownout"`
	} `json:"grid_strain"`
}
