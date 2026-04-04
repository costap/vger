package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/costap/vger/internal/adapters/signals"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/costap/vger/internal/domain"
	"github.com/spf13/cobra"
)

var trackAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Capture a new technology signal",
	Long: `Interactively capture a new technology or idea to track.

You will be prompted for: title, URL, source, category, and a brief note
explaining why you captured this signal.

Example:
  vger track add`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sigDir, err := signals.DefaultDir()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		store := signals.New(sigDir)

		id, err := store.NextID(cmd.Context())
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.SectionHeader(fmt.Sprintf("new signal — %s", id))

		r := bufio.NewReader(os.Stdin)

		title := prompt(r, "Title")
		if title == "" {
			return fmt.Errorf("title is required")
		}

		url := prompt(r, "URL")
		source := prompt(r, "Source (e.g. Blog post, Twitter/X, Colleague)")
		if source == "" {
			source = "Unknown"
		}

		category := promptChoice(r, "Category", domain.ValidSignalCategories, "other")
		note := prompt(r, "Why captured (1-3 sentences)")

		now := time.Now().UTC()
		sig := &domain.Signal{
			ID:        id,
			Title:     title,
			Date:      now.Format("2006-01-02"),
			Source:    source,
			URL:       url,
			Category:  category,
			Status:    domain.SignalStatusSpotted,
			Note:      note,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := store.Save(cmd.Context(), sig); err != nil {
			ui.RedAlert(err)
			return err
		}

		fmt.Println()
		ui.Field("ID", sig.ID)
		ui.Field("Title", sig.Title)
		ui.Field("Status", sig.Status)
		ui.Field("Category", sig.Category)
		ui.Complete(fmt.Sprintf("signal %s captured", sig.ID))
		return nil
	},
}

// prompt reads a line from stdin with a labelled prompt.
func prompt(r *bufio.Reader, label string) string {
	fmt.Printf("  %s: ", ui.LabelStyle().Render(strings.ToUpper(label)))
	line, _ := r.ReadString('\n')
	return strings.TrimSpace(line)
}

// promptChoice reads a line and validates against allowed values,
// returning the defaultVal if the input is empty or invalid.
func promptChoice(r *bufio.Reader, label string, choices []string, defaultVal string) string {
	fmt.Printf("  %s [%s]: ", ui.LabelStyle().Render(strings.ToUpper(label)), strings.Join(choices, "|"))
	line, _ := r.ReadString('\n')
	val := strings.TrimSpace(line)
	for _, c := range choices {
		if val == c {
			return val
		}
	}
	return defaultVal
}
