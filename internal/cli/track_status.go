package cli

import (
	"fmt"
	"time"

	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var trackStatusCmd = &cobra.Command{
	Use:   "status <id> <new-status>",
	Short: "Update the status of a signal",
	Long: `Update the investigation status of a tracked signal.

Valid statuses: spotted → evaluating → adopted | rejected | parked

Examples:
  vger track status 0001 evaluating
  vger track status 0001 adopted`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := fmt.Sprintf("%04s", args[0])
		newStatus := args[1]

		valid := false
		for _, s := range domain.ValidSignalStatuses {
			if s == newStatus {
				valid = true
				break
			}
		}
		if !valid {
			err := fmt.Errorf("invalid status %q — valid values: %v", newStatus, domain.ValidSignalStatuses)
			ui.RedAlert(err)
			return err
		}

		sigDir, err := signals.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		store := signals.New(sigDir)

		sig, err := store.Load(cmd.Context(), id)
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		if sig == nil {
			err := fmt.Errorf("signal %s not found", id)
			ui.RedAlert(err)
			return err
		}

		oldStatus := sig.Status
		sig.Status = newStatus
		sig.UpdatedAt = time.Now().UTC()

		if err := store.Save(cmd.Context(), sig); err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Field("Signal", sig.ID+" — "+sig.Title)
		ui.Field("Status", fmt.Sprintf("%s → %s", oldStatus, newStatus))
		ui.Complete("status updated")
		return nil
	},
}
