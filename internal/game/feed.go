package game

import "fmt"

type feed struct {
	items []string
	last  string
	count int
}

func (f *feed) push(text string) {
	if text == "" {
		return
	}
	if len(f.items) > 0 && text == f.last {
		f.count++
		f.items[len(f.items)-1] = fmt.Sprintf("%s (x%d)", text, f.count)
		return
	}
	f.items = append(f.items, text)
	f.last = text
	f.count = 1
	if len(f.items) > feedHistory {
		f.items = f.items[len(f.items)-feedHistory:]
	}
}

func (f *feed) lines() []string {
	out := make([]string, len(f.items))
	copy(out, f.items)
	return out
}

func (f *feed) active() bool { return len(f.items) > 0 }

func (f *feed) size() int { return len(f.items) }
