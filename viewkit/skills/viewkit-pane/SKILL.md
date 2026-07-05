---
name: viewkit-pane
description: >-
  Add or modify a pane (a content block in a viewkit screen). Use when working
  with layout.Pane, Registry.Pane, PaneFactory, ScreenSpec/PaneRef, a screen's
  defaultSpec, the cellFrame inset, or the Pane fields Interactive / MinTier /
  Slim / Pos. Covers the three-place wiring a pane needs and the (Pane, bool)
  visibility-predicate trap.
---

# Add / modify a viewkit pane

A **pane** is a named content block a screen arranges. Panes are registered by
string key in a generic `layout.Registry[Ctx]`, and screens reference them by key
via a `layout.ScreenSpec`. This is what makes screen composition data-driven
(and JSON-serializable / user-configurable).

## The three-place wiring — miss one and it silently breaks

Adding a pane touches **three** places. Both failure modes hide themselves, so do
all three:

1. **Register** it in the screen's registry (`Registry.Pane`). Miss this →
   `BuildScreen` returns an "unknown pane" error, and consumers typically fall back
   to the default spec, so the bug is invisible.
2. **List** its key in the screen's **default spec** (the `ScreenSpec.Panes` used
   when there's no saved/config layout). Miss this → the pane is registered but
   never placed, so it silently never renders.
3. **Inset the frame** inside the pane's `Render` before drawing a bordered
   `Panel`/`Box`, so borders don't overflow the pane's allotted width (see below).

## The `(Pane, bool)` return is a VISIBILITY predicate, not success

```go
r.Pane("market", "Market", func(c gamePaneCtx) (layout.Pane, bool) {
    s := c.m.econ.Get()
    return layout.Pane{...}, s.Eggs > 0   // ← false HIDES the pane this frame
})
```

The second return decides whether the pane appears **right now**, given runtime
state. Return `true` for always-visible panes. Because visibility is dynamic, the
focus ring can change between frames — never cache a pane index; re-derive it (see
**viewkit-screen**).

## The `cellFrame` inset convention

The layout hands your `Render` a `layout.Frame` sized to the pane's slot. A
bordered `Panel`/`Box` adds width, so inset first. goose uses a small helper —
copy this pattern:

```go
// internal/game/screen_spec.go
func cellFrame(f layout.Frame) layout.Frame {
    inner := layout.NewFrame(f.Width - 4) // leave room for border + padding
    if f.Focused {
        inner = inner.Focus()             // propagate focus so PanelFocus style applies
    }
    return inner
}
```

Always carry `Focused` through, or focused panes won't highlight.

## Minimal correct example

Registry (reference: `internal/game/panes.go`):

```go
type paneCtx struct{ m *Model }

func buildPanes() *layout.Registry[paneCtx] {
    r := layout.NewRegistry[paneCtx]()
    r.Pane("status", "Status", func(c paneCtx) (layout.Pane, bool) {
        return layout.Pane{
            Name:        "status",
            Title:       "Status",
            Interactive: true,        // set to join the focus ring
            MinTier:     layout.TierMedium, // hide when terminal is short
            Render: func(f layout.Frame) string {
                vk := cellFrame(f)
                return vk.Panel("Status", vk.Row("tokens", "1.2M"))
            },
        }, true                        // visibility predicate
    })
    return r
}
```

Default spec (reference: `internal/game/layout_config.go`):

```go
func defaultSpec(id string) layout.ScreenSpec {
    return layout.ScreenSpec{
        Layout: "single",             // or "flex-columns" / "flex-rows" / "grid"
        Panes: []layout.PaneRef{
            {Key: "status"},          // ← the new pane's key MUST appear here
        },
    }
}
```

## Pane fields

- `Interactive bool` — pane joins the tab focus ring; its frame gets `.Focus()`
  when selected.
- `MinTier Tier` — dropped when the terminal is shorter than this tier.
- `Pos *GridPos` — placement for the `grid` layout only (`{Col, Row, ColSpan, RowSpan}`).
- `Slim bool` — narrows the pane in multi-column layouts.
- `Name` must match the key used in the ring/spec; `Title` is display text.

## Verification

`go build ./...` then run the app (or a render test, see **viewkit-test**) and
confirm the pane appears, borders don't overflow, and — if `Interactive` — that
tab cycles onto it. `go test ./...` from the repo root.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
