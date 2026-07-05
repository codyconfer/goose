package layout_test

import "github.com/codyconfer/viewkit/layout"

func ExampleFrame() {
	f := layout.NewFrame(60)

	body := layout.Stack(
		f.Header("STATUS", "local"),
		f.Panel(
			"TOKENS",
			f.Row("balance", "12"),
			f.Row("rate", "0.8/s"),
		),
	)

	_ = layout.ViewportLayout(body, 12, 0)
}
