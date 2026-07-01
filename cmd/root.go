package cmd

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/codyconfer/goose/internal/buildinfo"
	"github.com/codyconfer/goose/internal/control"
	"github.com/codyconfer/goose/internal/game"
)

var rootCmd = &cobra.Command{
	Use:     "goose",
	Short:   "An idle-clicker economy of tokens and Goose Premium",
	Version: buildinfo.String(),
	Long: `goose is a tiny terminal economy game.

You run a flock of geese. Press enter to generate tokens, then spend those
tokens on GPUs, servers, data centers and clouds that generate Goose Premium for you.
Sell Goose Premium for more tokens — tokens are the only currency there is.

Run "goose" (or "goose play") to start generating.

While a game is running you can poke the live flock from another terminal with
"goose economy ..." and "goose events ..." — those commands land on the running
window without blocking.`,

	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return play()
	},
}

func play() error {
	p := tea.NewProgram(game.NewMenu(), tea.WithAltScreen())
	if srv, err := control.Listen(func(msg control.Message) {
		p.Send(game.ControlMsg(msg))
	}); err == nil {
		defer func() { _ = srv.Close() }()
	}
	_, err := p.Run()
	return err
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
