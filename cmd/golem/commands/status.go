package commands

import "github.com/spf13/cobra"

func NewStatusCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "status",
        Short: "Show Golem configuration status",
        RunE: func(cmd *cobra.Command, args []string) error {
            // TODO: implement
            return nil
        },
    }
}
