---
name: viewkit-keys
description: >-
  Define keybindings and footer hint legends with viewkit's keys package. Use
  when working with keys.Action, keys.Binding, keys.NewMap / Map.Action,
  keys.Cur / keys.Use / Scheme, or generating hints with Map.Hint / HintLabeled /
  Hints for layout.HintLine. Covers semantic actions, the glyph/label grouping
  rule, and how hints feed the footer legend.
---

# viewkit keys & hints

Input is modeled as **semantic actions**, not raw keys. You map keys → actions in a
`keys.Map`, switch on the action in your handler (see **viewkit-screen**), and
generate footer hints from the same bindings so the legend can't drift from the
bindings.

## Actions and the shared scheme

Predefined navigation actions live in the package: `Up, Down, Left, Right,
Confirm, Cancel, Quit, FocusNext, FocusPrev, Inc, Dec, Erase, PageUp, PageDown`.
Define app-specific actions as your own `keys.Action` string constants, namespaced:

```go
const (
    actBuy  keys.Action = "game.buy"
    actSell keys.Action = "game.sell"
)
```

`keys.Cur()` returns the active `Scheme` (default glyphs/keys for the shared
actions). Build a per-screen `*keys.Map` layering app bindings on the scheme:

```go
func gameKeymap() *keys.Map {
    sc := keys.Cur()
    return keys.NewMap(
        keys.Binding{Keys: []string{"ctrl+c", "q", "esc"}, Action: keys.Quit, Glyph: "esc/q", Label: "quit"},
        sc.Binding(keys.Confirm).WithLabel("generate"),   // reuse scheme glyph, set label
        sc.Binding(keys.Up),
        sc.Binding(keys.Down),
        keys.Binding{Keys: []string{"b", "right", "l"}, Action: actBuy, Glyph: "b/→/l", Label: "buy"},
    )
}
```

`sc.Binding(action).WithGlyph(...)/.WithLabel(...)` returns a modified copy — chain
to customize display without redefining keys.

## Resolve input

```go
action, ok := m.Action(msg.String())  // *keys.Map
if ok { /* switch on action */ }
```

## The glyph/label grouping rule

When several keys share one displayed hint, only the **first** binding in the group
carries the `Glyph`/`Label`; the siblings that map extra keys to related actions
omit them so the footer isn't duplicated:

```go
keys.Binding{Keys: []string{"B"}, Action: actMaxBuy, Glyph: "B/S", Label: "max queue"},
keys.Binding{Keys: []string{"S"}, Action: actMaxSell},              // no glyph/label
```

`Map.Hints(actions...)` intentionally emits only bindings that have a non-empty
`Glyph`, which is what implements this grouping in the legend.

## Footer hints

`Map` produces `[2]string{glyph, label}` pairs that feed `layout.Frame.HintLine`:

```go
f.HintLine(
    m.Hint(keys.Confirm),                 // {DisplayGlyph, Label}
    m.HintLabeled(keys.Up, "select"),     // override the label at the call site
    m.Hint(keys.Quit),
)
// or many at once (only bindings with a Glyph are included):
f.HintLine(m.Hints(keys.Confirm, actBuy, actSell)...)
```

`Binding.DisplayGlyph()` falls back to `strings.Join(Keys, "/")` when `Glyph` is
empty.

## Custom scheme (optional)

To change global defaults, install a scheme once at startup with `keys.Use(...)`
(`Scheme.With(overrides...)` for tweaks). Like theme, this is global — restore it
in tests with `defer keys.Use(keys.Default())`.

## Verification

`go build ./...`; render the screen (see **viewkit-test**) and assert the expected
glyph strings appear in the footer — that verifies the map is wired. `go test ./...`.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
