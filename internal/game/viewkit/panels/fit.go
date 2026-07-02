package panels

import "github.com/charmbracelet/lipgloss"

// Section is a stackable panel tagged with a drop priority.
//
// Priority Essential (0) is never dropped. Sections with a higher Priority are
// dropped earlier when the stack does not fit its row budget, so decorative
// panels should carry the largest numbers and load-bearing ones the smallest.
type Section struct {
	Content  string
	Priority int
}

// Essential marks a section that StackFit must never drop.
const Essential = 0

// StackFit stacks sections like Stack, but while the rendered height exceeds
// budget rows it drops the highest-Priority non-essential section (ties broken
// by dropping the bottom-most first, keeping the top anchored) until it fits or
// only essential sections remain. A budget <= 0 means "no limit": every section
// is kept and the result is identical to Stack.
func StackFit(budget int, sections ...Section) string {
	kept := sections
	if budget > 0 {
		kept = fitSections(sections, budget)
	}
	contents := make([]string, len(kept))
	for i, s := range kept {
		contents[i] = s.Content
	}
	return Stack(contents...)
}

// fitSections returns the sections to keep, dropping droppable ones until the
// stack fits budget or nothing droppable remains.
func fitSections(sections []Section, budget int) []Section {
	kept := make([]Section, len(sections))
	copy(kept, sections)

	for stackHeight(kept) > budget {
		victim := droppableVictim(kept)
		if victim < 0 {
			break
		}
		kept = append(kept[:victim], kept[victim+1:]...)
	}
	return kept
}

// droppableVictim returns the index of the next section to drop: the one with
// the highest Priority, breaking ties toward the bottom-most section. Returns
// -1 when only essential sections remain.
func droppableVictim(sections []Section) int {
	victim := -1
	for i, s := range sections {
		if s.Priority == Essential {
			continue
		}
		if victim < 0 || s.Priority >= sections[victim].Priority {
			victim = i
		}
	}
	return victim
}

// stackHeight reports the row count Stack would produce for these sections:
// the sum of the heights of the non-empty sections plus one blank separator
// between each adjacent pair.
func stackHeight(sections []Section) int {
	total, nonEmpty := 0, 0
	for _, s := range sections {
		if s.Content == "" {
			continue
		}
		total += lipgloss.Height(s.Content)
		nonEmpty++
	}
	if nonEmpty > 1 {
		total += nonEmpty - 1
	}
	return total
}
