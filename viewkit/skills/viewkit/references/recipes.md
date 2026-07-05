# viewkit recipes

Task-oriented entry points for common viewkit work. Start with the main
`viewkit` skill, then load the narrower skill for the concrete task.

## Add a new pane to an existing screen

1. Read `viewkit` for the frame model and the global theme/keys gotchas.
2. Read `viewkit-pane` for the register -> default spec -> render wiring.
3. If the pane needs charts or widgets, also read `viewkit-panels`.
4. Verify with `go test ./...` and tab through the focus ring if the pane is
   interactive.

## Build a new screen or rework a screen model

1. Read `viewkit`.
2. Read `viewkit-screen` for the screen build pattern and focus handling.
3. Read `viewkit-keys` if the screen adds or changes actions and footer hints.
4. Read `viewkit-test` to add a render-and-grep test for the new screen.

## Add a settings modal, form, or confirmation flow

1. Read `viewkit-forms`.
2. Read `viewkit-screen` for where `Handle` and `Insert` belong in the key loop.
3. If the modal changes global styling, read `viewkit-theme`.

## Add charts, ledgers, or compact widgets

1. Read `viewkit-panels`.
2. Convert domain data at the render boundary into `panels.Datum`,
   `panels.OHLC`, or `panels.LedgerRow`.
3. Use the pane's inset frame so bordered panels and chart widths stay inside
   their slot.

## Add notifications

1. Read `viewkit-notify`.
2. Hold a `*notify.Queue` in model state.
3. Push notifications on events, call `Beat()` on each tick, and render
   `Current()` with a panel or overlay.

## Change the theme or add a palette

1. Read `viewkit-theme`.
2. Treat `theme.Use(...)` as startup-global state, not a dependency to thread
   through render helpers.
3. Restore `theme.Use(theme.Default())` in any test that changes the active
   theme.
