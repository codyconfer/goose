package world

import (
	"bytes"
	"math"
	"math/rand"
	"text/template"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/notify"
	out "github.com/codyconfer/goose/internal/outcome"
)

type State struct {
	Seed       int64       `json:"seed"`
	Events     []Event     `json:"events"`
	Characters []Character `json:"characters"`
}

type Trigger struct {
	Type   string  `json:"type"`
	P      float64 `json:"p,omitempty"`
	Level  int     `json:"level,omitempty"`
	Cap    float64 `json:"cap,omitempty"`
	High   float64 `json:"high,omitempty"`
	Low    float64 `json:"low,omitempty"`
	Chance float64 `json:"chance,omitempty"`
}

func (t Trigger) Repeatable() bool {
	switch t.Type {
	case "chance", "egg_price", "margin":
		return true
	default:
		return false
	}
}

func (t Trigger) Fires(s economy.State, r *rand.Rand) bool {
	switch t.Type {
	case "chance":
		if r == nil {
			return false
		}
		return r.Float64() < t.P*s.Settings.EventMult()
	case "level":
		return s.Level() >= t.Level
	case "market_cap":
		return s.MarketCap() >= t.Cap
	case "egg_price":
		if r == nil {
			return false
		}
		p := s.EggPrice()
		outside := (t.High > 0 && p >= t.High) || (t.Low > 0 && p <= t.Low)
		return outside && r.Float64() < t.Chance
	case "margin":
		stress := s.MarginStress()
		if stress <= 0 {
			return false
		}
		if t.Chance >= 1 {
			return true
		}
		if r == nil {
			return false
		}
		chance := t.Chance * (0.45 + 0.55*stress) * (0.75 + 0.25*s.TrendStrength())
		if chance > 1 {
			chance = 1
		}
		return r.Float64() < chance
	default:
		return false
	}
}

type Condition struct {
	Left  Expr   `json:"left"`
	Op    string `json:"op"`
	Right Expr   `json:"right"`
}

type Expr struct {
	Op     string  `json:"op,omitempty"`
	Value  float64 `json:"value,omitempty"`
	Name   string  `json:"name,omitempty"`
	Key    string  `json:"key,omitempty"`
	Args   []Expr  `json:"args,omitempty"`
	Min    float64 `json:"min,omitempty"`
	Max    float64 `json:"max,omitempty"`
	MinInt int     `json:"min_int,omitempty"`
	MaxInt int     `json:"max_int,omitempty"`
}

type ValueDef struct {
	Key   string `json:"key"`
	Value Expr   `json:"value"`
}

type Effect struct {
	Type   string `json:"type"`
	Value  Expr   `json:"value,omitempty"`
	Factor Expr   `json:"factor,omitempty"`
	Count  int    `json:"count,omitempty"`
	Key    string `json:"key,omitempty"`
	Kind   string `json:"kind,omitempty"`
	Price  Expr   `json:"price,omitempty"`
}

type Outcome struct {
	Weight  float64    `json:"weight"`
	Tone    string     `json:"tone"`
	Title   string     `json:"title"`
	Message string     `json:"message"`
	Values  []ValueDef `json:"values,omitempty"`
	Effects []Effect   `json:"effects,omitempty"`
}

type Option struct {
	Label    string    `json:"label"`
	Desc     string    `json:"desc"`
	Outcomes []Outcome `json:"outcomes"`
}

type Event struct {
	Key        string             `json:"key"`
	Trigger    Trigger            `json:"trigger"`
	Conditions []Condition        `json:"conditions,omitempty"`
	Vars       map[string]float64 `json:"vars,omitempty"`
	Outcomes   []Outcome          `json:"outcomes"`
}

func (e Event) Eligible(s economy.State) bool {
	return matchAll(e.Conditions, stateEnv(s, e.Vars, nil, nil))
}

func (e Event) Apply(s economy.State, r *rand.Rand) out.Outcome {
	return resolveOutcomes(e.Outcomes, stateEnv(s, e.Vars, nil, r))
}

type Character struct {
	Key        string             `json:"key"`
	Headline   string             `json:"headline"`
	Name       string             `json:"name"`
	Pitch      string             `json:"pitch"`
	Chance     float64            `json:"chance"`
	Conditions []Condition        `json:"conditions,omitempty"`
	Vars       map[string]float64 `json:"vars,omitempty"`
	Text       map[string]string  `json:"text,omitempty"`
	Values     []ValueDef         `json:"values,omitempty"`
	Options    []Option           `json:"options"`
}

func (c Character) Eligible(s economy.State) bool {
	return matchAll(c.Conditions, stateEnv(s, c.Vars, c.Text, nil))
}

type ResolvedCharacter struct {
	Key      string
	Headline string
	Name     string
	Pitch    string
	Options  []ResolvedOption
}

type ResolvedOption struct {
	Label   string
	Desc    string
	Resolve func(s economy.State, r *rand.Rand) out.Outcome
}

func (c Character) Build(s economy.State) ResolvedCharacter {
	env := stateEnv(s, c.Vars, c.Text, nil)
	env.text["name"] = c.Name
	base := evalDefs(c.Values, env)
	env.numbers = mergeNumbers(env.numbers, base)
	pitch := render(c.Pitch, env.renderData())
	opts := make([]ResolvedOption, len(c.Options))
	for i, opt := range c.Options {
		label := render(opt.Label, env.renderData())
		desc := render(opt.Desc, env.renderData())
		capturedNumbers := copyNumbers(env.numbers)
		capturedText := copyText(env.text)
		capturedOutcomes := opt.Outcomes
		opts[i] = ResolvedOption{
			Label: label,
			Desc:  desc,
			Resolve: func(state economy.State, r *rand.Rand) out.Outcome {
				return resolveOutcomes(capturedOutcomes, stateEnv(state, capturedNumbers, capturedText, r))
			},
		}
	}
	return ResolvedCharacter{
		Key:      c.Key,
		Headline: c.Headline,
		Name:     c.Name,
		Pitch:    pitch,
		Options:  opts,
	}
}

type evalEnv struct {
	state   economy.State
	numbers map[string]float64
	text    map[string]string
	rng     *rand.Rand
}

func stateEnv(state economy.State, numbers map[string]float64, text map[string]string, rng *rand.Rand) evalEnv {
	return evalEnv{
		state:   state,
		numbers: copyNumbers(numbers),
		text:    copyText(text),
		rng:     rng,
	}
}

func copyNumbers(in map[string]float64) map[string]float64 {
	if len(in) == 0 {
		return map[string]float64{}
	}
	out := make(map[string]float64, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyText(in map[string]string) map[string]string {
	if len(in) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func mergeNumbers(a, b map[string]float64) map[string]float64 {
	out := copyNumbers(a)
	for k, v := range b {
		out[k] = v
	}
	return out
}

func evalDefs(defs []ValueDef, env evalEnv) map[string]float64 {
	out := map[string]float64{}
	merged := copyNumbers(env.numbers)
	for _, def := range defs {
		next := evalExpr(def.Value, evalEnv{
			state:   env.state,
			numbers: merged,
			text:    env.text,
			rng:     env.rng,
		})
		out[def.Key] = next
		merged[def.Key] = next
	}
	return out
}

func matchAll(conds []Condition, env evalEnv) bool {
	for _, cond := range conds {
		left := evalExpr(cond.Left, env)
		right := evalExpr(cond.Right, env)
		switch cond.Op {
		case "gt":
			if !(left > right) {
				return false
			}
		case "gte":
			if !(left >= right) {
				return false
			}
		case "lt":
			if !(left < right) {
				return false
			}
		case "lte":
			if !(left <= right) {
				return false
			}
		case "eq":
			if left != right {
				return false
			}
		default:
			return false
		}
	}
	return true
}

func evalExpr(expr Expr, env evalEnv) float64 {
	switch expr.Op {
	case "", "const":
		return expr.Value
	case "var":
		return env.numbers[expr.Name]
	case "state":
		return stateValue(env.state, expr.Name)
	case "owned":
		return float64(env.state.Count(expr.Key))
	case "add":
		sum := 0.0
		for _, arg := range expr.Args {
			sum += evalExpr(arg, env)
		}
		return sum
	case "mul":
		prod := 1.0
		for _, arg := range expr.Args {
			prod *= evalExpr(arg, env)
		}
		return prod
	case "sub":
		if len(expr.Args) == 0 {
			return 0
		}
		diff := evalExpr(expr.Args[0], env)
		for _, arg := range expr.Args[1:] {
			diff -= evalExpr(arg, env)
		}
		return diff
	case "div":
		if len(expr.Args) < 2 {
			return 0
		}
		num := evalExpr(expr.Args[0], env)
		den := evalExpr(expr.Args[1], env)
		if den == 0 {
			return 0
		}
		return num / den
	case "min":
		if len(expr.Args) == 0 {
			return 0
		}
		best := evalExpr(expr.Args[0], env)
		for _, arg := range expr.Args[1:] {
			best = math.Min(best, evalExpr(arg, env))
		}
		return best
	case "max":
		if len(expr.Args) == 0 {
			return 0
		}
		best := evalExpr(expr.Args[0], env)
		for _, arg := range expr.Args[1:] {
			best = math.Max(best, evalExpr(arg, env))
		}
		return best
	case "affordable":
		if len(expr.Args) == 0 {
			return 0
		}
		// Outcome costs are allowed to exceed the balance and push tokens
		// negative; we only guard against a negative charge.
		cost := evalExpr(expr.Args[0], env)
		if cost < 0 {
			return 0
		}
		return cost
	case "rand":
		if env.rng == nil {
			return expr.Min + (expr.Max-expr.Min)/2
		}
		return expr.Min + env.rng.Float64()*(expr.Max-expr.Min)
	case "rand_int":
		lo, hi := expr.MinInt, expr.MaxInt
		if hi < lo {
			lo, hi = hi, lo
		}
		if env.rng == nil || hi <= lo {
			return float64(lo)
		}
		return float64(lo + env.rng.Intn(hi-lo+1))
	default:
		return 0
	}
}

func stateValue(s economy.State, name string) float64 {
	switch name {
	case "tokens":
		return s.Tokens
	case "total_earned":
		return s.TotalEarned
	case "per_click":
		return s.PerClick
	case "eggs":
		return s.Eggs
	case "consumers":
		return s.Consumers
	case "price_factor":
		return s.PriceFactor
	case "market_cap":
		return s.MarketCap()
	case "level":
		return float64(s.Level())
	case "egg_price":
		return s.EggPrice()
	case "sell_price":
		return s.SellPrice()
	case "tokens_per_second":
		return s.TokensPerSecond()
	case "eggs_per_second":
		return s.EggsPerSecond()
	case "margin_stress":
		return s.MarginStress()
	case "trend_strength":
		return s.TrendStrength()
	default:
		return 0
	}
}

func resolveOutcomes(specs []Outcome, env evalEnv) out.Outcome {
	if len(specs) == 0 {
		return out.Outcome{}
	}
	spec := pickOutcome(specs, env.rng)
	values := evalDefs(spec.Values, env)
	mergedNumbers := mergeNumbers(env.numbers, values)
	mergedEnv := evalEnv{
		state:   env.state,
		numbers: mergedNumbers,
		text:    copyText(env.text),
		rng:     env.rng,
	}
	data := mergedEnv.renderData()
	cmds := make([]economy.Command, 0, len(spec.Effects))
	for _, eff := range spec.Effects {
		cmds = append(cmds, commandFor(eff, mergedEnv))
	}
	return out.New(notify.Note(
		tone(spec.Tone),
		render(spec.Title, data),
		render(spec.Message, data),
	), cmds...)
}

func pickOutcome(specs []Outcome, r *rand.Rand) Outcome {
	total := 0.0
	for _, spec := range specs {
		if spec.Weight > 0 {
			total += spec.Weight
		}
	}
	if total <= 0 || r == nil {
		return specs[len(specs)-1]
	}
	x := r.Float64() * total
	for _, spec := range specs {
		if spec.Weight <= 0 {
			continue
		}
		if x < spec.Weight {
			return spec
		}
		x -= spec.Weight
	}
	return specs[len(specs)-1]
}

func commandFor(spec Effect, env evalEnv) economy.Command {
	switch spec.Type {
	case "earn":
		return economy.Earn(evalExpr(spec.Value, env))
	case "spend":
		return economy.Spend(evalExpr(spec.Value, env))
	case "grant_producer":
		return economy.GrantProducer(render(spec.Key, env.renderData()), spec.Count)
	case "grow_crowd":
		return economy.GrowCrowd(evalExpr(spec.Factor, env))
	case "add_consumers":
		return economy.AddConsumers(evalExpr(spec.Value, env))
	case "trade":
		return economy.Trade(txKind(spec.Kind), evalExpr(spec.Value, env), evalExpr(spec.Price, env))
	case "seize_best":
		return economy.SeizeBest()
	case "freeze":
		return economy.Freeze(evalExpr(spec.Value, env), render(spec.Key, env.renderData()))
	case "margin_call":
		return economy.MarginCall(evalExpr(spec.Factor, env))
	case "shock_price":
		return economy.ShockPrice(evalExpr(spec.Factor, env))
	default:
		return economy.Command{}
	}
}

func txKind(kind string) economy.TxKind {
	switch kind {
	case "buy_eggs":
		return economy.TxBuyEggs
	case "sell_eggs":
		return economy.TxSellEggs
	default:
		return economy.TxBuyEggs
	}
}

func tone(raw string) notify.Tone {
	switch raw {
	case "positive":
		return notify.TonePositive
	case "warning":
		return notify.ToneWarning
	case "negative":
		return notify.ToneNegative
	default:
		return notify.ToneNeutral
	}
}

func render(tmpl string, data map[string]any) string {
	if tmpl == "" {
		return ""
	}
	t, err := template.New("world").Funcs(template.FuncMap{
		"num": economy.FormatNum,
	}).Parse(tmpl)
	if err != nil {
		return tmpl
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return tmpl
	}
	return buf.String()
}

func (env evalEnv) renderData() map[string]any {
	data := make(map[string]any, len(env.text)+len(env.numbers))
	for k, v := range env.text {
		data[k] = v
	}
	for k, v := range env.numbers {
		data[k] = v
	}
	return data
}
