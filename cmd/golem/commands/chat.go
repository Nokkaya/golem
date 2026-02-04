package commands

import "github.com/spf13/cobra"

func NewChatCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "chat",
        Short: "Chat with Golem",
        RunE: func(cmd *cobra.Command, args []string) error {
            // TODO: implement
            return nil
        },
    }
}
