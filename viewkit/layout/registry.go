package layout

import "sort"

type PaneFactory[Ctx any] func(ctx Ctx) (pane Pane, ok bool)

type LayoutFactory func(params Params) (Layout, error)

type PaneInfo struct {
	Key   string
	Title string
}

type Registry[Ctx any] struct {
	panes    map[string]PaneFactory[Ctx]
	paneInfo map[string]PaneInfo
	order    []string
	layouts  map[string]LayoutFactory
}

func NewRegistry[Ctx any]() *Registry[Ctx] {
	r := &Registry[Ctx]{
		panes:    map[string]PaneFactory[Ctx]{},
		paneInfo: map[string]PaneInfo{},
		layouts:  map[string]LayoutFactory{},
	}
	r.LayoutFn("single", func(Params) (Layout, error) { return SingleColumn{}, nil })
	r.LayoutFn("flex-columns", func(p Params) (Layout, error) {
		return FlexColumns{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("flex-rows", func(p Params) (Layout, error) {
		return FlexRows{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("sections", func(p Params) (Layout, error) {
		return FlexSections{
			MinWidth: p.Int("minWidth", DefaultFlexMinWidth),
			MaxCols:  p.Int("maxCols", DefaultFlexMaxCols),
		}, nil
	})
	r.LayoutFn("grid", func(p Params) (Layout, error) {
		return Grid{Cols: p.Int("cols", 1), Rows: p.Int("rows", 0)}, nil
	})
	return r
}

func (r *Registry[Ctx]) Pane(key, title string, f PaneFactory[Ctx]) *Registry[Ctx] {
	if _, exists := r.panes[key]; !exists {
		r.order = append(r.order, key)
	}
	r.panes[key] = f
	r.paneInfo[key] = PaneInfo{Key: key, Title: title}
	return r
}

func (r *Registry[Ctx]) LayoutFn(key string, f LayoutFactory) *Registry[Ctx] {
	r.layouts[key] = f
	return r
}

func (r *Registry[Ctx]) PaneKeys() []PaneInfo {
	out := make([]PaneInfo, 0, len(r.order))
	for _, k := range r.order {
		out = append(out, r.paneInfo[k])
	}
	return out
}

func (r *Registry[Ctx]) LayoutKeys() []string {
	out := make([]string, 0, len(r.layouts))
	for k := range r.layouts {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
