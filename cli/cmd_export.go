package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newExportCmd(gf *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "export",
		Short: "Write all chapters from the database to Markdown files",
		Long: `Read all chapters from the local SQLite database and write each one
to a Markdown file in the export directory.

Output: one .md file per chapter, named by slug (e.g., getting-started.md).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildConfig(gf)

			db, err := papl.OpenDB(cfg.DBPath)
			if err != nil {
				return fmt.Errorf("open db: %w", err)
			}
			defer db.Close()

			task := &papl.ExportTask{Config: cfg, DB: db}

			fmt.Fprintf(cmd.OutOrStdout(), "exporting to %s...\n", cfg.ExportDir)
			m, err := task.Run(cmd.Context(), func(s *papl.ExportState) {
				papl.PrintExportProgress(s)
			})
			fmt.Println()
			if err != nil {
				return fmt.Errorf("export: %w", err)
			}
			fmt.Printf("export done: %d files in %s\n", m.Files, m.Duration.Round(time.Millisecond))
			fmt.Printf("output dir: %s\n", cfg.ExportDir)
			return nil
		},
	}
}
