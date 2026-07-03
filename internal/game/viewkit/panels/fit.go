package panels

// Tier is a discrete terminal-height band. Panels declare the smallest tier at
// which they appear, so each band shows an explicitly authored set rather than
// whatever a measure-and-drop pass happens to leave.
type Tier int

const (
	TierShort  Tier = iota // cramped: bare essentials only (zero value)
	TierMedium             // minimum-supported height: biggest panels drop
	TierTall               // spacious: every panel shows
)

type Section struct {
	Content string
	MinTier Tier // smallest tier at which this section appears; zero = always
}

// StackFit keeps the sections visible at tier (nested: short ⊆ medium ⊆ tall)
// and stacks them. Overflow within a tier is left to the viewport to clip.
func StackFit(tier Tier, sections ...Section) string {
	contents := make([]string, 0, len(sections))
	for _, s := range sections {
		if tier >= s.MinTier {
			contents = append(contents, s.Content)
		}
	}
	return Stack(contents...)
}
