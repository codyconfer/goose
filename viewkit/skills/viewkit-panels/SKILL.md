---
name: viewkit-panels
description: >-
  Render charts and widgets with viewkit's panels package. Use when calling
  panels.Bar / Line / Candle / Pie / Spectrum / Ledger / Markdown / Clock, the
  small widgets Meter / Toggle / Flash / ProgressBar, the animated Matrix rain
  (panels.Rain / Matrix), or the index helpers ClampIndex / MoveIndex / StepIndex. Covers the
  neutral-struct + formatter-callback data contract and how charts fit a
  layout.Frame.
---

# viewkit panels (charts & widgets)

`panels` renders **neutral data** into aligned, themed output. Data crosses the
boundary as plain structs (`panels.Datum`, `panels.OHLC`, `panels.LedgerRow`) plus
a `func(float64) string` formatter — never your domain types. Every chart takes a
`layout.Frame` (for width/theme) and returns a string.

## The data contract

```go
type Datum     struct { Label string; Value float64 }
type OHLC      struct { Open, High, Low, Close float64 }
type LedgerRow struct { Label string; Delta float64 }
```

Convert your domain objects into these at the render boundary; supply a formatter
for numbers (e.g. `func(v float64) string { return fmt.Sprintf("%.0f", v) }`).

## Charts

```go
// Bar chart (last arg is the empty-state message)
panels.Bar(f, "GPUs", []panels.Datum{
    {Label: "gpu", Value: 12}, {Label: "cloud", Value: 30},
}, 40, fmtNum, "no data")

// Scrollable bar (visible rows, offset)
panels.BarScroll(f, "GPUs", data, 40, fmtNum, "no data", 8, offset)

// Line plot (width, height, optional footer lines)
panels.Line(f, "Price", series, f.Width, 8, fmtVal, "24h")

// Candlestick / OHLC
panels.Candle(f, "OHLC", candles, f.Width, 10, fmtVal)

// Pie / proportion (barWidth — clamp it to something reasonable)
panels.Pie(f, "Mix", data, min(f.Width, 48), fmtNum, "empty")

// Spectrum analyzer / equalizer (levels & peaks are per-band magnitudes in [0,1])
panels.Spectrum(f, "EQ", levels, 6, "silent",
    panels.SpectrumOpts{Peaks: peaks, BarWide: 2, BarGap: 1})

// Ledger table (unit, visible rows, offset, empty msg)
panels.Ledger(f, "Flows", rows, "🪙", fmtNum, 8, offset, "no rows")

// Markdown (and a titled panel variant)
panels.Markdown(f, mdSource)
panels.MarkdownPanel(f, "Docs", mdSource)

// Clocks
panels.Clock(f, "UTC", t, panels.ClockOpts{TwentyFour: true, HideSeconds: true})
panels.BinaryClock(f, "BIN", t)
```

## Animation (matrix rain)

`Rain` is a **stateful** "matrix rain" grid — the animated-state half of a
state+render pair, exactly like `notify.Queue` + `panels.NotificationToast`. Build
it once, advance one frame per tick with `Beat()`, and render with `Matrix`:

```go
r := panels.NewRain(f.BodyWidth(), 10, time.Now().UnixNano()) // width, rows, seed
r.Beat()                                                      // once per tick
panels.Matrix(f, "MATRIX", r)                                 // render this frame
```

Drive `r.Beat()` from your Bubble Tea tick loop — the same place you call
`notifs.Beat()` — and `r.Resize(f.BodyWidth(), rows)` on `tea.WindowSizeMsg`. Rows
are explicit (not derived from `Frame.Height`); the panel border adds 2 lines.
Colors adapt to the active theme (bright head → `Accent`, body → `Can`, tail →
`Dim`). A fixed seed makes the animation reproducible for tests.

All chart renderers show the empty-state string (or degrade gracefully) when data
is missing — pass a sensible `empty` message rather than pre-checking length.

## Small widgets

```go
panels.Meter(frac, panels.MeterWidth(f.Width, 22)) // progress meter; frac in [0,1]
panels.ProgressBar(frac, width)
panels.Toggle("PUT", "CALL", leftActive)            // two-state pill
panels.Flash("saved!")                              // transient one-liner
```

`panels.MeterWidth(frameWidth, desired)` clamps a desired width to what fits.

## Index helpers (for cursors / selections)

```go
i = panels.ClampIndex(i, total)      // clamp into [0, total)
i = panels.MoveIndex(i, +1, total)   // step, CLAMPS at the ends
i = panels.StepIndex(i, +1, total)   // step, WRAPS around
```

Use these for menu cursors and carousel selection instead of hand-rolling bounds
math — `MoveIndex` (clamp) vs `StepIndex` (wrap) is the choice that matters.

## Fitting a chart to a pane

Inside a pane's `Render`, work off the inset frame (see **viewkit-pane** /
`cellFrame`) and cap chart width so borders don't overflow. Reference:
`internal/game/screen_spec.go` (`renderBook` clamps the pie width to 48).

## Verification

`go build ./...`; render at a few widths and confirm charts don't overflow their
pane and empty states read correctly. `go test ./...` from the repo root — the
`panels/*_test.go` files show expected output shapes.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
