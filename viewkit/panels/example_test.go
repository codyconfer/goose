package panels_test

import (
	"fmt"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/panels"
)

func ExampleBar() {
	f := layout.NewFrame(60)
	fmtNum := func(v float64) string { return fmt.Sprintf("%.0f", v) }

	_ = panels.Bar(f, "GPUs", []panels.Datum{
		{Label: "gpu", Value: 12},
		{Label: "cloud", Value: 30},
	}, 24, fmtNum, "no data")

	_ = panels.MarkdownPanel(f, "Notes", "- watch supply\n- watch price")
}
