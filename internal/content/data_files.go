package content

var dataFiles = map[string]string{
	"data/balance.json": `
{
  "cost_growth": 1.15,

  "base_price": 2.5,
  "consumer_appetite": 0.5,
  "crowd_headroom": 0.6,
  "crowd_adjust_rate": 0.4,

  "price_hoard_on_offer": 0.05,
  "price_min_supply": 0.1,
  "price_adjust_rate": 0.28,
  "price_volatility": 0.2,
  "price_trend_decay": 0.05,
  "price_trend_vol": 0.04,
  "price_trend_max": 0.12,
  "price_shock_trend": 0.08,
  "price_floor": 0.4,
  "price_ceil": 2.0,

  "crier_bonus_per_level": 0.3,
  "blitz_bonus_per_level": 0.25,

  "trade_floor_rate": 3.0,
  "trade_epsilon": 1e-6,

  "spec_unlock_level": 4,
  "spec_expiry_seconds": 45,
  "spec_margin_penalty": 0.25,
  "spec_leverages": [1, 2, 5, 10],
  "spec_premiums": [10, 50, 200, 1000, 10000],

  "decommission_refund": 0.5,
  "ledger_max": 20,

  "max_offline_seconds": 600
}
`,
	"data/levels.json": `
[0, 50, 750, 10000, 150000, 2500000, 40000000, 600000000, 9000000000, 140000000000, 2000000000000]
`,
	"data/producers.json": `
[
  { "key": "gpu",         "name": "GPU",                "icon": "🎮",  "base_cost": 15,             "token_rate": 0.2,      "egg_rate": 0.05,    "unlock_level": 1,  "desc": "One consumer card in a shoebox. A trickle of tokens and the occasional egg. Everyone starts here; most people should have stopped here." },
  { "key": "server",      "name": "Server",             "icon": "🖥️",  "base_cost": 120,            "token_rate": 1.5,      "egg_rate": 0.4,     "unlock_level": 2,  "desc": "An honest goose in a box, earning real tokens and shipping sellable Goose Premium. Suspiciously profitable — enjoy it while it lasts." },
  { "key": "rack",        "name": "Server Rack",        "icon": "🗄️",  "base_cost": 1300,           "token_rate": 10,       "egg_rate": 2.5,     "unlock_level": 3,  "desc": "A whole rack of geese. Mints the good stuff, ships Goose Premium, and hums loud enough to drown out the CFO." },
  { "key": "datacenter",  "name": "Data Center",        "icon": "🏢",  "base_cost": 14000,          "token_rate": 65,       "egg_rate": 16,      "unlock_level": 4,  "desc": "Industrial-scale honking. Announced with a press release, a governor, and a number nobody will ever reconcile against actual demand." },
  { "key": "hyperscaler", "name": "Hyper Scaler",       "icon": "☁️", "base_cost": 200000,         "token_rate": 420,      "egg_rate": 100,     "unlock_level": 5,  "desc": "Geese with a pension plan, a wholesale egg contract, and a trillion in compute commitments due 'later.' Too big to fail, allegedly." },
  { "key": "starcloud",   "name": "Star Cloud",         "icon": "🛰️",  "base_cost": 3300000,        "token_rate": 2900,     "egg_rate": 700,     "unlock_level": 6,  "desc": "Geese. In orbit. For inference. The pitch deck said the vacuum is free cooling and no one on the cap table was sober enough to disagree." },
  { "key": "stargate",    "name": "Stargate",           "icon": "🌌",  "base_cost": 45000000,       "token_rate": 20000,    "egg_rate": 4800,    "unlock_level": 7,  "desc": "A half-trillion-token megacluster announced onstage beside a head of state. Financing is a rounding error to be sorted out later; the honking is audible from space." },
  { "key": "sovereign",   "name": "Sovereign Cloud",    "icon": "🏛️",  "base_cost": 650000000,      "token_rate": 140000,   "egg_rate": 33000,   "unlock_level": 8,  "desc": "An entire nation-state decides it needs its own goose 'for national security.' The eggs are classified and the invoices are eternal." },
  { "key": "fusion",      "name": "Fusion Reactor",     "icon": "⚛️", "base_cost": 9500000000,     "token_rate": 950000,   "egg_rate": 230000,  "unlock_level": 9,  "desc": "The only way to power this many geese. Commercially viable any day now — has been for forty years. Delivers Goose Premium at ludicrous scale in the meantime." },
  { "key": "dyson",       "name": "Dyson Swarm",        "icon": "☀️", "base_cost": 140000000000,   "token_rate": 6500000,  "egg_rate": 1600000, "unlock_level": 10, "desc": "We ran out of planet, so now we wrap the sun in geese to power the training run. The roadmap swore this was a Q3 deliverable." },
  { "key": "factory",     "name": "Golden Egg Factory", "icon": "🏭",  "base_cost": 2200000000000,  "token_rate": 45000000, "egg_rate": 11000000, "unlock_level": 11, "desc": "You cut the goose open and it's clockwork all the way down — a factory that manufactures golden eggs that ship Goose Premium. Masa was right. Own the goose. Valuation: one quadrillion, do not check the math." }
]
`,
	"data/settings.json": `
{
  "level_pace": {
    "label": "Level Pace",
    "desc": "How long the climb to each new level takes.",
    "default": 1,
    "options": [
      { "label": "FAST", "mult": 0.5 },
      { "label": "STANDARD", "mult": 1.0 },
      { "label": "LONG", "mult": 2.5 },
      { "label": "EXTREMELY LONG", "mult": 6.0 },
      { "label": "ETERNAL", "mult": 20.0 }
    ]
  },
  "event_pace": {
    "label": "Event Pace",
    "desc": "Frequency of events and character encounters.",
    "default": 2,
    "options": [
      { "label": "Rarely", "mult": 0.35 },
      { "label": "Sparsely", "mult": 0.65 },
      { "label": "Common", "mult": 1.0 },
      { "label": "Often", "mult": 1.75 },
      { "label": "Deluge", "mult": 3.0 }
    ]
  },
  "market_pace": {
    "label": "Market Pace",
    "desc": "How fast your assets grow!",
    "default": 2,
    "options": [
      { "label": "DotCom bubble", "mult": 0.5 },
      { "label": "Web 2.0", "mult": 0.75 },
      { "label": "BLOCKCHAIN", "mult": 1.0 },
      { "label": "AI HYPERSCALERS", "mult": 2.0 },
      { "label": "DATACENTERS IN SPACE", "mult": 5.0 },
      { "label": "SINGULARITY GOONING", "mult": 25.0 }
    ]
  }
}
`,
	"data/events.json": `
{
  "lucky_egg": {
    "title": "🍀 Lucky Egg",
    "message_fmt": "You find a rare golden egg and flip it to a collector who 'really gets the vision' for %s bonus tokens."
  },
  "golden_hour": {
    "title": "🌟 Golden Hour",
    "message_fmt": "The whole flock catches the same standup energy at once — +%s tokens before the vibe wears off."
  },
  "market_boom": {
    "title": "📈 Market Boom",
    "message_fmt": "Demand surges on no news whatsoever — buyers hurl %s tokens at you purely to avoid missing out."
  },
  "wandering_goose": {
    "title": "🪂 Soft Landing",
    "message": "A goose washes up from a startup that just burned its last round and joins your flock for free — résumé still warm, eyes slightly haunted."
  },
  "market_day": {
    "title": "🛒 Market Day",
    "message_fmt": "A rival data center liquidates its inventory — eggs flood the stalls dirt-cheap and you snap up %s of them for %s tokens!"
  },
  "fox_raid": {
    "title": "🧠 Brain Drain",
    "message_fmt": "A rival lab waves nine-figure signing bonuses at your sharpest geese. You counter with 'retention' and %s tokens fly out the door just to keep the flock from walking."
  },
  "margin_call": {
    "title": "🏦 Margin Call",
    "message_fmt": "A twitchy prop desk marks your book to market and pulls the plug. Leverage giveth; the margin clerk taketh %s%% of it back."
  },
  "flash_crash": {
    "title": "📉 Flash Crash",
    "message": "A single skeptical thread goes viral and the whole egg complex gaps down before anyone can find the sell button. Calls get vaporized; the puts look prophetic."
  },
  "melt_up": {
    "title": "🚀 Melt-Up",
    "message": "An analyst 'updates his model' and eggs rip vertical on zero new information. Momentum funds pile in; the shorts are quietly getting a phone call."
  },
  "press_darling": {
    "title": "📰 Local Darlings",
    "message_fmt": "A breathless blog post calls your flock 'the goose that will eat search' — fresh buyers crowd in and you pocket %s tokens."
  },
  "ipo_rumor": {
    "title": "🏦 IPO Rumors",
    "message_fmt": "Bankers catch wind of your egg empire and start whispering 'S-1' at each other — speculative buying nets you %s tokens."
  },
  "selling_frenzy": {
    "title": "🤑 Selling Frenzy",
    "message_fmt": "Eggs are trading at a premium nobody can justify — you unload %s of them for %s tokens before the music stops."
  },
  "circular_investment": {
    "gain": {
      "title": "🔄 Circular Investment",
      "message_fmt": "Two data-center barons announce they're investing in each other and, somehow, in you. The market cap balloons and %s tokens spill over. Nobody reads the footnotes."
    },
    "loss": {
      "title": "🔗 The Circle Broke",
      "message_fmt": "One baron misses a payment and the whole daisy-chain unwinds at once. The 'revenue' was everyone's money going in a circle; %s tokens of it keeps going, right out the door."
    }
  },
  "roi_reckoning": {
    "title": "🧮 ROI Reckoning",
    "message_fmt": "A CFO somewhere finally asks, 'wait — what's the return on all these eggs?' Budgets freeze, %s tokens get clawed back, and the momentum crowd sobers up fast."
  },
  "token_burn": {
    "title": "🔥 Token Burn",
    "message_fmt": "An intern left the goose running overnight on \"max reasoning.\" You wake up to a %s-token bill and a slide deck that should have been an email."
  },
  "gpu_shortage": {
    "title": "🪫 GPU Shortage",
    "message_fmt": "Every card on Earth is allocated to someone with a bigger check. You pay %s tokens in scalper markup just to keep the geese warm."
  },
  "vaporware_keynote": {
    "gain": {
      "title": "🎤 Vaporware Keynote",
      "message_fmt": "You demo an egg that doesn't exist yet to rapturous applause. Pre-orders and hype net %s tokens before anyone asks for a ship date."
    },
    "fail": {
      "title": "🎬 The Demo Failed On Stage",
      "message": "The pre-rendered egg freezes, then crashes, then displays a stack trace to a live audience. The clip loops all night with a laugh track. The believers wince; the tourists leave."
    }
  },
  "chip_delay": {
    "title": "🍽️ Dinner-Plate Chip Slips to 2027",
    "message": "Your custom silicon — a motherboard the size of a dinner plate — quietly slips another year. The believers hold; the tourists drift off."
  },
  "stargate_groundbreaking": {
    "title": "🌌 Stargate Groundbreaking",
    "message_fmt": "You break ground on a half-trillion-token megacluster next to a head of state and a gold shovel. The market swoons and %s tokens rain down. Financing? Later."
  },
  "sovereign_mandate": {
    "title": "🗽 Sovereign AI Mandate",
    "message_fmt": "A nation-state declares your flock critical infrastructure and signs a blank check worth %s tokens. The eggs are now classified."
  },
  "ludicrous_valuation": {
    "title": "🏭 Quadrillion-Token Valuation",
    "message_fmt": "You unveil the Golden Egg Factory inside the goose. Analysts stop analyzing and simply believe. %s tokens materialize. Own the goose, not the eggs."
  },
  "open_weights_dump": {
    "title": "📂 Rival Open-Sources Their Eggs",
    "message": "A lab you'd never heard of dumps a comparable egg on the internet for free \"to accelerate humanity.\" Your moat evaporates overnight and the whole complex re-prices lower."
  },
  "efficiency_memo": {
    "gain": {
      "title": "✂️ Unlocking Efficiencies",
      "message_fmt": "You 'right-size the flock to focus on the mission' and the market rewards the discipline: +%s tokens on the news. The remaining geese work weekends now."
    },
    "backlash": {
      "title": "😠 The Layoff Backlash",
      "message": "The \"efficiency\" memo leaks with the CEO's yacht in the background. The cut geese talk to reporters, morale craters, and the crowd sours on the whole vibe."
    }
  },
  "superintelligence_blog": {
    "gain": {
      "title": "🧠 'The Gentle Singularity'",
      "message_fmt": "You publish a 3,000-word blog promising the eggs will soon cure disease, fix the climate, and possibly love you back. Belief spikes and %s tokens of true believers arrive."
    },
    "miss": {
      "title": "🥱 Nobody Read the Blog",
      "message": "You publish the manifesto about superintelligence. It gets twelve likes and a reply asking when the eggs will actually ship. The moment passes."
    }
  },
  "data_breach": {
    "title": "🔓 Egg Data Breach",
    "message_fmt": "Every customer's egg-buying history leaks, tastefully, onto a hacking forum. %s tokens in 'incident response' and free credit monitoring, and a chunk of the crowd never comes back."
  },
  "grid_strain": {
    "surcharge": {
      "title": "⚡ Grid Strain Surcharge",
      "message_fmt": "Your data centers now draw more power than a mid-size town, and the utility has noticed. A %s-token peak-demand surcharge lands, plus a stern letter about 'the community.'"
    },
    "brownout": {
      "title": "🔌 Rolling Brownout",
      "message": "The regional grid buckles under your compute and the utility throttles you off-peak. The geese sit in the dark for a bit while the town keeps its lights on."
    }
  },
  "bagholder_capitulation": {
    "title": "🫠 Bagholder Capitulation",
    "message": "Prices have been ugly long enough that the last diamond-handed believers finally give up and post their loss porn. The capitulation clears the froth but thins the crowd."
  },
  "hype_cycle_peak": {
    "title": "🎢 Peak of Inflated Expectations",
    "message_fmt": "Every magazine cover is a goose. Your barber has egg exposure. It's obviously the top and everyone's buying anyway — you ride the mania for %s tokens."
  }
}
`,
	"data/text.json": `
{
  "app": {
    "title": "🪿 GOLDEN GOOSE",
    "subtitle": "own the goose, not the eggs™",
    "tagline": "an idle egg-economy bubble simulator",
    "quit": "🔌 The flock powers down. Your net asset value is safe (unlike your market cap). See you soon!"
  },
  "menu": {
    "new_game": "Found the Company",
    "continue": "Load a Save",
    "exit": "Exit (into liquidity)",
    "save_error": "Couldn't read that save."
  },
  "status": {
    "panel": "FLOCK",
    "rate_fmt": "+%s/sec  ·  +%s/tap",
    "level_fmt": "⭐ Level %d",
    "progress_fmt": "%s / %s 🥚 → Lv.%d",
    "max_level": "max level — the flock is legendary"
  },
  "tapper": {
    "offline_fmt": "💤 While you touched grass, the flock quietly out-earned your day job by %s tokens."
  },
  "activity": {
    "idle": "… the flock waddles about, disrupting nothing in particular. Anything could happen."
  },
  "market": {
    "panel": "EGG MARKET",
    "price_steady": "→ steady (boring, sustainable)",
    "price_demand": "↑ in demand (do NOT ask about ROI)",
    "price_glut": "↓ glut (open-source eggs again)",
    "stock_label": "🥚 stock",
    "market_cap_label": "📈 market cap",
    "laying_label": "📦 Goose Premium",
    "selling_label": "🛒 selling",
    "consumers_label": "👥 consumers",
    "price_label": "🏷️ price"
  },
  "capex": {
    "panel": "CAPEX  ·  spend to remain competitive",
    "producer_bought_fmt": "%s Bought a %s.",
    "producer_denied_fmt": "Not enough tokens for a %s.",
    "producer_sold_fmt": "♻️ Decommissioned a %s for %s 🪙.",
    "producer_cant_sell_fmt": "No %s to decommission.",
    "locked_teaser_fmt": "🔒 %s — reach Level %d",
    "upgrade_bought_fmt": "%s %s upgraded.",
    "upgrade_denied_fmt": "Not enough tokens for %s yet.",
    "upgrade_cant_sell": "Upgrades can't be decommissioned."
  },
  "trade": {
    "desk_title": "📊 TRADE DESK",
    "purse_panel": "PURSE",
    "market_price_label": "market price",
    "consumers_pay_label": "consumers pay",
    "new_order_panel": "NEW ORDER",
    "direction_label": "direction",
    "amount_label": "amount",
    "estimate_label": "estimate",
    "buy_toggle": "◂ Buy eggs",
    "sell_toggle": "Sell eggs ▸",
    "spend_fmt": "≈ %s 🪙 to spend",
    "proceeds_fmt": "≈ %s 🪙 proceeds",
    "max_fmt": "Max (%s)",
    "cleared_flash": "Cleared the trade queue.",
    "cancelled_flash": "Cancelled the active order.",
    "queued_fmt": "Queued: %s %s 🥚.",
    "nothing_to_schedule": "Nothing to schedule — pick a non-zero amount.",
    "verb_buy": "Buy",
    "verb_sell": "Sell",
    "completed_sell_fmt": "✅ Sold %s 🥚 to the crowd.",
    "completed_buy_fmt": "✅ Bought %s 🥚 off the market.",
    "queue_panel": "TRADE QUEUE  ·  worked every beat",
    "queue_consumers_label": "👥 Consumers",
    "queue_consumers_suffix": " · buying eggs",
    "queue_quiet": "quiet — no consumers and no orders yet",
    "price_chart_panel": "EGG PRICE",
    "price_chart_title_fmt": "EGG PRICE  ·  last %d beats",
    "price_chart_gathering": "gathering market data…",
    "trend_flat": "→ flat",
    "trend_up_fmt": "▲ %.0f%%",
    "trend_down_fmt": "▼ %.0f%%",
    "now_prefix": "now ",
    "ledger_panel": "LEDGER  ·  recent transactions",
    "ledger_empty": "no transactions yet",
    "ledger_buy_eggs_fmt": "🥚 Bought %s eggs",
    "ledger_sell_eggs_fmt": "🥚 Sold %s eggs",
    "ledger_buy_producer_prefix": "🏗️ Bought ",
    "ledger_sell_producer_prefix": "♻️ Decommissioned ",
    "ledger_upgrade_prefix": "✨ Upgraded ",
    "ledger_option_open_fmt": "📄 Opened %s",
    "ledger_option_settle_fmt": "📄 Settled %s",
    "spec_locked_fmt": "🔒 Derivatives Desk — reach Level %d to discover financial engineering",
    "flow_panel": "MARKET FLOW  ·  eggs per second",
    "flow_laying": "Goose Premium",
    "flow_selling": "sold",
    "flow_demand": "wanted"
  },
  "spec": {
    "desk_title": "📉 DERIVATIVES DESK  ·  bet on the bet",
    "purse_panel": "PURSE",
    "price_label": "spot egg price",
    "exposure_label": "leveraged exposure",
    "ticket_panel": "WRITE A CONTRACT",
    "direction_label": "thesis",
    "call_toggle": "◂ Call (long the bubble)",
    "put_toggle": "Put (short the bubble) ▸",
    "call_thesis": "eggs go up and to the right, forever",
    "put_thesis": "someone will eventually ask about revenue",
    "premium_label": "premium",
    "leverage_label": "leverage",
    "expiry_label": "settles in",
    "expiry_fmt": "%.0fs",
    "risk_label": "downside",
    "wipe_warn_fmt": "margin call if spot moves %s against you",
    "positions_panel": "OPEN POSITIONS  ·  marked every beat",
    "positions_empty": "flat — no exposure, no glory",
    "pos_desc_fmt": "%s %s @ %s",
    "expires_in_fmt": "%s left",
    "opened_fmt": "📄 Opened %s for %s 🪙.",
    "cant_afford": "Not enough tokens to post that premium.",
    "closed_fmt": "📕 Closed %s for %s 🪙.",
    "closed_all_flash": "📚 Closed out the whole book.",
    "nothing_to_close": "No open positions to close.",
    "settled_win_fmt": "📈 %s expired in the money — +%s 🪙.",
    "settled_loss_fmt": "📉 %s expired worthless — kissed %s 🪙 goodbye.",
    "margin_call_title": "⚠️ Margin Call",
    "margin_call_fmt": "A twitchy prop desk marks your book to market and pulls the plug. Leverage giveth; the margin clerk taketh %s of it back.",
    "call_word": "Call",
    "put_word": "Put",
    "pnl_panel": "OPEN P&L  ·  marked to spot",
    "mix_panel": "BOOK COMPOSITION  ·  where the money sits",
    "mix_empty": "no capital deployed — you're all cash and vibes",
    "mix_cash": "cash",
    "mix_eggs": "egg inventory",
    "mix_exposure": "leveraged exposure"
  },
  "character": {
    "prompt": "How do you respond?",
    "back_hint": "back to the grind"
  }
}
`,
	"data/trade_sizes.json": `
[10, 50, 100, 500, 1000, 5000, 25000]
`,
	"data/tuning.json": `
{
  "up_beat_rate_ms": 300,

  "flash_beats": 4,
  "notif_beats": 15,
  "outcome_beats": 20,
  "character_timeout_beats": 23,

  "notif_queue_cap": 8,

  "pulse_decay_rate": 3.0,
  "buy_rate_smoothing": 0.8
}
`,
	"data/upgrades.json": `
[
  { "key": "click", "name": "Enter the Flow State", "icon": "🧘", "desc": "Double the tokens you earn per tap. You, locked in and 'building' at 2am, deep in the flow — allegedly.",                  "base_cost": 50,   "growth": 8 },
  { "key": "crier", "name": "Jet Set Huang",        "icon": "🧥", "desc": "He zips up the magical leather jacket, takes the stage, and consumers pay ~30% more per egg. It's not a bubble if the keynote is good enough.", "base_cost": 150,  "growth": 7 },
  { "key": "blitz", "name": "Blitzscaling",  "icon": "🌀",  "desc": "Every producer outputs +25% per level. Grow at all costs, worry about unit economics never.",               "base_cost": 4000, "growth": 6 }
]
`,
}
