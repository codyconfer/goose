package theme

import "github.com/charmbracelet/lipgloss"

const (
	BodyWidth    = 81
	MinBodyWidth = 24

	MinScreenWidth     = 80
	MinBodyHeight      = 35
	TallBodyHeight     = 46
	AppMarginY         = 1
	AppMarginX         = 2
	ScreenPaddingWidth = AppMarginX*2 + 4
	MinScreenBodyWidth = MinScreenWidth - ScreenPaddingWidth

	RuleWidth = BodyWidth + 4
)

type Palette struct {
	Accent   lipgloss.Color
	Border   lipgloss.Color
	Muted    lipgloss.Color
	Text     lipgloss.Color
	Selected lipgloss.Color
	Success  lipgloss.Color
	Warning  lipgloss.Color
	Failure  lipgloss.Color
	Info     lipgloss.Color
	Series2  lipgloss.Color
	Bg       lipgloss.Color
}

func New(p Palette) Theme {
	return Theme{
		Title:  lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Accent: lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Dim:    lipgloss.NewStyle().Foreground(p.Muted),
		Val:    lipgloss.NewStyle().Foreground(p.Text),
		Key:    lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Can:    lipgloss.NewStyle().Foreground(p.Success),
		Cant:   lipgloss.NewStyle().Foreground(p.Failure),

		Panel:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Border).Padding(0, 1).Width(BodyWidth + 2),
		PanelFocus: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Selected).Padding(0, 1).Width(BodyWidth + 2),
		PanelTitle: lipgloss.NewStyle().Bold(true).Foreground(p.Accent),
		Card:       lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Accent).Padding(0, 1).Width(BodyWidth + 2).Align(lipgloss.Center),
		AppFrame:   lipgloss.NewStyle().Margin(AppMarginY, AppMarginX),

		NotifPositive: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Success).Foreground(p.Success).Padding(0, 1),
		NotifNeutral:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Info).Foreground(p.Info).Padding(0, 1),
		NotifWarning:  lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Warning).Foreground(p.Warning).Padding(0, 1),
		NotifNegative: lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Failure).Foreground(p.Failure).Padding(0, 1),
		NotifIdle:     lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(p.Muted).Foreground(p.Muted).Padding(0, 1),
		NotifTitle:    lipgloss.NewStyle().Bold(true),

		Series: []lipgloss.Style{
			lipgloss.NewStyle().Foreground(p.Accent),
			lipgloss.NewStyle().Foreground(p.Success),
			lipgloss.NewStyle().Foreground(p.Warning),
			lipgloss.NewStyle().Foreground(p.Failure),
			lipgloss.NewStyle().Foreground(p.Series2),
			lipgloss.NewStyle().Foreground(p.Muted),
		},

		Bg: p.Bg,

		TooNarrowTitle: DefaultTooNarrowTitle,
		TooNarrowNeed:  DefaultTooNarrowNeed,
		TooNarrowBody:  DefaultTooNarrowBody,
	}
}

var (
	TitleSty  lipgloss.Style
	AccentSty lipgloss.Style
	DimSty    lipgloss.Style
	ValSty    lipgloss.Style
	KeySty    lipgloss.Style
	CanSty    lipgloss.Style
	CantSty   lipgloss.Style

	AppFrame lipgloss.Style

	PanelSty      lipgloss.Style
	PanelFocusSty lipgloss.Style
	PanelTitleSty lipgloss.Style
	CardSty       lipgloss.Style

	NotifPositiveSty lipgloss.Style
	NotifNeutralSty  lipgloss.Style
	NotifWarningSty  lipgloss.Style
	NotifNegativeSty lipgloss.Style
	NotifIdleSty     lipgloss.Style
	NotifTitleSty    lipgloss.Style

	Series []lipgloss.Style
)

const (
	DefaultTooNarrowTitle = "TERMINAL TOO NARROW"
	DefaultTooNarrowNeed  = "Need at least %d columns."
	DefaultTooNarrowBody  = "Current width: %s columns. Resize the terminal to at least %d characters wide to use this screen."
)

type Theme struct {
	Title  lipgloss.Style
	Accent lipgloss.Style
	Dim    lipgloss.Style
	Val    lipgloss.Style
	Key    lipgloss.Style
	Can    lipgloss.Style
	Cant   lipgloss.Style

	Panel      lipgloss.Style
	PanelFocus lipgloss.Style
	PanelTitle lipgloss.Style
	Card       lipgloss.Style
	AppFrame   lipgloss.Style

	NotifPositive lipgloss.Style
	NotifNeutral  lipgloss.Style
	NotifWarning  lipgloss.Style
	NotifNegative lipgloss.Style
	NotifIdle     lipgloss.Style
	NotifTitle    lipgloss.Style

	Series []lipgloss.Style

	Bg lipgloss.Color

	TooNarrowTitle string
	TooNarrowNeed  string
	TooNarrowBody  string
}

func Default() Theme { return New(muninPalette) }

var current = func() *Theme { t := Default(); syncExported(t); return &t }()

func Cur() *Theme { return current }

func Use(t Theme) {
	current = &t
	syncExported(t)
}

func syncExported(t Theme) {
	TitleSty = t.Title
	AccentSty = t.Accent
	DimSty = t.Dim
	ValSty = t.Val
	KeySty = t.Key
	CanSty = t.Can
	CantSty = t.Cant

	AppFrame = t.AppFrame

	PanelSty = t.Panel
	PanelFocusSty = t.PanelFocus
	PanelTitleSty = t.PanelTitle
	CardSty = t.Card

	NotifPositiveSty = t.NotifPositive
	NotifNeutralSty = t.NotifNeutral
	NotifWarningSty = t.NotifWarning
	NotifNegativeSty = t.NotifNegative
	NotifIdleSty = t.NotifIdle
	NotifTitleSty = t.NotifTitle

	Series = t.Series
}
