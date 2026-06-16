package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newInfoCmd(gf *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show database and queue statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildConfig(gf)

			db, err := papl.OpenDB(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			stats, err := db.Stats()
			if err != nil {
				return fmt.Errorf("db stats: %w", err)
			}
			fmt.Printf("editions:  %d\n", stats.Editions)
			fmt.Printf("chapters:  %d\n", stats.Chapters)
			fmt.Printf("exercises: %d\n", stats.Exercises)
			fmt.Printf("db size:   %.1f MB\n", float64(stats.DBSize)/1e6)

			if _, err := os.Stat(cfg.StatePath); err == nil {
				state, err := papl.OpenState(cfg.StatePath)
				if err != nil {
					return fmt.Errorf("open state: %w", err)
				}
				defer state.Close()
				pending, inProg, done, failed := state.QueueStats()
				fmt.Printf("\nqueue pending:     %d\n", pending)
				fmt.Printf("queue in_progress: %d\n", inProg)
				fmt.Printf("queue done:        %d\n", done)
				fmt.Printf("queue failed:      %d\n", failed)
			}
			return nil
		},
	}
}
