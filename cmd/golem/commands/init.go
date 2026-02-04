package commands

import "github.com/spf13/cobra"

func NewInitCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "init",
        Short: "Initialize Golem configuration",
        RunE: func(cmd *cobra.Command, args []string) error {
            // TODO: implement
            return nil
        },
    }
}
