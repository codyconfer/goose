package panels

import (
	"strings"
	"testing"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
)

func TestNotificationToast(t *testing.T) {
	out := stripANSI(NotificationToast(layout.DefaultFrame(), notify.Positive("Saved", "all good")))
	for _, want := range []string{"✓", "Saved", "all good"} {
		if !strings.Contains(out, want) {
			t.Errorf("toast missing %q:\n%s", want, out)
		}
	}
}

func TestNotificationPanelEmpty(t *testing.T) {
	out := stripANSI(NotificationPanel(layout.DefaultFrame(), "ALERTS", nil))
	if !strings.Contains(out, "no notifications") {
		t.Fatalf("empty panel missing placeholder:\n%s", out)
	}
}

func TestNotificationPanelLists(t *testing.T) {
	ns := []notify.Notification{
		notify.Warning("Latency", "p99 climbing"),
		notify.Negative("Down", "region us-east"),
	}
	out := stripANSI(NotificationPanel(layout.DefaultFrame(), "ALERTS", ns))
	for _, want := range []string{"ALERTS", "Latency", "p99 climbing", "Down", "region us-east"} {
		if !strings.Contains(out, want) {
			t.Errorf("panel missing %q:\n%s", want, out)
		}
	}
}

func TestNotificationOverlayFloatsCard(t *testing.T) {
	bg := strings.TrimRight(strings.Repeat(strings.Repeat(".", 60)+"\n", 12), "\n")
	out := stripANSI(NotificationOverlay(bg, layout.NewFrame(30), notify.Neutral("Heads up", "something happened")))
	if !strings.Contains(out, "Heads up") {
		t.Fatalf("overlay missing card title:\n%s", out)
	}

	if !strings.Contains(out, ".") {
		t.Errorf("overlay swallowed the whole background:\n%s", out)
	}
}
