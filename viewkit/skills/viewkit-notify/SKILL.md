---
name: viewkit-notify
description: >-
  Show transient toast notifications with viewkit's notify package. Use when
  working with notify.Notification / notify.Tone, notify.Queue (NewQueue, Push,
  Beat, Current, Active), or rendering them with panels.NotificationToast /
  NotificationCard / NotificationOverlay / NotificationPanel. Covers the TTL
  queue and the tick-render loop.
---

# viewkit notify (transient toasts)

`notify` models notifications and a **TTL queue**; `panels` renders them. You hold
a `*notify.Queue` in your model, `Push` messages with a lifetime measured in
"beats", `Beat()` once per tick, and render whatever's `Current()`.

## The model

```go
type Notification struct { Title, Message string; Tone Tone }
// Tone: TonePositive, ToneNeutral, ToneWarning, ToneNegative
```

Construct via the tone constructors or a literal:

```go
notify.Positive("Saved", "flock hatched")
notify.Negative("Margin call", "position liquidated")
notify.Note(notify.ToneWarning, "Heads up", "prices are moving")
```

## The queue: Push / Beat / Current

```go
m.notifs = notify.NewQueue(notifQueueCap)          // once, in your constructor

m.notifs.Push(notify.Positive("Saved", "ok"), notifBeats) // ttl in beats
```

Tick the TTL once per simulation frame, then render the head:

```go
// in Update (per tick):
m.notifs.Beat()

// in View:
if n, ok := m.notifs.Current(); ok {
    return panels.NotificationCard(vk, n)   // or Toast / Overlay
}
```

`Active()` reports whether anything is queued; `Len()` the count.

## Rendering choices

```go
panels.NotificationCard(f, n)                         // bordered card, inline
panels.NotificationToast(f, n)                        // compact toast
panels.NotificationPanel(f, "Alerts", ns)             // a list of notifications
panels.NotificationOverlay(bg, f, n, layout.Center)   // float over existing body
```

Use `NotificationOverlay` to float a toast on top of the current screen without
reflowing it; the others compose as normal sections.

## Verification

`go build ./...`; push a notification and confirm it appears, then that it clears
after its TTL of `Beat()` calls. `go test ./...` — `notify/queue_test.go` covers
TTL expiry.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
