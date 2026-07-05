package game

import (
	"fmt"

	"github.com/codyconfer/viewkit/layout"
	"github.com/codyconfer/viewkit/notify"
	"github.com/codyconfer/viewkit/panels"

	"github.com/codyconfer/goose/internal/content"
	"github.com/codyconfer/goose/internal/economy"
)

var (
	specPremiums  = economy.SpecPremiums
	specLeverages = economy.SpecLeverages
)

func cellFrame(f layout.Frame) layout.Frame {
	inner := layout.NewFrame(f.Width - 4)
	if f.Focused {
		inner = inner.Focus()
	}
	return inner
}

func renderBook(m *Model, vk layout.Frame) string {
	s := m.econ.Get()
	data := []panels.Datum{
		{Label: content.Text.Spec.MixCash, Value: s.Tokens},
		{Label: content.Text.Spec.MixEggs, Value: s.Eggs * s.EggPrice()},
		{Label: content.Text.Spec.MixExposure, Value: s.LeveragedExposure()},
	}
	pieW := vk.Width
	if pieW > 48 {
		pieW = 48
	}
	return panels.Pie(vk, content.Text.Spec.MixPanel, data, pieW, economy.FormatNum, content.Text.Spec.MixEmpty)
}

func specWord(k economy.PosKind) string {
	if k == economy.PosPut {
		return content.Text.Spec.PutWord
	}
	return content.Text.Spec.CallWord
}

func absF(f float64) float64 {
	if f < 0 {
		return -f
	}
	return f
}

func positionSettleMsg(res economy.PosResult) string {
	if res.PnL >= 0 {
		return fmt.Sprintf(content.Text.Spec.SettledWinFmt, res.Pos.Desc(), economy.FormatNum(res.Payout))
	}
	return fmt.Sprintf(content.Text.Spec.SettledLossFmt, res.Pos.Desc(), economy.FormatNum(-res.PnL))
}

func marginCallNotif(res economy.PosResult) notify.Notification {
	msg := fmt.Sprintf("%s tripped maintenance and got liquidated. %s 🪙 came back; the rest went to the desk.", res.Pos.Desc(), economy.FormatNum(res.Payout))
	if res.Payout <= 0 {
		msg = fmt.Sprintf("%s tripped maintenance and got liquidated. The desk kept the whole premium.", res.Pos.Desc())
	}
	return notify.Notification{
		Title:   content.Text.Spec.MarginCallTitle,
		Message: msg,
		Tone:    notify.ToneNegative,
	}
}
