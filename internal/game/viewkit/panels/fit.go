package panels

import "github.com/charmbracelet/lipgloss"

type Section struct {
	Content  string
	Priority int
}

const Essential = 0

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
