package characters

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/economy"
)

type Type int

const (
	VC Type = iota
	Wook
	Booster
	Politician
	Engineer
	MiddleClass
	Analyst
	ShortSeller
)

func (t Type) Headline() string {
	switch t {
	case VC:
		return "💼 A VENTURE CAPITALIST APPEARS"
	case Wook:
		return "🍄 A VISIONARY APPEARS"
	case Booster:
		return "🧥 THE LEATHER JACKET ARRIVES"
	case Politician:
		return "🏛️ AN ELECTED OFFICIAL APPEARS"
	case Engineer:
		return "🔧 AN ENGINEER APPEARS"
	case MiddleClass:
		return "🧺 THE MIDDLE CLASS ARRIVES"
	case Analyst:
		return "📊 A SELL-SIDE ANALYST APPEARS"
	case ShortSeller:
		return "🐻 A SHORT-SELLER PUBLISHES"
	default:
		return "✨ A STRANGER APPEARS"
	}
}

type Option struct {
	Label   string
	Desc    string
	Resolve func(s economy.State, r *rand.Rand) Outcome
}

type Character struct {
	Type    Type
	Name    string
	Pitch   string
	Stakes  float64
	Options []Option
}
type spawner struct {
	Eligible func(s economy.State) bool
	Chance   float64
	Spawn    func(s economy.State, r *rand.Rand) Character
}

var spawners = []spawner{
	{Eligible: VCEligible, Chance: vcChance, Spawn: NewVC},
	{Eligible: WookEligible, Chance: wookChance, Spawn: NewWook},
	{Eligible: BoosterEligible, Chance: boosterChance, Spawn: NewBooster},
	{Eligible: PoliticianEligible, Chance: politicianChance, Spawn: NewPolitician},
	{Eligible: EngineerEligible, Chance: engineerChance, Spawn: NewEngineer},
	{Eligible: MiddleClassEligible, Chance: middleClassChance, Spawn: NewMiddleClass},
	{Eligible: AnalystEligible, Chance: analystChance, Spawn: NewAnalyst},
	{Eligible: ShortSellerEligible, Chance: shortSellerChance, Spawn: NewShortSeller},
}

func Roll(s economy.State, r *rand.Rand) (Character, bool) {
	mult := s.Settings.EventMult()
	for _, i := range r.Perm(len(spawners)) {
		sp := spawners[i]
		if sp.Eligible != nil && !sp.Eligible(s) {
			continue
		}
		if r.Float64() < sp.Chance*mult {
			return sp.Spawn(s, r), true
		}
	}
	return Character{}, false
}
