package theme_test

import "github.com/codyconfer/viewkit/theme"

func ExampleUse() {
	defer theme.Use(theme.Default())

	if th, ok := theme.Named("monokai"); ok {
		theme.Use(th)
	}

	body := theme.Cur().Accent.Render("ready")
	_ = theme.Screen(body, theme.MinScreenWidth, theme.MinBodyHeight)
}
