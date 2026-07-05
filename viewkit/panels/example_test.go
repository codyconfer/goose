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

func ExampleSpectrum() {
	f := layout.NewFrame(40)

	levels := []float64{0.2, 0.6, 0.9, 0.5, 0.3, 0.7}
	peaks := []float64{0.4, 0.7, 0.95, 0.8, 0.5, 0.9}
	_ = panels.Spectrum(f, "SPECTRUM", levels, 8, "silent", panels.SpectrumOpts{Peaks: peaks})
}

func ExampleMatrix() {
	f := layout.NewFrame(40)

	r := panels.NewRain(f.BodyWidth(), 10, 1)
	for i := 0; i < 5; i++ {
		r.Beat()
	}
	_ = panels.Matrix(f, "MATRIX", r)
}
