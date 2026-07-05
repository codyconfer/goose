---
name: viewkit-screen
description: >-
  Build or modify a viewkit screen / Bubble Tea model — the top-level render +
  input loop. Use when wiring layout.Registry, layout.ScreenSpec, BuildScreen,
  Screen.Render, the focus ring (Ring.At/Step), a keys.Map action-switch in
  handleKey, or assembling a view() with layout.Stack + Frame helpers +
  theme.Screen. Covers the register→spec→build data-driven screen pattern and the
  build-with-default-fallback resilience trick.
---

# Build / modify a viewkit screen

A screen owns three things: **input** (map keys → actions → state changes),
**render** (compose strings), and its **panes** (via a data-driven registry). State
(cursor, scroll offsets, focus index) lives in *your* model — viewkit holds none.

## The data-driven screen pattern

1. Define a **pane context** struct carrying whatever the pane factories need
   (`*Model`, the screen struct, etc.).
2. Build a **registry** once (package-level global) with `layout.NewRegistry[Ctx]()`
   and register panes (see **viewkit-pane**).
3. Each frame, build a live `Screen` from a `ScreenSpec` + the current context via
   `layout.BuildScreen`, then `Screen.Render(frame, tier, focusIndex)`.

Wrap `BuildScreen` so a bad saved/config spec **falls back to the default** instead
of erroring the whole view (reference: `internal/game/panes.go`):

```go
func buildScreen[C any](id string, ctx C, reg *layout.Registry[C]) layout.Screen {
    scr, err := layout.BuildScreen(layoutSpec(id), ctx, reg) // saved/user spec
    if err != nil {
        scr, _ = layout.BuildScreen(defaultSpec(id), ctx, reg) // safe fallback
    }
    return scr
}
```

## The screen contract (consumer-defined)

goose defines a tiny interface its `Model` delegates to (`internal/game/screen.go`):

```go
type screen interface {
    handleKey(m *Model, msg tea.KeyMsg) tea.Cmd
    view(m *Model) string
    simulates() bool
}
```

You don't have to copy this exact shape, but the three responsibilities are the
same in any Bubble Tea app: `Update` routes keys, `View` returns a string.

## Input: keymap + action-switch

Get a `*keys.Map` from a per-screen constructor (layered on `keys.Cur()` — see
**viewkit-keys**), resolve the input, then switch on the semantic action:

```go
func (s *myScreen) handleKey(m *Model, msg tea.KeyMsg) tea.Cmd {
    action, ok := s.keys().Action(msg.String())
    if !ok {
        return nil                      // (for a text field, feed unconsumed runes to a form here)
    }
    switch action {
    case keys.Quit:
        return tea.Quit
    case keys.FocusNext:
        s.focus = s.build(m).Ring().Step(s.focus, 1)
    case keys.FocusPrev:
        s.focus = s.build(m).Ring().Step(s.focus, -1)
    case keys.Confirm:
        // ... mutate model
    }
    return nil
}
```

## Focus is an index resolved against a fresh ring each frame

Store only an `int` focus index. Resolve it to a pane name through the ring built
from the *current* screen — because pane visibility is dynamic, the ring can change
between frames. Never cache a pane index across frames.

```go
func (s *myScreen) focusedPane(m *Model) string {
    return s.build(m).Ring().At(s.focus) // Ring.At is bounds-safe
}
```

## Render: compose sections with Stack

Build a `[]string` of sections and join with `layout.Stack`; the pane body comes
from `Screen.Render` (reference: `internal/game/screen_game.go:151`):

```go
func (s *myScreen) view(m *Model) string {
    sections := []string{
        m.frame().Header("MY APP", "subtitle"),
        s.build(m).Render(m.bodyFrame(), m.heightTier(), s.focus),
        m.frame().HintLine(s.keys().Hint(keys.Confirm), s.keys().Hint(keys.Quit)),
    }
    return layout.Stack(sections...)
}
```

At the very top level (the model's `View()`), guard width and paint the background:

```go
func (m Model) View() string {
    if !layout.FitsScreenWidth(m.width) {
        return theme.Screen(theme.AppFrame.Render(layout.TooNarrow(m.width)), m.width, m.height)
    }
    body := theme.AppFrame.Render(layout.ViewportLayout(m.screen.view(&m), layout.ContentRows(m.height), m.pageScroll))
    return theme.Screen(body, m.width, m.height)
}
```

## Frame plumbing

- `layout.ScreenFrame(width)` — the content frame from a raw terminal width
  (subtracts screen padding).
- Height → tier with `layout.TierForHeight(height)`; pass the tier into
  `Screen.Render` so `MinTier` panes drop on short terminals.
- Per-pane inset happens inside `Render` (see **viewkit-pane** / `cellFrame`).

## Verification

`go build ./...`; run the app or a render test (**viewkit-test**) and confirm: the
screen composes, tab cycles focus across interactive panes, a bad saved spec falls
back to default rather than blanking, and short/narrow terminals degrade (tiers /
`TooNarrow`). `go test ./...` from the repo root.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
