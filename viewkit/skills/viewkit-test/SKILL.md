---
name: viewkit-test
description: >-
  Test code that consumes viewkit. Use when writing Go tests that render a screen
  and assert on output, verify keymaps/hint legends, check theme/background, or
  sandbox a config file. Covers the render-and-grep idiom, color-profile pinning
  for deterministic ANSI, HOME-sandboxed config tests, and the mandatory
  theme.Use / keys.Use restore that prevents cross-test bleed.
---

# Testing viewkit consumers

viewkit renders to strings, so tests are mostly **render then assert on substrings**.
The only sharp edges are the process-global theme/keys state (must be restored) and
ANSI determinism (pin the color profile).

## Restore global state — always

Theme and keybindings are process globals (`theme.Use`/`keys.Use`). A test that
installs one leaks into every later test. Restore with `defer` at the top:

```go
func TestThing(t *testing.T) {
    defer theme.Use(theme.Default())   // required if the test (or code) calls theme.Use
    // defer keys.Use(keys.Default())  // add if the test changes the key scheme
    ...
}
```

Missing this is the #1 cause of flaky, order-dependent viewkit tests.

## Render-and-grep idiom

Force a known width/height, render, assert substrings. A tiny helper keeps tests
terse (reference: `internal/game/screen_hints_test.go`):

```go
func renderForHints(m Model) string {
    m.width = theme.MinScreenWidth   // ≥ min so it doesn't hit the TooNarrow path
    m.height = 80
    return m.View()
}

func TestFooterShowsHints(t *testing.T) {
    m := New(...)
    view := renderForHints(m)
    for _, want := range []string{"enter/space", "↑/↓/j/k", "esc/q"} {
        if !strings.Contains(view, want) {
            t.Fatalf("view missing %q:\n%s", want, view)
        }
    }
}
```

Asserting on the glyph strings is how you verify a **keymap is wired** without
simulating input — the footer legend is generated from the bindings.

## Exercise a specific screen in isolation

Construct the model, then assign the screen struct directly and drive
`handleKey`:

```go
m := New(...)
s := &mySettingsScreen{}
m.screen = s
s.handleKey(&m, tea.KeyMsg{Type: tea.KeyRight})   // simulate a keypress
```

## Deterministic ANSI: pin the color profile

To assert on raw escape codes (e.g. a background fill), pin lipgloss to TrueColor
and restore it (reference: `internal/game/theme_picker_test.go:59`):

```go
prev := lipgloss.ColorProfile()
lipgloss.SetColorProfile(termenv.TrueColor)
defer lipgloss.SetColorProfile(prev)
defer theme.Use(theme.Default())

th, _ := theme.Named("solarized-light")
theme.Use(th)
if !strings.Contains(renderForHints(m), "48;2;253;246;227") { // the bg RGB
    t.Fatal("View() missing themed background fill")
}
```

Compare theme colors via `theme.Cur().Accent.GetForeground()` and a
`theme.Named(key)` helper rather than hardcoding hex.

## Sandbox config files with a temp HOME

If the code reads/writes `~/.<app>/…`, redirect HOME to a temp dir so tests don't
touch the real home and can assert on what was written:

```go
home := t.TempDir()
t.Setenv("HOME", home)
// ... drive the code that persists ...
data, err := os.ReadFile(filepath.Join(home, ".goose", "theme.json"))
```

## Verification

Run `go test ./...` from the repo root. Then run it again with `-count=1 -shuffle=on`
— if results change with order, a test is missing a `defer theme.Use(...)` /
`keys.Use(...)` restore.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
