package game

import "github.com/codyconfer/viewkit/list"

func selectListKey(l *list.Model, key string) {
	for {
		it, ok := l.Selected()
		if !ok || it.Key == key {
			return
		}
		prev := it.Key
		l.Move(1)
		if cur, _ := l.Selected(); cur.Key == prev {
			return
		}
	}
}
