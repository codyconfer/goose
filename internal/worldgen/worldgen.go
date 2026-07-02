package worldgen

import (
	"math/rand"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/world"
)

const DefaultSeed int64 = 1

func Generate(seed int64) *world.State {
	r := rand.New(rand.NewSource(seed))
	out := &world.State{
		Seed:       seed,
		Events:     make([]world.Event, 0, len(content.Narrative.Events)),
		Characters: make([]world.Character, 0, len(content.Narrative.Characters)),
	}
	for _, tmpl := range content.Narrative.Events {
		out.Events = append(out.Events, generateEvent(tmpl, r))
	}
	for _, tmpl := range content.Narrative.Characters {
		out.Characters = append(out.Characters, generateCharacter(tmpl, r))
	}
	return out
}

func generateEvent(tmpl content.EventTemplate, r *rand.Rand) world.Event {
	return world.Event{
		Key:        tmpl.Key,
		Trigger:    generateTrigger(tmpl.Trigger, r),
		Conditions: convertConditions(tmpl.Conditions),
		Vars:       generateVars(tmpl.Vars, r),
		Outcomes:   generateOutcomes(tmpl.Outcomes, r),
	}
}

func generateCharacter(tmpl content.CharacterTemplate, r *rand.Rand) world.Character {
	text := generateText(tmpl.Text, r)
	name := text["name"]
	if name == "" {
		name = tmpl.Key
	}
	return world.Character{
		Key:        tmpl.Key,
		Headline:   tmpl.Headline,
		Name:       name,
		Pitch:      chooseText(tmpl.Pitches, r),
		Chance:     resolveRange(tmpl.Chance, r),
		Conditions: convertConditions(tmpl.Conditions),
		Vars:       generateVars(tmpl.Vars, r),
		Text:       text,
		Values:     convertValueDefs(tmpl.Values),
		Options:    generateOptions(tmpl.Options, r),
	}
}

func generateTrigger(tmpl content.TriggerTemplate, r *rand.Rand) world.Trigger {
	return world.Trigger{
		Type:   tmpl.Type,
		P:      resolveRange(tmpl.P, r),
		Level:  tmpl.Level,
		Cap:    tmpl.Cap,
		High:   tmpl.High,
		Low:    tmpl.Low,
		Chance: resolveRange(tmpl.Chance, r),
	}
}

func generateVars(vars []content.GeneratedVar, r *rand.Rand) map[string]float64 {
	if len(vars) == 0 {
		return nil
	}
	out := make(map[string]float64, len(vars))
	for _, v := range vars {
		switch v.Kind {
		case "int":
			lo, hi := v.MinInt, v.MaxInt
			if hi < lo {
				lo, hi = hi, lo
			}
			if hi <= lo {
				out[v.Key] = float64(lo)
				continue
			}
			out[v.Key] = float64(lo + r.Intn(hi-lo+1))
		default:
			if v.Max == v.Min {
				out[v.Key] = v.Min
				continue
			}
			out[v.Key] = v.Min + r.Float64()*(v.Max-v.Min)
		}
	}
	return out
}

func generateText(src map[string][]string, r *rand.Rand) map[string]string {
	if len(src) == 0 {
		return nil
	}
	out := make(map[string]string, len(src))
	for key, options := range src {
		out[key] = chooseText(options, r)
	}
	return out
}

func generateOptions(opts []content.OptionTemplate, r *rand.Rand) []world.Option {
	out := make([]world.Option, len(opts))
	for i, opt := range opts {
		out[i] = world.Option{
			Label:    opt.Label,
			Desc:     opt.Desc,
			Outcomes: generateOutcomes(opt.Outcomes, r),
		}
	}
	return out
}

func generateOutcomes(templates []content.OutcomeTemplate, r *rand.Rand) []world.Outcome {
	out := make([]world.Outcome, len(templates))
	for i, tmpl := range templates {
		out[i] = world.Outcome{
			Weight:  resolveRange(tmpl.Weight, r),
			Tone:    tmpl.Tone,
			Title:   chooseText(tmpl.Titles, r),
			Message: chooseText(tmpl.Messages, r),
			Values:  convertValueDefs(tmpl.Values),
			Effects: convertEffects(tmpl.Effects),
		}
	}
	return out
}

func chooseText(options []string, r *rand.Rand) string {
	if len(options) == 0 {
		return ""
	}
	return options[r.Intn(len(options))]
}

func resolveRange(spec content.RangeValue, r *rand.Rand) float64 {
	switch {
	case spec.Max != 0 || spec.Min != 0:
		if spec.Max == spec.Min {
			return spec.Min
		}
		return spec.Min + r.Float64()*(spec.Max-spec.Min)
	default:
		return spec.Value
	}
}

func convertConditions(src []content.Condition) []world.Condition {
	if len(src) == 0 {
		return nil
	}
	out := make([]world.Condition, len(src))
	for i, cond := range src {
		out[i] = world.Condition{
			Left:  convertExpr(cond.Left),
			Op:    cond.Op,
			Right: convertExpr(cond.Right),
		}
	}
	return out
}

func convertValueDefs(src []content.ValueDef) []world.ValueDef {
	if len(src) == 0 {
		return nil
	}
	out := make([]world.ValueDef, len(src))
	for i, def := range src {
		out[i] = world.ValueDef{Key: def.Key, Value: convertExpr(def.Value)}
	}
	return out
}

func convertEffects(src []content.Effect) []world.Effect {
	if len(src) == 0 {
		return nil
	}
	out := make([]world.Effect, len(src))
	for i, eff := range src {
		out[i] = world.Effect{
			Type:   eff.Type,
			Value:  convertExpr(eff.Value),
			Factor: convertExpr(eff.Factor),
			Count:  eff.Count,
			Key:    eff.Key,
			Kind:   eff.Kind,
			Price:  convertExpr(eff.Price),
		}
	}
	return out
}

func convertExpr(expr content.Expr) world.Expr {
	out := world.Expr{
		Op:     expr.Op,
		Value:  expr.Value,
		Name:   expr.Name,
		Key:    expr.Key,
		Min:    expr.Min,
		Max:    expr.Max,
		MinInt: expr.MinInt,
		MaxInt: expr.MaxInt,
	}
	if len(expr.Args) > 0 {
		out.Args = make([]world.Expr, len(expr.Args))
		for i, arg := range expr.Args {
			out.Args[i] = convertExpr(arg)
		}
	}
	return out
}
