package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/codyconfer/goose/internal/control"
	"github.com/codyconfer/goose/internal/events"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "Send an events command to a running game",
	Long: `Send an events command to a running game.

One-shot events (milestones like "press_darling" or "ipo_rumor") fire once and
are then remembered as fired. "fire" marks one as already fired so it won't
trigger; "unfire" clears that memory so it can trigger again.`,
}

func init() {
	eventsCmd.AddCommand(eventsFireCmd, eventsUnfireCmd)
	rootCmd.AddCommand(eventsCmd)
}

func dispatchEvents(label string, cmds ...events.Command) error {
	if err := control.Send(control.Message{Label: label, Events: cmds}); err != nil {
		return err
	}
	fmt.Println("→ sent to running game:", label)
	return nil
}

var eventsFireCmd = &cobra.Command{
	Use:   "fire <event-key>",
	Short: "Mark a one-shot event as already fired (suppress it)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return dispatchEvents(fmt.Sprintf("🔕 Suppressed %q", args[0]), events.Fire(args[0]))
	},
}

var eventsUnfireCmd = &cobra.Command{
	Use:   "unfire <event-key>",
	Short: "Clear a one-shot's fired flag so it can trigger again",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return dispatchEvents(fmt.Sprintf("🔔 Re-armed %q", args[0]), events.Unfire(args[0]))
	},
}
