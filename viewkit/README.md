# viewkit

A small, dependency-light toolkit for building **terminal UIs** with
[Bubble Tea](https://github.com/charmbracelet/bubbletea) /
[Lip Gloss](https://github.com/charmbracelet/lipgloss). It gives you three
things:

1. **`layout`** — the `Frame` render primitive plus structural tools: headers,
   rules, sections, panels/boxes, sticky footers, a width-clamped scrollable
   body/viewport, scroll state, responsive height "tiers", and a focus `Ring`
   for tab-cycling between panels (`Frame`, `Header`, `Panel`, `Section`,
   `SplitStickyFooter`, `ScrollableBody`, `ScrollState`, `Tier`, `Ring`,
   `ScreenFrame`, `TooNarrow`, …).
2. **`panels`** — data visualization that renders neutral structs / `[]float64`
   into aligned, styled charts: bar, line, candlestick (OHLC), pie/proportion,
   and ledger, plus small formatting widgets (`Meter`, `Toggle`, `Flash`).
3. **`theme`** — a `theme.Theme` value (palette + copy) with `theme.Default()`;
   install your own with `theme.Use` to restyle everything.

viewkit depends only on `charmbracelet/lipgloss`, `charmbracelet/x/ansi`, and
the standard library. The package dependency flows `panels → layout → theme`.

## Design contract

viewkit is deliberately domain-agnostic: data crosses the boundary as neutral
structs (`panels.Datum`, `panels.OHLC`, `panels.LedgerRow`) and formatter
callbacks (`func(float64) string`) — never your application's domain types. A
`layout.Frame` carries the render width and focus state; construct one with
`layout.NewFrame(width)` and pass it to the charts.

```go
frame := layout.NewFrame(80)

// structural layout — methods on the frame
body := frame.Panel("STATUS", frame.Row("tokens", "1.2M"))

// charts — functions that take the frame
chart := panels.Bar(frame, "GPUs", []panels.Datum{
    {Label: "gpu", Value: 12},
    {Label: "cloud", Value: 30},
}, 40, fmtNum, "no data")
```

### Theming

The palette and UI copy are injectable. Build a `theme.Theme` from
`theme.Default()`, override what you want, and install it once at startup with
`theme.Use` — every panel and layout helper renders from the active theme
(`theme.Cur()`):

```go
th := theme.Default()
th.Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
th.TooNarrowTitle = "SCREEN TOO SMALL"
theme.Use(th)
```

Structural dimensions (`theme.BodyWidth`, `theme.MinScreenWidth`, the height
tiers, …) are exported constants — viewkit's layout contract — rather than part
of `Theme`; set per-view width via `layout.NewFrame(width)`.

## Status

Currently developed inside the [goose](https://github.com/codyconfer/goose)
repository as a separate Go module, wired via a workspace (`go.work`). It is
intended to move to its own repository once its API stabilizes.

## Linting

The module ships its own `.golangci.yml`. While it lives in the goose
workspace it is vetted via `go vet` and tested with `go test ./...` from the
module root; run `golangci-lint run` from the `viewkit/` directory once it has
a standalone toolchain (or from the workspace root).
