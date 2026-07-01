package game

import (
	"strings"
	"testing"
	"time"

	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
)

func TestTradeDeskOpensAndCloses(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m = send(m, key("t"))
	ts, ok := m.screen.(*tradeScreen)
	if !ok {
		t.Fatalf("pressing t opened %T, want *tradeScreen", m.screen)
	}
	if ts.prev == nil {
		t.Fatal("trade desk forgot the game screen it came from")
	}
	m = send(m, key("esc"))
	if _, ok := m.screen.(*gameScreen); !ok {
		t.Fatalf("esc returned to %T, want *gameScreen", m.screen)
	}
}

func TestTradeDeskSchedulesBuyOrder(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("t"))
	m = send(m, key("up"))
	m = send(m, key("enter"))
	if len(m.econ.Get().Transactions) != 1 {
		t.Fatalf("queue len=%d, want 1", len(m.econ.Get().Transactions))
	}
	o := m.econ.Get().Transactions[0]
	if o.Kind != economy.TxBuyEggs || o.Amount != 50 {
		t.Fatalf("queued %+v, want Buy 50", o)
	}
}

func TestTradeDeskTogglesToSell(t *testing.T) {
	s := economy.NewState()
	s.Eggs = 100
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("t"))
	m = send(m, key("right"))
	m = send(m, key("enter"))
	if len(m.econ.Get().Transactions) != 1 || m.econ.Get().Transactions[0].Kind != economy.TxSellEggs {
		t.Fatalf("expected a queued sell order, got %v", m.econ.Get().Transactions)
	}
}

func TestTradeDeskClearAndCancel(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	m := New(economy.FromState(s), events.NewMachine(), 0)
	m = send(m, key("t"))
	m = send(m, key("enter"))
	m = send(m, key("enter"))
	if len(m.econ.Get().Transactions) != 2 {
		t.Fatalf("queue len=%d, want 2", len(m.econ.Get().Transactions))
	}
	m = send(m, key("x"))
	if len(m.econ.Get().Transactions) != 1 {
		t.Fatalf("after cancel len=%d, want 1", len(m.econ.Get().Transactions))
	}
	m = send(m, key("c"))
	if len(m.econ.Get().Transactions) != 0 {
		t.Fatalf("after clear len=%d, want 0", len(m.econ.Get().Transactions))
	}
}

func TestQueueProcessesOnHeartbeat(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	s.Consumers = 100
	econ := economy.FromState(s)
	econ.ScheduleTrade(economy.TxBuyEggs, 20)
	m := New(econ, events.NewMachine(), 0)

	n, _ := m.Update(upBeatMsg(time.Now().Add(time.Second)))
	m = n.(Model)
	if len(m.econ.Get().Transactions) != 0 {
		t.Fatalf("order not worked off the queue: %d left", len(m.econ.Get().Transactions))
	}
	if m.econ.Get().Eggs < 20-1e-6 {
		t.Fatalf("bought %v eggs, want ~20", m.econ.Get().Eggs)
	}
}

func TestTradeDeskViewRenders(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	econ.ScheduleTrade(economy.TxBuyEggs, 100)
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	v := m.View()
	for _, want := range []string{"TRADE DESK", "NEW ORDER", "TRADE QUEUE", "Buy"} {
		if !strings.Contains(v, want) {
			t.Errorf("trade desk view missing %q", want)
		}
	}
}

func TestTradeDeskSimulates(t *testing.T) {
	if !(&tradeScreen{}).simulates() {
		t.Error("trade desk should keep the heartbeat running")
	}
}

func TestTradeDeskShowsLedger(t *testing.T) {
	s := economy.NewState()
	s.Tokens = 1000
	econ := economy.FromState(s)
	econ.BuyProducer(economy.Producers[0])
	m := New(econ, events.NewMachine(), 0)
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	v := m.View()
	if !strings.Contains(v, "LEDGER") {
		t.Error("trade desk missing the ledger panel")
	}
	if !strings.Contains(v, "Bought") {
		t.Error("ledger missing the producer purchase entry")
	}
}

func TestLedgerDescriptionsUseAssetIcons(t *testing.T) {
	producer := economy.Producers[0]
	producerRow := ledgerDesc(economy.Transaction{Kind: economy.TxBuyProducer, Label: producer.Name})
	if !strings.Contains(producerRow, producer.Icon) {
		t.Fatalf("producer ledger row %q missing icon %q", producerRow, producer.Icon)
	}

	upgrade := economy.Upgrades[0]
	upgradeRow := ledgerDesc(economy.Transaction{Kind: economy.TxUpgrade, Label: upgrade.Name})
	if !strings.Contains(upgradeRow, upgrade.Icon) {
		t.Fatalf("upgrade ledger row %q missing icon %q", upgradeRow, upgrade.Icon)
	}
}

func TestGameScreenDecommissions(t *testing.T) {
	s := economy.NewState()
	s.Owned["gpu"] = 2
	m := New(economy.FromState(s), events.NewMachine(), 0)

	m.screen.(*gameScreen).cursor = 3
	before := m.econ.Get().Tokens
	m = send(m, key("s"))
	if m.econ.Get().Count("gpu") != 1 {
		t.Fatalf("gpu count=%d, want 1 after decommission", m.econ.Get().Count("gpu"))
	}
	if m.econ.Get().Tokens <= before {
		t.Fatal("decommission did not refund tokens")
	}
	if len(m.econ.Get().Ledger) == 0 {
		t.Fatal("decommission not recorded in the ledger")
	}
}

func TestPriceChartInView(t *testing.T) {
	s := economy.NewState()
	s.Owned["server"] = 3
	m := New(economy.FromState(s), events.NewMachine(), 0)
	for i := 0; i < 30; i++ {
		m.priceAccum = 3.0
		m.beatSlow()
	}
	if len(m.candles) < 2 {
		t.Fatalf("expected the chart to accumulate candles, got %d", len(m.candles))
	}
	m.screen = &tradeScreen{prev: &gameScreen{}, kind: economy.TxBuyEggs}
	if v := m.View(); !strings.Contains(v, "EGG PRICE") {
		t.Error("trade desk view missing the egg price chart")
	}
}

func TestModelLoadsPersistedPriceChart(t *testing.T) {
	s := economy.NewState()
	s.PriceCandles = []economy.PriceCandle{{Open: 1, High: 3, Low: 0.5, Close: 2}}
	s.PriceCandleBeats = 2

	m := New(economy.FromState(s), events.NewMachine(), 0)
	if len(m.candles) != 1 {
		t.Fatalf("candles len=%d, want 1", len(m.candles))
	}
	if got := m.candles[0]; got.open != 1 || got.high != 3 || got.low != 0.5 || got.close != 2 {
		t.Fatalf("loaded candle=%+v", got)
	}
	if m.candleBeats != 2 {
		t.Fatalf("candle beats=%d, want 2", m.candleBeats)
	}
}

func TestRecordPricePersistsChartData(t *testing.T) {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	m.candles = []candle{{open: 1, high: 1, low: 1, close: 1}}
	m.candleBeats = candleSamples

	m.recordPrice()

	stored, beats := m.econ.PriceChart()
	if len(stored) != 2 {
		t.Fatalf("stored candles len=%d, want 2", len(stored))
	}
	if beats != 1 {
		t.Fatalf("stored candle beats=%d, want 1", beats)
	}
}
