---
name: viewkit-theme
description: >-
  Work with viewkit themes, palettes, colors, and the screen background. Use when
  touching theme.Use / theme.Cur, the exported style vars (theme.DimSty,
  AccentSty, PanelSty, …), theme.Palette / theme.New, theme.Named / theme.Keys /
  theme.DisplayName, theme.Screen, or the layout-contract constants
  (BodyWidth, MinScreenWidth, tiers). Covers the global-singleton model, adding a
  palette, painting the background, and the mandatory test restore.
---

# viewkit theming

The theme is a **process-global singleton**. There is no `Theme` value to thread —
install one at startup and everything reads the active theme.

## Install once, read globally

```go
theme.Use(theme.Default())              // or theme.New(myPalette), or a named theme
```

`theme.Use` sets the current theme **and** syncs a set of exported style vars.
Read styles either way:

```go
theme.Cur().Dim.Render("quiet")         // via the active Theme
theme.DimSty.Render("quiet")            // via the synced exported var (equivalent)
theme.AccentSty.Render("loud")
```

Anti-pattern: adding a `Theme` parameter to your render functions. Don't — read the
globals.

## A Theme is built from a Palette

`theme.Palette` is 11 named color roles; `theme.New(p)` maps them to lipgloss
styles. Build a custom look by constructing a palette:

```go
p := theme.Palette{
    Accent:   lipgloss.Color("#6e9fff"),
    Border:   lipgloss.Color("#44474e"),
    Muted:    lipgloss.Color("#9c9fa3"),
    Text:     lipgloss.Color("#ececed"),
    Selected: lipgloss.Color("#ff9900"),
    Success:  lipgloss.Color("#6ccf8e"),
    Warning:  lipgloss.Color("#fbad37"),
    Failure:  lipgloss.Color("#ff5286"),
    Info:     lipgloss.Color("#6e9fff"),
    Series2:  lipgloss.Color("#d4a0ff"),
    Bg:       lipgloss.Color("#1c1e26"),  // "" ⇒ no background paint
}
theme.Use(theme.New(p))
```

You can also start from a preset and override a few fields:

```go
th := theme.Default()
th.Accent = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
th.TooNarrowTitle = "SCREEN TOO SMALL"   // UI copy is overridable too
theme.Use(th)
```

## Named palettes

Built-in keys: `default` (Munin), `solarized-dark`, `solarized-light`,
`one-dark-vivid`, `monokai`, `classic`.

```go
keys := theme.Keys()                     // []string of keys
if t, ok := theme.Named("monokai"); ok { // false ⇒ returns Default()
    theme.Use(t)
}
label := theme.DisplayName("monokai")    // "Monokai"
```

To add a **new named** palette, edit the `registry` slice in
`viewkit/theme/palettes.go` (it is package-private, so a consumer can't register a
named theme from outside — for a one-off, just `theme.Use(theme.New(p))`).

## Background

The background color is painted by `theme.Screen`, not by a style you apply
yourself. Call it once at the top of your model's `View()`:

```go
return theme.Screen(body, m.width, m.height) // fills palette Bg across the screen
```

If the palette's `Bg` is `""`, it's a no-op (transparent).

## Layout-contract constants (not part of Theme)

Structural dimensions are exported **constants**, deliberately outside `Theme`:
`BodyWidth=81`, `MinBodyWidth=24`, `MinScreenWidth=80`, `MinBodyHeight=35`,
`TallBodyHeight=46`, `AppMarginX=2`, `AppMarginY=1`, `ScreenPaddingWidth`,
`RuleWidth`. Set per-view width with `layout.NewFrame(width)`; use the constants
for width/height guards (`layout.FitsScreenWidth`, tiers) rather than hardcoding.

## Persisting a theme choice

The consumer owns persistence. goose stores the chosen key in `~/.goose/theme.json`
and re-applies it on load (reference: `internal/game/theme_config.go`):

```go
if t, ok := theme.Named(cfg.Theme); ok {
    theme.Use(t)
} else {
    theme.Use(theme.Default())
}
```

## Tests MUST restore global state

Because the theme is global, a test that calls `theme.Use` leaks into later tests.
Always restore:

```go
func TestSomething(t *testing.T) {
    defer theme.Use(theme.Default())     // ← required
    theme.Use(mustNamed("solarized-light"))
    // ...
}
```

See **viewkit-test** for the full testing patterns (color-profile pinning, etc.).

## Verification

`go build ./...`; run the app and confirm colors/background apply and switching
themes restyles everything. Run `go test ./...` from the repo root and confirm no
cross-test color bleed (a symptom of a missing `defer theme.Use(...)`).

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
