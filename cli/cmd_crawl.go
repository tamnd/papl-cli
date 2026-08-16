package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newCrawlCmd(gf *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "crawl",
		Short: "Fetch each chapter page and store its content in the database",
		Long: `Download each queued chapter from papl.cs.brown.edu, convert the HTML
to Markdown, and store the result in the local SQLite database.

Run 'papl seed' first to populate the crawl queue.`,
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

			if state.PendingCount() == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "queue is empty -- run 'papl seed' first")
				return nil
			}

			client := papl.NewClient(cfg.Delay, cfg.Timeout)
			task := &papl.CrawlTask{Config: cfg, Client: client, DB: db, StateDB: state}

			fmt.Fprintln(cmd.OutOrStdout(), "crawling PAPL chapters...")
			m, err := task.Run(cmd.Context(), func(s *papl.CrawlState) {
				papl.PrintCrawlProgress(s)
			})
			fmt.Println()
			if err != nil {
				return fmt.Errorf("crawl: %w", err)
			}
			fmt.Printf("crawl done: %d fetched, %d failed in %s\n",
				m.Done, m.Failed, m.Duration.Round(time.Second))
			return nil
		},
	}
}
