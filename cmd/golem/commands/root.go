package commands

import (
    "github.com/spf13/cobra"
)

// NewRootCmd creates the root command
func NewRootCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "golem",
        Short: "Golem - Lightweight AI Assistant",
        Long:  `Golem is a lightweight personal AI assistant built with Go and Eino.`,
    }

    cmd.AddCommand(
        NewInitCmd(),
        NewChatCmd(),
        NewRunCmd(),
        NewStatusCmd(),
    )

    return cmd
}
