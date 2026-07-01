package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/codyconfer/goose/internal/control"
	"github.com/codyconfer/goose/internal/economy"
)

var economyCmd = &cobra.Command{
	Use:   "economy",
	Short: "Send an economy command to a running game",
	Long: `Send an economy command to a running game.

Each subcommand fires one mutation at the live flock over the control socket and
returns immediately — run it from a second terminal while "goose play" is open
in another and watch the banner appear there.`,
}

func init() {
	economyCmd.AddCommand(
		economyEarnCmd,
		economySpendCmd,
		economyGrantCmd,
		economyGrowCrowdCmd,
		economyAddConsumersCmd,
		economyTradeCmd,
		economySeizeCmd,
	)
	rootCmd.AddCommand(economyCmd)
}

func dispatchEcon(label string, cmds ...economy.Command) error {
	if err := control.Send(control.Message{Label: label, Econ: cmds}); err != nil {
		return err
	}
	fmt.Println("→ sent to running game:", label)
	return nil
}

func parseFloat(name, raw string) (float64, error) {
	n, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: must be a number", name, raw)
	}
	return n, nil
}

var economyEarnCmd = &cobra.Command{
	Use:   "earn <tokens>",
	Short: "Credit tokens to the flock",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseFloat("amount", args[0])
		if err != nil {
			return err
		}
		return dispatchEcon(fmt.Sprintf("💸 Granted %s tokens", economy.FormatNum(n)), economy.Earn(n))
	},
}

var economySpendCmd = &cobra.Command{
	Use:   "spend <tokens>",
	Short: "Debit tokens from the flock (clamped at zero)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseFloat("amount", args[0])
		if err != nil {
			return err
		}
		return dispatchEcon(fmt.Sprintf("🧾 Charged %s tokens", economy.FormatNum(n)), economy.Spend(n))
	},
}

var economyGrantCmd = &cobra.Command{
	Use:   "grant <producer-key> <count>",
	Short: "Grant producers for free (keys: gpu, server, rack, datacenter, hyperscaler, starcloud)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		count, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid count %q: must be a whole number", args[1])
		}
		return dispatchEcon(fmt.Sprintf("🎁 Granted %d × %s", count, key), economy.GrantProducer(key, count))
	},
}

var economyGrowCrowdCmd = &cobra.Command{
	Use:   "grow-crowd <factor>",
	Short: "Scale the consumer crowd by a multiplier (>1 swells, <1 thins)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := parseFloat("factor", args[0])
		if err != nil {
			return err
		}
		return dispatchEcon(fmt.Sprintf("📣 Crowd × %.2f", f), economy.GrowCrowd(f))
	},
}

var economyAddConsumersCmd = &cobra.Command{
	Use:   "add-consumers <count>",
	Short: "Add a flat number of consumers (negative removes)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		n, err := parseFloat("count", args[0])
		if err != nil {
			return err
		}
		return dispatchEcon(fmt.Sprintf("👥 %+.0f consumers", n), economy.AddConsumers(n))
	},
}

var economyTradeCmd = &cobra.Command{
	Use:   "trade <buy|sell> <eggs> <price>",
	Short: "Move eggs across the market at a given price",
	Args:  cobra.ExactArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		var kind economy.TxKind
		switch args[0] {
		case "buy":
			kind = economy.TxBuyEggs
		case "sell":
			kind = economy.TxSellEggs
		default:
			return fmt.Errorf("invalid direction %q: want \"buy\" or \"sell\"", args[0])
		}
		eggs, err := parseFloat("eggs", args[1])
		if err != nil {
			return err
		}
		price, err := parseFloat("price", args[2])
		if err != nil {
			return err
		}
		return dispatchEcon(
			fmt.Sprintf("🛒 %s %s eggs @ %s", args[0], economy.FormatNum(eggs), economy.FormatNum(price)),
			economy.Trade(kind, eggs, price),
		)
	},
}

var economySeizeCmd = &cobra.Command{
	Use:   "seize",
	Short: "Repossess one unit of the most-owned producer",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return dispatchEcon("🦈 Seized collateral", economy.SeizeBest())
	},
}
