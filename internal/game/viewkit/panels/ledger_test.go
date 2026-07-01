package panels

import (
	"strings"
	"testing"
)

func TestLedgerEmpty(t *testing.T) {
	out := Ledger("Ledger", nil, "🪙", fnum, 8, 0, "nothing yet")
	if !strings.Contains(out, "nothing yet") {
		t.Fatalf("empty ledger missing placeholder:\n%s", out)
	}
}

func TestLedgerDeltaSigns(t *testing.T) {
	rows := []LedgerRow{
		{Label: "bought eggs", Delta: 5},
		{Label: "sold eggs", Delta: -3},
		{Label: "no movement", Delta: 0},
	}
	out := Ledger("Ledger", rows, "🪙", fnum, 8, 0, "")
	for _, want := range []string{"bought eggs", "+5", "-3", "—"} {
		if !strings.Contains(out, want) {
			t.Errorf("ledger output missing %q:\n%s", want, out)
		}
	}
}

func TestLedgerScrollFooter(t *testing.T) {
	rows := make([]LedgerRow, 20)
	for i := range rows {
		rows[i] = LedgerRow{Label: "row", Delta: 1}
	}
	out := Ledger("Ledger", rows, "🪙", fnum, 8, 0, "")
	if !strings.Contains(out, "of 20") {
		t.Fatalf("long ledger missing scroll footer:\n%s", out)
	}
}
