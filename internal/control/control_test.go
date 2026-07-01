package control

import (
	"testing"
	"time"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestSendReachesListener(t *testing.T) {
	t.Setenv("HOME", t.TempDir())

	got := make(chan Message, 1)
	srv, err := Listen(func(m Message) { got <- m })
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	defer srv.Close()

	want := Message{
		Label:  "test",
		Econ:   []economy.Command{economy.Earn(100), economy.GrantProducer("gpu", 2)},
		Events: []events.Command{events.Fire("press_darling")},
	}
	if err := Send(want); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case m := <-got:
		if m.Label != "test" || len(m.Econ) != 2 || len(m.Events) != 1 {
			t.Fatalf("garbled message: %+v", m)
		}
		if m.Econ[0].Effect != economy.EffectEarn || m.Econ[0].Amount != 100 {
			t.Fatalf("bad econ[0]: %+v", m.Econ[0])
		}
		if m.Events[0].Effect != events.EffectFire || m.Events[0].Key != "press_darling" {
			t.Fatalf("bad events[0]: %+v", m.Events[0])
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener never received the message")
	}
}

func TestSendNoListener(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if err := Send(Message{Label: "x"}); err != ErrNoGame {
		t.Fatalf("got %v, want ErrNoGame", err)
	}
}
