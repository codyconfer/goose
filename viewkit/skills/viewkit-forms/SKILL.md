---
name: viewkit-forms
description: >-
  Build text inputs, selects, toggles, and confirm dialogs with viewkit's forms
  package. Use when working with forms.Form / forms.Field / FieldKind
  (FieldText, FieldMultiline, FieldSelect, FieldMultiselect, FieldRadio,
  FieldToggle), Form.Handle / Insert / Values, or forms.Confirm. Covers the
  keys.Action-driven handling and the raw-runes text-entry path.
---

# viewkit forms

`forms` gives interactive inputs driven by `keys.Action`s (the same semantic
actions as the rest of the app) and rendered via `layout`. State lives in the
`*Form` you hold in your model.

## A form is fields you Handle then read

```go
fm := forms.NewForm(
    forms.Field{Key: "name", Label: "Name",  Kind: forms.FieldText},
    forms.Field{Key: "risk", Label: "Risk",  Kind: forms.FieldSelect, Options: []string{"low", "med", "high"}},
    forms.Field{Key: "auto", Label: "Auto",  Kind: forms.FieldToggle},
)
```

Field kinds: `FieldText`, `FieldMultiline`, `FieldSelect`, `FieldMultiselect`,
`FieldRadio`, `FieldToggle`.

## Wire it into your key handler

`Form.Handle(action)` consumes navigation/edit actions (Up/Down move between
fields; Left/Right/Inc/Dec change selects & toggles; Erase backspaces; Confirm
activates) and returns whether it consumed the action. **Typed characters are not
actions** — feed the raw runes in the `!ok` branch with `Insert` (reference:
`internal/game/screen_settings.go:114`):

```go
action, ok := myKeymap().Action(msg.String())
if !ok {
    fm.Insert(string(msg.Runes))   // append typed text to the focused text field
    return nil
}
if fm.Handle(action) {
    return nil                     // form consumed it
}
switch action { /* app actions */ }
```

Read values back by key:

```go
vals := fm.Values()          // map[string]any
name := vals["name"].(string)
// or per-field:
fm.Focused().Text            // current focused text field's contents
```

## Render

```go
body := fm.Render(f, "SETTINGS")                 // inline
over := fm.Overlay(bg, f, "SETTINGS", layout.Center) // floating modal over bg
```

## Confirm dialogs

```go
c := &forms.Confirm{Title: "Delete?", Message: "This can't be undone.",
                    YesLabel: "delete", NoLabel: "cancel"}

switch c.Handle(action) {         // Left/Right toggle; Confirm/Cancel resolve
case forms.Submitted:
    // c.Yes tells you which side
case forms.Cancelled:
    // dismissed
case forms.Pending:
    // still open
}
render := c.Overlay(bg, f, layout.Center)
```

## Verification

`go build ./...`; drive `Handle`/`Insert` in a test (see **viewkit-test**) and
assert `Values()` and rendered output. `go test ./...` — `forms/forms_test.go`
shows every field kind.

Full API: see the `viewkit` skill's [references/api.md](../viewkit/references/api.md).
