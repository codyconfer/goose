package theme

import "github.com/charmbracelet/lipgloss"

var (
	muninPalette = Palette{
		Accent:   lipgloss.Color("#6e9fff"),
		Border:   lipgloss.Color("#44474e"),
		Muted:    lipgloss.Color("#9c9fa3"),
		Text:     lipgloss.Color("#ececed"),
		Selected: lipgloss.Color("#ff9900"),
		Success:  lipgloss.Color("#6ccf8e"),
		Warning:  lipgloss.Color("#fbad37"),
		Failure:  lipgloss.Color("#ff5286"),
		Info:     lipgloss.Color("#6e9fff"),
		Series2:  lipgloss.Color("#d4a0ff"),
		Bg:       lipgloss.Color("#1c1e26"),
	}

	solarizedDarkPalette = Palette{
		Accent:   lipgloss.Color("#268bd2"),
		Border:   lipgloss.Color("#073642"),
		Muted:    lipgloss.Color("#586e75"),
		Text:     lipgloss.Color("#93a1a1"),
		Selected: lipgloss.Color("#cb4b16"),
		Success:  lipgloss.Color("#859900"),
		Warning:  lipgloss.Color("#b58900"),
		Failure:  lipgloss.Color("#dc322f"),
		Info:     lipgloss.Color("#2aa198"),
		Series2:  lipgloss.Color("#6c71c4"),
		Bg:       lipgloss.Color("#002b36"),
	}

	solarizedLightPalette = Palette{
		Accent:   lipgloss.Color("#268bd2"),
		Border:   lipgloss.Color("#eee8d5"),
		Muted:    lipgloss.Color("#93a1a1"),
		Text:     lipgloss.Color("#657b83"),
		Selected: lipgloss.Color("#cb4b16"),
		Success:  lipgloss.Color("#859900"),
		Warning:  lipgloss.Color("#b58900"),
		Failure:  lipgloss.Color("#dc322f"),
		Info:     lipgloss.Color("#2aa198"),
		Series2:  lipgloss.Color("#6c71c4"),
		Bg:       lipgloss.Color("#fdf6e3"),
	}

	oneDarkVividPalette = Palette{
		Accent:   lipgloss.Color("#61afef"),
		Border:   lipgloss.Color("#3e4451"),
		Muted:    lipgloss.Color("#5c6370"),
		Text:     lipgloss.Color("#abb2bf"),
		Selected: lipgloss.Color("#ff9d5c"),
		Success:  lipgloss.Color("#a5e075"),
		Warning:  lipgloss.Color("#f0c674"),
		Failure:  lipgloss.Color("#ff616e"),
		Info:     lipgloss.Color("#4cd1e0"),
		Series2:  lipgloss.Color("#de73ff"),
		Bg:       lipgloss.Color("#282c34"),
	}

	monokaiPalette = Palette{
		Accent:   lipgloss.Color("#66d9ef"),
		Border:   lipgloss.Color("#49483e"),
		Muted:    lipgloss.Color("#75715e"),
		Text:     lipgloss.Color("#f8f8f2"),
		Selected: lipgloss.Color("#fd971f"),
		Success:  lipgloss.Color("#a6e22e"),
		Warning:  lipgloss.Color("#e6db74"),
		Failure:  lipgloss.Color("#f92672"),
		Info:     lipgloss.Color("#66d9ef"),
		Series2:  lipgloss.Color("#ae81ff"),
		Bg:       lipgloss.Color("#272822"),
	}

	classicPalette = Palette{
		Accent:   lipgloss.Color("220"),
		Border:   lipgloss.Color("245"),
		Muted:    lipgloss.Color("245"),
		Text:     lipgloss.Color("252"),
		Selected: lipgloss.Color("220"),
		Success:  lipgloss.Color("78"),
		Warning:  lipgloss.Color("208"),
		Failure:  lipgloss.Color("203"),
		Info:     lipgloss.Color("141"),
		Series2:  lipgloss.Color("141"),
		Bg:       lipgloss.Color(""),
	}

	retroDarkPalette = Palette{
		Accent:   lipgloss.Color("#29a3dc"),
		Border:   lipgloss.Color("#33322e"),
		Muted:    lipgloss.Color("#8a887c"),
		Text:     lipgloss.Color("#f4f1e8"),
		Selected: lipgloss.Color("#f26522"),
		Success:  lipgloss.Color("#4caf50"),
		Warning:  lipgloss.Color("#f7d417"),
		Failure:  lipgloss.Color("#e6338c"),
		Info:     lipgloss.Color("#29a3dc"),
		Series2:  lipgloss.Color("#9b6dff"),
		Bg:       lipgloss.Color("#0a0a0a"),
	}

	retroLightPalette = Palette{
		Accent:   lipgloss.Color("#16a89c"),
		Border:   lipgloss.Color("#d9cfb5"),
		Muted:    lipgloss.Color("#8a7f66"),
		Text:     lipgloss.Color("#1c1a17"),
		Selected: lipgloss.Color("#e0492e"),
		Success:  lipgloss.Color("#4a9b3f"),
		Warning:  lipgloss.Color("#e8a41c"),
		Failure:  lipgloss.Color("#c0392b"),
		Info:     lipgloss.Color("#16a89c"),
		Series2:  lipgloss.Color("#f28c28"),
		Bg:       lipgloss.Color("#f4ecd8"),
	}
)

type registryEntry struct {
	key     string
	name    string
	palette Palette
}

var registry = []registryEntry{
	{key: "default", name: "Munin Default", palette: muninPalette},
	{key: "solarized-dark", name: "Solarized Dark", palette: solarizedDarkPalette},
	{key: "solarized-light", name: "Solarized Light", palette: solarizedLightPalette},
	{key: "one-dark-vivid", name: "One Dark Vivid", palette: oneDarkVividPalette},
	{key: "monokai", name: "Monokai", palette: monokaiPalette},
	{key: "classic", name: "Classic", palette: classicPalette},
	{key: "retro-dark", name: "Retro Dark", palette: retroDarkPalette},
	{key: "retro-light", name: "Retro Light", palette: retroLightPalette},
}

func Keys() []string {
	out := make([]string, len(registry))
	for i, e := range registry {
		out[i] = e.key
	}
	return out
}

func Named(key string) (Theme, bool) {
	for _, e := range registry {
		if e.key == key {
			return New(e.palette), true
		}
	}
	return Default(), false
}

func DisplayName(key string) string {
	for _, e := range registry {
		if e.key == key {
			return e.name
		}
	}
	return key
}
