package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newSeedCmd(gf *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "seed",
		Short: "Parse the table of contents and enqueue all chapter URLs",
		Long: `Fetch the PAPL table of contents page, discover all chapter URLs, and
enqueue any not yet visited. Run once before the first crawl. Re-running is
safe: already-visited chapters are skipped.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildConfig(gf)

			db, err := papl.OpenDB(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			state, err := papl.OpenState(cfg.StatePath)
			if err != nil {
				return fmt.Errorf("open state: %w", err)
			}
			defer state.Close()

			client := papl.NewClient(cfg.Delay, cfg.Timeout)
			task := &papl.SeedTask{Config: cfg, DB: db, StateDB: state, Client: client}

			fmt.Fprintln(cmd.OutOrStdout(), "seeding PAPL textbook...")
			m, err := task.Run(cmd.Context(), func(s *papl.SeedState) {
				papl.PrintSeedProgress(s)
			})
			fmt.Println()
			if err != nil {
				return fmt.Errorf("seed: %w", err)
			}
			fmt.Printf("seed done: %d pages, %d chapters, %d enqueued in %s\n",
				m.Pages, m.Chapters, m.Enqueued, m.Duration.Round(time.Millisecond))
			return nil
		},
	}
}
