package content

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

type Condition struct {
	Left  Expr   `json:"left"`
	Op    string `json:"op"`
	Right Expr   `json:"right"`
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

type RangeValue struct {
	Value float64 `json:"value,omitempty"`
	Min   float64 `json:"min,omitempty"`
	Max   float64 `json:"max,omitempty"`
}

type GeneratedVar struct {
	Key    string  `json:"key"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	MinInt int     `json:"min_int,omitempty"`
	MaxInt int     `json:"max_int,omitempty"`
	Kind   string  `json:"kind,omitempty"`
}

type TriggerTemplate struct {
	Type   string     `json:"type"`
	P      RangeValue `json:"p,omitempty"`
	Level  int        `json:"level,omitempty"`
	Cap    float64    `json:"cap,omitempty"`
	High   float64    `json:"high,omitempty"`
	Low    float64    `json:"low,omitempty"`
	Chance RangeValue `json:"chance,omitempty"`
}

type OutcomeTemplate struct {
	Weight   RangeValue `json:"weight"`
	Tone     string     `json:"tone"`
	Titles   []string   `json:"titles"`
	Messages []string   `json:"messages"`
	Values   []ValueDef `json:"values,omitempty"`
	Effects  []Effect   `json:"effects,omitempty"`
}

type OptionTemplate struct {
	Label    string            `json:"label"`
	Desc     string            `json:"desc"`
	Outcomes []OutcomeTemplate `json:"outcomes"`
}

type EventTemplate struct {
	Key        string            `json:"key"`
	Trigger    TriggerTemplate   `json:"trigger"`
	Conditions []Condition       `json:"conditions,omitempty"`
	Vars       []GeneratedVar    `json:"vars,omitempty"`
	Outcomes   []OutcomeTemplate `json:"outcomes"`
}

type CharacterTemplate struct {
	Key        string              `json:"key"`
	Headline   string              `json:"headline"`
	Chance     RangeValue          `json:"chance"`
	Conditions []Condition         `json:"conditions,omitempty"`
	Text       map[string][]string `json:"text,omitempty"`
	Pitches    []string            `json:"pitches"`
	Vars       []GeneratedVar      `json:"vars,omitempty"`
	Values     []ValueDef          `json:"values,omitempty"`
	Options    []OptionTemplate    `json:"options"`
}

type NarrativeData struct {
	Events     []EventTemplate     `json:"events"`
	Characters []CharacterTemplate `json:"characters"`
}

var Narrative NarrativeData
