package notify

import "testing"

func n(title string) Notification { return Notification{Title: title, Tone: ToneNeutral} }

func TestQueueEmpty(t *testing.T) {
	q := NewQueue(0)
	if q.Active() {
		t.Fatal("fresh queue should be inactive")
	}
	if _, ok := q.Current(); ok {
		t.Fatal("empty queue should have no current notification")
	}
	q.Beat()
}

func TestQueueShowsHeadUntilExpiry(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 2)

	for i := range 2 {
		cur, ok := q.Current()
		if !ok || cur.Title != "a" {
			t.Fatalf("beat %d: want head a, got %q ok=%v", i, cur.Title, ok)
		}
		q.Beat()
	}
	if q.Active() {
		t.Fatal("queue should drain after the head's TTL elapses")
	}
}

func TestQueueAdvancesThroughBacklog(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 1)
	q.Push(n("b"), 1)
	if q.Len() != 2 {
		t.Fatalf("want 2 queued, got %d", q.Len())
	}

	if cur, _ := q.Current(); cur.Title != "a" {
		t.Fatalf("want head a, got %q", cur.Title)
	}
	q.Beat()
	if cur, ok := q.Current(); !ok || cur.Title != "b" {
		t.Fatalf("want head b after a expires, got %q ok=%v", cur.Title, ok)
	}
	q.Beat()
	if q.Active() {
		t.Fatal("queue should be empty after both expire")
	}
}

func TestQueueDropsZeroTTL(t *testing.T) {
	q := NewQueue(0)
	q.Push(n("a"), 0)
	q.Push(n("b"), -5)
	if q.Active() {
		t.Fatal("non-positive TTL notifications should be dropped")
	}
}

func TestQueueCapDropsOldestPending(t *testing.T) {
	q := NewQueue(2)
	q.Push(n("head"), 10)
	q.Push(n("old"), 10)
	q.Push(n("new"), 10)

	if q.Len() != 2 {
		t.Fatalf("want 2 after cap, got %d", q.Len())
	}
	if cur, _ := q.Current(); cur.Title != "head" {
		t.Fatalf("cap must not evict the on-screen head, got %q", cur.Title)
	}
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	q.Beat()
	if cur, ok := q.Current(); !ok || cur.Title != "new" {
		t.Fatalf("want surviving pending to be new, got %q ok=%v", cur.Title, ok)
	}
}
