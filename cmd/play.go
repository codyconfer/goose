package cmd

import "github.com/spf13/cobra"

var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Start the goose economy",
	RunE: func(cmd *cobra.Command, args []string) error {
		return play()
	},
}

func init() {
	rootCmd.AddCommand(playCmd)
}
