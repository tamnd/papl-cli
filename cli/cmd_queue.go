package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newQueueCmd(gf *globalFlags) *cobra.Command {
	var status string
	var limit int

	cmd := &cobra.Command{
		Use:   "queue",
		Short: "List crawl-queue items by status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildConfig(gf)

			state, err := papl.OpenState(cfg.StatePath)
			if err != nil {
				return fmt.Errorf("open state: %w", err)
			}
			defer state.Close()

			items, err := state.ListQueue(status, limit)
			if err != nil {
				return fmt.Errorf("list queue: %w", err)
			}
			if len(items) == 0 {
				fmt.Printf("no %s items in queue\n", status)
				return nil
			}
			fmt.Printf("%-12s %s\n", "STATUS", "URL")
			for _, it := range items {
				fmt.Printf("%-12s %s\n", status, it.URL)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "pending", "Filter by status: pending|done|failed|in_progress")
	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum items to list (0 = no limit)")
	return cmd
}
