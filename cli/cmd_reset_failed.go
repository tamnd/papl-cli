package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tamnd/papl-cli/papl"
)

func newResetFailedCmd(gf *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "reset-failed",
		Short: "Reset all failed queue items to pending for retry",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := buildConfig(gf)

			state, err := papl.OpenState(cfg.StatePath)
			if err != nil {
				return fmt.Errorf("open state: %w", err)
			}
			defer state.Close()

			n, err := state.ResetFailed()
			if err != nil {
				return fmt.Errorf("reset failed: %w", err)
			}
			fmt.Printf("reset %d failed items\n", n)
			return nil
		},
	}
}
