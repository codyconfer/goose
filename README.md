# 🪿 goose

[![GitHub Release](https://shields.io)](https://github.com)
 [![CI](https://github.com/codyconfer/goose/actions/workflows/ci.yml/badge.svg)](https://github.com/codyconfer/goose/actions/workflows/ci.yml) [![Release](https://github.com/codyconfer/goose/actions/workflows/release.yml/badge.svg)](https://github.com/codyconfer/goose/actions/workflows/release.yml)

A tiny terminal **idle-clicker economy** where **tokens** are the currency and
**golden goose eggs** are the goods you sell.

Press `enter` to earn tokens. Spend tokens on GPUs, servers, data centers and
clouds that earn tokens *and* lay eggs for you. The flock keeps working while you watch
— and even while the game is closed (offline progress, capped at 10 minutes).

Built with [Cobra](https://github.com/spf13/cobra) and
[Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Run

```sh
go run .
go run . play
```

Install it:

```sh
go install github.com/codyconfer/goose@latest
goose
```

## Controls

| Key        | Action                          |
| ---------- | ------------------------------- |
| `enter`    | Earn a token                    |
| `↑` / `↓`  | Move the market selection       |
| `b` / `→`  | Buy the selected item           |
| `t`        | Open the trade desk             |
| `q` / `esc`| Quit (your progress is saved)   |

Save menu:

| Key        | Action                          |
| ---------- | ------------------------------- |
| `enter`    | Load the selected save          |
| `n`        | Create a new save               |
| `r`        | Rename the selected save        |
| `x`        | Delete the selected save        |
| `q` / `esc`| Quit                            |

## The economy

- **Tokens** (🪙) are the currency. Everything is bought with tokens.
- **Enter the Flow State** doubles how many tokens you earn per press.
- **Producers** (GPUs, servers, racks, data centers, clouds) earn tokens
  passively. Each one you buy costs 15% more than the last, the classic
  idle-clicker curve. New producers reveal themselves as you grow.

### Eggs & consumers

Producers don't just earn tokens — they also lay **eggs** (🥚) as a side
product. A town of **consumers** buys those eggs from the flock for tokens, and
they trade through the very same market queue you do:

- The buying **crowd** grows to match your egg output and wanders off if supply
  dries up. Each beat the crowd places a standing order to buy a share of your
  eggs, paying tokens — your egg income stream. It absorbs only part of what you
  lay, so the rest of the hoard piles up toward your next level.
- The **market price** (tokens per egg) is driven by **market interest** —
  buyer demand measured against the eggs available to meet it. Eager buyers and
  thin supply bid the price up; a glut of unsold eggs drags it down. Hoard a
  mountain of eggs and the price sags; sell them off and it recovers. On top of
  that fundamental the price carries sticky **market sentiment**: noisy ticks
  still kick it around every few seconds, but rallies and selloffs now tend to
  bleed across multiple candles before fundamentals pull them back. Watch for
  trend follow-through before you buy, sell, or write derivatives.
- **🧥 Jet Set Huang** upgrades raise the price consumers will pay (+30% each).
- The 🛒 **Market Day** event lets you snap up dirt-cheap eggs at a glut price.

So there are two income streams: tokens you *earn* (by hand and from the flock),
and tokens consumers pay you for *eggs*. The market panel shows your egg stock,
laying rate, sales per second, the crowd size, and the live price.

### Trade desk

Press `t` to open the **trade desk** and place your own orders on the market
alongside the crowd:

- A live **candlestick chart** plots the egg price — green candles where it
  closed up, red where it closed down, with high/low wicks — so you can read the
  trend before you trade. Each candle aggregates a few price re-rolls.
- **Buy** orders spend tokens to stock eggs (handy for rushing a level); **sell**
  orders cash eggs in for tokens at the consumer price. Use `←`/`→` to pick a
  direction, `↑`/`↓` to size the order (presets up to a balance-aware **Max**),
  and `enter` to queue it.
- The queue is worked **every beat**, each order filling at the market's
  throughput (the crowd's appetite, plus a small floor so even a young flock can
  trade) and capped by what your balance or hoard can back. A buy order patiently
  waits for tokens; a sell order waits for eggs.
- `x` cancels the active order, `c` clears the queue. The trade queue panel shows
  the consumer crowd's standing order at the top and your own orders below it,
  with live progress, on both the desk and the main screen.

Selling eggs never costs you a level: levels are earned on your **all-time high**
egg count, so spending your hoard down is always safe.

### Events

On every heartbeat there's a small chance something happens to your flock —
a lucky egg, a golden hour, a market boom, a wandering goose, a market day, or a
fox raid that makes off with some tokens. See `internal/events/events.go`.

### Venture capitalists

Once your flock is established, a **venture capitalist** may waddle in (rarely).
This pauses the idle screen and opens a **negotiation view** where you choose
how to respond:

- **Take the money** — usually a clean token injection plus new customers, but
  the terms occasionally bite (they claw the tokens back and seize a producer).
- **Counter-offer** — high stakes: double the cheque and a free Golden Goose,
  or they storm off and cost you tokens and consumers.
- **Politely decline** — safe; keep the flock independent.

Each option has random positive and negative outcomes. The encounter and its
choices live in `internal/characters/vc.go`; the dialog view is in
`internal/game/model.go`.

Named saves live in `~/.goose/save.db`. Older one-slot saves are imported as
`Flock 1` the first time the new save manager opens them.

## Develop

```sh
go test ./...
go build ./...
```

### Pre-commit hooks

A tracked hook formats, lints, and tests staged Go changes before each commit.
Enable it once per clone (installs `golangci-lint` and points git at
`.githooks/`):

```sh
make hooks
```

Handy targets: `make fmt` (gofmt + goimports), `make lint` (`go vet` +
golangci-lint), `make test`, and `make check` (all three). Lint rules live in
`.golangci.yml`; the hook itself is `.githooks/pre-commit`.

## Release builds

```sh
make version
make release
```

Release artifacts are written to `dist/`:

- Linux x64 AppImage: `goose_<version>_linux_x86_64.AppImage`
- Linux arm64 AppImage: `goose_<version>_linux_aarch64.AppImage`
- macOS arm64 tarball: `goose_<version>_darwin_arm64.tar.gz`
- Windows x64 zip: `goose_<version>_windows_amd64.zip`
- Windows arm64 zip: `goose_<version>_windows_arm64.zip`

Versions are derived automatically from git tags. A clean build on tag
`v1.2.3` produces `1.2.3`; commits after that tag produce a patch-dev version
such as `1.2.4-dev.3+gabc1234`; dirty worktrees add `dirty` build metadata.
Before the first tag, builds start from `0.1.0-dev.<commit-count>+g<commit>`.

Linux AppImage targets require `appimagetool` on `PATH`, or set
`APPIMAGETOOL=/path/to/appimagetool`. Windows archive targets require `zip`.
Build metadata is compiled into the binary and is available with
`goose --version`.

## Contributing

Contributions are welcome — see [CONTRIBUTORS.md](CONTRIBUTORS.md).

## License

🪿 goose is licensed under the [GNU Affero General Public License v3.0](LICENSE).
