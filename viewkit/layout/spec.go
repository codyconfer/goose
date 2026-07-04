package layout

import (
	"encoding/json"
	"fmt"
)

type Params map[string]any

func (p Params) Int(key string, def int) int {
	v, ok := p[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case json.Number:
		if i, err := n.Int64(); err == nil {
			return int(i)
		}
	}
	return def
}

type PaneRef struct {
	Key     string   `json:"key"`
	Pos     *GridPos `json:"pos,omitempty"`
	MinTier *Tier    `json:"minTier,omitempty"`
}

type ScreenSpec struct {
	Layout       string    `json:"layout"`
	LayoutParams Params    `json:"layoutParams,omitempty"`
	Panes        []PaneRef `json:"panes"`
}

func BuildScreen[Ctx any](s ScreenSpec, ctx Ctx, r *Registry[Ctx]) (Screen, error) {
	if r == nil {
		return Screen{}, fmt.Errorf("layout: nil registry")
	}

	lf, ok := r.layouts[s.Layout]
	if !ok {
		return Screen{}, fmt.Errorf("layout: unknown layout %q", s.Layout)
	}
	l, err := lf(s.LayoutParams)
	if err != nil {
		return Screen{}, fmt.Errorf("layout: build layout %q: %w", s.Layout, err)
	}
	if l == nil {
		return Screen{}, fmt.Errorf("layout: layout %q produced nil Layout", s.Layout)
	}

	panes := make([]Pane, 0, len(s.Panes))
	for _, ref := range s.Panes {
		pf, ok := r.panes[ref.Key]
		if !ok {
			return Screen{}, fmt.Errorf("layout: unknown pane %q", ref.Key)
		}
		p, ok := pf(ctx)
		if !ok {
			continue
		}
		if ref.Pos != nil {
			pos := *ref.Pos
			p.Pos = &pos
		}
		if ref.MinTier != nil {
			p.MinTier = *ref.MinTier
		}
		panes = append(panes, p)
	}

	return Screen{Layout: l, Panes: panes}, nil
}
