package notify

type Queue struct {
	items []queued
	cap   int
}

type queued struct {
	n   Notification
	ttl int
}

func NewQueue(cap int) *Queue { return &Queue{cap: cap} }

func (q *Queue) Push(n Notification, ttl int) {
	if ttl <= 0 {
		return
	}
	q.items = append(q.items, queued{n: n, ttl: ttl})
	if q.cap > 0 && len(q.items) > q.cap {
		q.items = append(q.items[:1], q.items[2:]...)
	}
}

func (q *Queue) Beat() {
	if len(q.items) == 0 {
		return
	}
	if q.items[0].ttl--; q.items[0].ttl <= 0 {
		q.items = q.items[1:]
	}
}

func (q *Queue) Current() (Notification, bool) {
	if len(q.items) == 0 {
		return Notification{}, false
	}
	return q.items[0].n, true
}

func (q *Queue) Active() bool { return len(q.items) > 0 }

func (q *Queue) Len() int { return len(q.items) }
