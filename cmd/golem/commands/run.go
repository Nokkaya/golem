package commands

import "github.com/spf13/cobra"

func NewRunCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "run",
        Short: "Start Golem server (Telegram + scheduled tasks)",
        RunE: func(cmd *cobra.Command, args []string) error {
            // TODO: implement
            return nil
        },
    }
}
