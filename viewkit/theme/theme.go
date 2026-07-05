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

var (
	gold   = lipgloss.Color("220")
	dim    = lipgloss.Color("245")
	gray   = lipgloss.Color("252")
	gn     = lipgloss.Color("78")
	rd     = lipgloss.Color("203")
	purple = lipgloss.Color("141")
	orange = lipgloss.Color("208")

	TitleSty  = lipgloss.NewStyle().Bold(true).Foreground(gold)
	AccentSty = lipgloss.NewStyle().Bold(true).Foreground(gold)
	DimSty    = lipgloss.NewStyle().Foreground(dim)
	ValSty    = lipgloss.NewStyle().Foreground(gray)
	KeySty    = lipgloss.NewStyle().Bold(true).Foreground(gold)
	CanSty    = lipgloss.NewStyle().Foreground(gn)
	CantSty   = lipgloss.NewStyle().Foreground(rd)

	AppFrame = lipgloss.NewStyle().Margin(AppMarginY, AppMarginX)

	PanelSty      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(0, 1).Width(BodyWidth + 2)
	PanelFocusSty = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gold).Padding(0, 1).Width(BodyWidth + 2)
	PanelTitleSty = lipgloss.NewStyle().Bold(true).Foreground(gold)
	CardSty       = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gold).Padding(0, 1).Width(BodyWidth + 2).Align(lipgloss.Center)

	NotifPositiveSty = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gn).Foreground(gn).Padding(0, 1)
	NotifNeutralSty  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Foreground(purple).Padding(0, 1)
	NotifWarningSty  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(orange).Foreground(orange).Padding(0, 1)
	NotifNegativeSty = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(rd).Foreground(rd).Padding(0, 1)
	NotifIdleSty     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Foreground(dim).Padding(0, 1)
	NotifTitleSty    = lipgloss.NewStyle().Bold(true)

	Series = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(gold),
		lipgloss.NewStyle().Foreground(gn),
		lipgloss.NewStyle().Foreground(purple),
		lipgloss.NewStyle().Foreground(orange),
		lipgloss.NewStyle().Foreground(rd),
		lipgloss.NewStyle().Foreground(gray),
	}
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

	TooNarrowTitle string
	TooNarrowNeed  string
	TooNarrowBody  string
}

func Default() Theme {
	return Theme{
		Title:  TitleSty,
		Accent: AccentSty,
		Dim:    DimSty,
		Val:    ValSty,
		Key:    KeySty,
		Can:    CanSty,
		Cant:   CantSty,

		Panel:      PanelSty,
		PanelFocus: PanelFocusSty,
		PanelTitle: PanelTitleSty,
		Card:       CardSty,
		AppFrame:   AppFrame,

		NotifPositive: NotifPositiveSty,
		NotifNeutral:  NotifNeutralSty,
		NotifWarning:  NotifWarningSty,
		NotifNegative: NotifNegativeSty,
		NotifIdle:     NotifIdleSty,
		NotifTitle:    NotifTitleSty,

		Series: Series,

		TooNarrowTitle: DefaultTooNarrowTitle,
		TooNarrowNeed:  DefaultTooNarrowNeed,
		TooNarrowBody:  DefaultTooNarrowBody,
	}
}

var current = func() *Theme { t := Default(); return &t }()

func Cur() *Theme { return current }

func Use(t Theme) { current = &t }
