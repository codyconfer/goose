package theme

import "github.com/charmbracelet/lipgloss"

const (
	BodyWidth    = 81
	MinBodyWidth = 24

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

	TitleSty = lipgloss.NewStyle().Bold(true).Foreground(gold)
	EggSty   = lipgloss.NewStyle().Bold(true).Foreground(gold)
	DimSty   = lipgloss.NewStyle().Foreground(dim)
	ValSty   = lipgloss.NewStyle().Foreground(gray)
	KeySty   = lipgloss.NewStyle().Bold(true).Foreground(gold)
	CanSty   = lipgloss.NewStyle().Foreground(gn)
	CantSty  = lipgloss.NewStyle().Foreground(rd)

	AppFrame = lipgloss.NewStyle().Margin(1, 2)

	PanelSty      = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(0, 1).Width(BodyWidth + 2)
	PanelTitleSty = lipgloss.NewStyle().Bold(true).Foreground(gold)
	TapCardSty    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gold).Padding(0, 1).Width(BodyWidth + 2).Align(lipgloss.Center)

	NotifPositiveSty = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(gn).Foreground(gn).Padding(0, 1)
	NotifNeutralSty  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Foreground(purple).Padding(0, 1)
	NotifWarningSty  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(orange).Foreground(orange).Padding(0, 1)
	NotifNegativeSty = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(rd).Foreground(rd).Padding(0, 1)
	NotifIdleSty     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Foreground(dim).Padding(0, 1)

	Series = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(gold),
		lipgloss.NewStyle().Foreground(gn),
		lipgloss.NewStyle().Foreground(purple),
		lipgloss.NewStyle().Foreground(orange),
		lipgloss.NewStyle().Foreground(rd),
		lipgloss.NewStyle().Foreground(gray),
	}
)
