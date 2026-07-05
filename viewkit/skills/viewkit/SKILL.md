---
name: viewkit
description: >-
  Overview and router for building terminal UIs with the viewkit Go toolkit
  (github.com/codyconfer/viewkit — packages layout, panels, theme, keys, forms,
  notify). Read this BEFORE writing or modifying any code that imports a
  viewkit package, or when you see layout.Frame, layout.Pane, layout.Registry,
  layout.ScreenSpec, theme.Use/theme.Cur, keys.Map, panels.Bar/Pie/Line,
  forms.Form, or notify.Queue. Explains the Frame render model, the
  domain-agnostic data contract, the two big gotchas (global theme/keys state;
  render returns strings), and points to the task-specific viewkit-* skills.
---

# viewkit

`viewkit` is a small, dependency-light Go toolkit for **terminal UIs**, built on
`lipgloss`. It emits styled `string`s (Bubble-Tea-compatible, not coupled). It is
deliberately **domain-agnostic**: your data crosses the API as neutral structs
(`panels.Datum`, `panels.OHLC`, `panels.LedgerRow`) plus `func(float64) string`
formatters — never your own domain types.

Module path: `github.com/codyconfer/viewkit`. Import subpackages directly, e.g.
`github.com/codyconfer/viewkit/layout`.

## Package map

| Package  | What it gives you |
|----------|-------------------|
| `layout` | The `Frame` render primitive; structural helpers (`Header`, `Panel`, `Row`, `Stack`, `HintLine`); height **tiers**; scrolling/viewport; the **focus ring**; and the data-driven **pane/layout/screen** system (`Registry`, `ScreenSpec`, `BuildScreen`, `SingleColumn`/`Grid`/`FlexColumns`/`FlexRows`). |
| `panels` | Charts (`Bar`, `Line`, `Candle`, `Pie`, `Ledger`, `Markdown`, `Clock`) and small widgets (`Meter`, `Toggle`, `Flash`, `ProgressBar`) + index helpers (`ClampIndex`, `MoveIndex`, `StepIndex`). |
| `theme`  | The active `Theme` (palette + styles + copy), `theme.Use`/`theme.Cur`, exported style vars (`DimSty`, `AccentSty`, …), named palettes, and `theme.Screen` (background). |
| `keys`   | Semantic `Action`s, `Binding`s, a `Scheme` (`keys.Cur`/`keys.Use`), and a `Map` that resolves input → action and generates footer hints. |
| `forms`  | Interactive `Form`/`Field` and `Confirm` dialogs, driven by `keys.Action`s. |
| `notify` | `Notification`/`Tone` and a TTL `Queue` for transient toasts. |

Dependency flow: `panels`/`forms` → `layout` → `theme`; `keys` and `notify` are leaf helpers.

## The Frame model

Everything renders relative to a `layout.Frame` — it carries the render `Width`,
`Height`, and `Focused` state. Build the top one from the terminal width and thread
narrower child frames down into panes/charts:

```go
f := layout.NewFrame(width)          // clamps to [theme.MinBodyWidth, …]
body := f.Panel("STATUS", f.Row("tokens", "1.2M"))
chart := panels.Bar(f, "GPUs", data, 40, fmtNum, "no data")
```

Structural helpers exist both as free functions (default width) and as `Frame`
methods (use the frame's width) — prefer the method form: `f.Header`, `f.Panel`,
`f.Row`, `f.Spread`, `f.HintLine`, `f.Stack` (via `layout.Stack`).

## Two gotchas that trip up every newcomer

1. **Theme and keys are process-global singletons, not values you thread.**
   Install once at startup with `theme.Use(...)` / `keys.Use(...)`, then read the
   active state via `theme.Cur()`, the exported style vars (`theme.DimSty`,
   `theme.AccentSty`, …), and `keys.Cur()`. Do **not** invent a `Theme` parameter
   to pass around. In tests this means you must restore global state —
   `defer theme.Use(theme.Default())`. See **viewkit-theme** and **viewkit-test**.

2. **Rendering returns strings; there is no widget tree.** A screen's `View()`
   composes strings with `layout.Stack`/`StackFit` and paints the background with
   `theme.Screen(body, w, h)`. State (cursor, scroll offset, focus index) lives in
   *your* model, not in viewkit.

## Which skill to use

- **Add or change a pane** (a content block in a screen) → **viewkit-pane**
- **Build or change a screen / model** (interface, keymap, view assembly, registry) → **viewkit-screen**
- **Themes, palettes, colors, background** → **viewkit-theme**
- **Charts, meters, tables, widgets** → **viewkit-panels**
- **Keybindings and footer hint legends** → **viewkit-keys**
- **Text input, selects, confirm dialogs** → **viewkit-forms**
- **Transient toast notifications** → **viewkit-notify**
- **Testing viewkit-consuming code** → **viewkit-test**

Full public API by package: [references/api.md](references/api.md).

## Reference implementation

`goose/internal/game/` is the canonical consumer. When unsure how a pattern is
wired end-to-end, read the cited files there — and note the `_test.go` suites in
each viewkit package are the de-facto usage examples.

## Verification

viewkit lives in a Go workspace (`go.work` uses `.` and `./viewkit`). After any
change, from the repo root run: `go build ./...`, `go vet ./...`, `go test ./...`.
