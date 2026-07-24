package game

import (
	"math/rand"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/codyconfer/viewkit/notify"

	"github.com/codyconfer/goose/internal/characters"
	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
	"github.com/codyconfer/goose/internal/events"
	"github.com/codyconfer/goose/internal/store"
	"github.com/codyconfer/goose/internal/world"
)

type Model struct {
	econ    *economy.Machine
	events  *events.Machine
	world   *world.State
	items   []capexItem
	rng     *rand.Rand
	clock   clock
	upBeats int

	width, height int
	pageScroll    int

	pulse      float64
	flash      string
	flashTTL   int
	offline    float64
	offlineTTL int
	feed       feed

	sellRate    float64
	priceAccum  float64
	candles     []candle
	candleBeats int

	vizBins  []float64
	vizPeaks []float64

	notifs *notify.Queue

	saveID   int64
	saveName string
	saves    []store.SaveInfo

	screen   screen
	quitting bool
}

func New(s *economy.Machine, ev *events.Machine, offline float64) Model {
	loadLayoutConfig()
	loadThemeConfig()
	loadClockConfig()
	m := Model{
		econ:   s,
		events: ev,
		world:  world.Generate(world.DefaultSeed),
		items:  capexItems(),
		clock:  newClock(time.Now()),
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		screen: &gameScreen{},
		notifs: notify.NewQueue(notifQueueCap),

		vizBins:  make([]float64, vizBands),
		vizPeaks: make([]float64, vizBands),
	}
	m.setOffline(offline)
	m.loadPriceChart()
	return m
}

func NewMenu() Model {
	m := New(economy.NewMachine(), events.NewMachine(), 0)
	saves, err := store.ListSaves()
	if err != nil {
		m.setFlash(content.Text.Menu.SaveError)
	}
	m.saves = saves
	m.screen = &menuScreen{items: menuItems(saves)}
	return m
}

func (m Model) Init() tea.Cmd { return upBeat() }

func (m *Model) setFlash(s string) {
	m.flash = s
	m.flashTTL = flashBeats
}

func (m *Model) setOffline(v float64) {
	m.offline = v
	if v > 0 {
		m.offlineTTL = offlineBeats
	} else {
		m.offlineTTL = 0
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.clampPageScroll()
		return m, nil

	case upBeatMsg:
		prev := m.screen
		dt := m.clock.beat(time.Time(msg))
		if !m.screen.simulates() {
			return m, upBeat()
		}
		m.upBeats++
		m.beatFast(dt)
		if m.upBeats%10 == 0 {
			m.beatSlow()
		}
		if m.upBeats%20 == 0 {
			m.beatChars()
		}
		if m.upBeats%4 == 0 {
			m.beatMid()
		}
		if cs, ok := m.screen.(*characterScreen); ok {
			cs.tick(&m)
		}
		m.syncPageScroll(prev)
		return m, upBeat()

	case tea.KeyMsg:
		if m.handlePageScroll(msg) {
			return m, nil
		}
		prev := m.screen
		cmd := m.screen.handleKey(&m, msg)
		m.syncPageScroll(prev)
		return m, cmd

	case ControlMsg:
		prev := m.screen
		m.applyControl(msg)
		m.syncPageScroll(prev)
		return m, nil
	}
	return m, nil
}

func (m *Model) syncPageScroll(prev screen) {
	if prev != m.screen {
		m.pageScroll = 0
		return
	}
	m.clampPageScroll()
}

func (m *Model) beatFast(dt float64) {
	if m.econ.TickFreeze(dt) {
		if dt > 0 {
			m.sellRate *= buyRateSmoothing
		}
	} else {
		m.econ.Produce(dt)

		m.econ.UpdateConsumers(dt)
		m.priceAccum += dt
		sold, _ := m.econ.RunMarket(dt)
		rep := m.econ.ProcessTransactions(dt)
		if dt > 0 {
			m.sellRate = m.sellRate*buyRateSmoothing + ((sold+rep.SoldEggs)/dt)*(1-buyRateSmoothing)
		}
		for _, o := range rep.Completed {
			m.feed.push(tradeCompletedMsg(o))
		}
		for _, res := range m.econ.TickPositions(dt) {
			if res.MarginCall {
				m.notifs.Push(marginCallNotif(res), notifBeats)
			} else {
				m.feed.push(positionSettleMsg(res))
			}
		}
	}

	if m.pulse > 0 {
		if m.pulse -= dt * pulseDecayRate; m.pulse < 0 {
			m.pulse = 0
		}
	}
	if m.flashTTL > 0 {
		if m.flashTTL--; m.flashTTL == 0 {
			m.flash = ""
		}
	}
	if m.offlineTTL > 0 {
		if m.offlineTTL--; m.offlineTTL == 0 {
			m.offline = 0
		}
	}
	m.updateViz(dt)
	m.notifs.Beat()
}

func (m *Model) beatMid() {
	if _, ok := m.screen.(*gameScreen); !ok {
		return
	}
	if m.econ.Get().Frozen() {
		return
	}

	if out, ok := m.events.Roll(m.world.Events, m.econ.Get(), m.rng); ok {
		m.econ.ApplyWindfall(out.Notif.Title, out.Cmds)
		m.notifs.Push(out.Notif, notifBeats)
	}
}

func (m *Model) beatSlow() {
	m.econ.BaselineYield()
	m.econ.UpdatePrice(m.priceAccum, m.rng)
	m.priceAccum = 0
	for _, ev := range m.econ.RunAgents() {
		m.feed.push(agentFiredMsg(ev))
	}
	m.recordPrice()
}

func (m *Model) beatChars() {
	gs, ok := m.screen.(*gameScreen)
	if !ok {
		return
	}
	if m.econ.Get().Frozen() {
		return
	}
	if ch, ok := characters.Roll(m.world.Characters, m.econ.Get(), m.rng); ok {
		m.screen = &characterScreen{char: &ch, prev: gs}
	}
}

func (m *Model) recordPrice() {
	p := m.econ.Get().EggPrice()
	if len(m.candles) == 0 || m.candleBeats >= candleSamples {
		m.candles = append(m.candles, candle{open: p, high: p, low: p, close: p})
		m.candleBeats = 1
		if len(m.candles) > priceChartHistory {
			m.candles = m.candles[len(m.candles)-priceChartHistory:]
		}
		m.syncPriceChart()
		return
	}
	c := &m.candles[len(m.candles)-1]
	c.close = p
	if p > c.high {
		c.high = p
	}
	if p < c.low {
		c.low = p
	}
	m.candleBeats++
	m.syncPriceChart()
}

func (m *Model) loadPriceChart() {
	stored, beats := m.econ.PriceChart()
	m.candles = make([]candle, 0, len(stored))
	for _, c := range stored {
		if c.Open <= 0 || c.High <= 0 || c.Low <= 0 || c.Close <= 0 {
			continue
		}
		m.candles = append(m.candles, candle{open: c.Open, high: c.High, low: c.Low, close: c.Close})
	}
	if len(m.candles) == 0 {
		price := m.econ.Get().EggPrice()
		m.candles = []candle{{open: price, high: price, low: price, close: price}}
	}
	if len(m.candles) > priceChartHistory {
		m.candles = m.candles[len(m.candles)-priceChartHistory:]
	}
	m.candleBeats = beats
	if m.candleBeats <= 0 {
		m.candleBeats = 1
	}
	if m.candleBeats > candleSamples {
		m.candleBeats = candleSamples
	}
	m.syncPriceChart()
}

func (m *Model) syncPriceChart() {
	stored := make([]economy.PriceCandle, len(m.candles))
	for i, c := range m.candles {
		stored[i] = economy.PriceCandle{Open: c.open, High: c.high, Low: c.low, Close: c.close}
	}
	m.econ.SetPriceChart(stored, m.candleBeats)
}

func (m *Model) save() error {
	m.syncPriceChart()
	if m.saveID <= 0 {
		info, err := store.CreateSave(m.nextSaveName(), m.econ, m.events, m.world)
		if err != nil {
			return err
		}
		m.saveID = info.ID
		m.saveName = info.Name
		return nil
	}
	return store.Save(m.saveID, m.econ, m.events, m.world)
}

func (m *Model) nextSaveName() string {
	saves, err := store.ListSaves()
	if err != nil {
		saves = m.saves
	}
	return store.NextSaveName(saves)
}

func (m *Model) refreshSaves(ms *menuScreen) {
	saves, err := store.ListSaves()
	if err != nil {
		m.setFlash(content.Text.Menu.SaveError)
		return
	}
	m.saves = saves
	ms.items = menuItems(saves)
	ms.ensure()
	sel := ""
	if it, ok := ms.list.Selected(); ok {
		sel = it.Key
	}
	ms.syncItems(sel)
}

func (m Model) nextVisible(from int) int {
	for i := from + 1; i < len(m.items); i++ {
		if m.unlocked(m.items[i]) {
			return i
		}
	}
	return from
}

func (m Model) prevVisible(from int) int {
	for i := from - 1; i >= 0; i-- {
		if m.unlocked(m.items[i]) {
			return i
		}
	}
	return from
}

func (m Model) unlocked(it capexItem) bool { return it.unlocked(m.econ.Get()) }

type upBeatMsg time.Time

func upBeat() tea.Cmd {
	return tea.Tick(upBeatRate, func(t time.Time) tea.Msg { return upBeatMsg(t) })
}

type clock struct {
	last time.Time
}

func newClock(now time.Time) clock {
	return clock{last: now}
}

func (c *clock) beat(now time.Time) float64 {
	dt := now.Sub(c.last).Seconds()
	c.last = now
	return dt
}
