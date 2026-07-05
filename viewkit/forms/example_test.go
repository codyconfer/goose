package forms_test

import (
	"github.com/codyconfer/viewkit/forms"
	"github.com/codyconfer/viewkit/keys"
	"github.com/codyconfer/viewkit/layout"
)

func ExampleForm() {
	fm := forms.NewForm(
		forms.Field{Key: "name", Label: "Name", Kind: forms.FieldText},
		forms.Field{
			Key:     "risk",
			Label:   "Risk",
			Kind:    forms.FieldSelect,
			Options: []string{"low", "med", "high"},
		},
		forms.Field{Key: "auto", Label: "Auto", Kind: forms.FieldToggle},
	)

	fm.Insert("goose")
	fm.Handle(keys.Down)
	fm.Handle(keys.Right)
	fm.Handle(keys.Down)
	fm.Handle(keys.Confirm)

	_ = fm.Values()
	_ = fm.Render(layout.NewFrame(60), "SETTINGS")
}
