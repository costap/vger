package cli

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/adapters/quota"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

// tierLimits holds Gemini API rate limits for a specific model+tier combination.
// RPD of -1 means unlimited. Zero values mean unknown.
type tierLimits struct {
	RPD int // requests per day
	RPM int // requests per minute
	TPM int // tokens per minute
}

// modelTierLimits maps model prefix → billing tier → observed limits.
// Tier 1 values are sourced from AI Studio rate limits UI, April 2025 (screenshot-verified).
// Free tier RPD values match the prior knownRPD map.
var modelTierLimits = map[string]map[string]tierLimits{
	"gemini-2.5-flash": {
		"free": {RPD: 1_500, RPM: 10, TPM: 250_000},
		"1":    {RPD: 10_000, RPM: 1_000, TPM: 1_000_000},
	},
	"gemini-2.5-pro": {
		"free": {RPD: 100, RPM: 5, TPM: 250_000},
		"1":    {RPD: 1_000, RPM: 150, TPM: 2_000_000},
	},
	"gemini-1.5-flash": {
		"free": {RPD: 1_500, RPM: 15, TPM: 1_000_000},
	},
	"gemini-1.5-pro": {
		"free": {RPD: 50, RPM: 2, TPM: 32_000},
	},
}

// lookupModelLimits finds the limits for the given model+tier combination.
// Model matching is by prefix so variant suffixes (e.g. "-preview-…") are handled.
func lookupModelLimits(model, tier string) (tierLimits, bool) {
	if tier == "" {
		return tierLimits{}, false
	}
	for prefix, tiers := range modelTierLimits {
		if strings.HasPrefix(model, prefix) {
			if lim, ok := tiers[tier]; ok {
				return lim, true
			}
		}
	}
	return tierLimits{}, false
}

// modelRPD returns the RPD limit for the given model and tier.
// Returns -1 for unlimited, 0 for unknown.
func modelRPD(model, tier string) int {
	lim, ok := lookupModelLimits(model, tier)
	if !ok {
		return 0
	}
	return lim.RPD
}

// lookupModelTPM returns the tokens-per-minute limit for the given model and tier.
// Returns 0 if unknown (meaning no throttle should be applied).
func lookupModelTPM(model, tier string) int {
	lim, ok := lookupModelLimits(model, tier)
	if !ok {
		return 0
	}
	return lim.TPM
}

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Show today's Gemini API usage tracked by vger",
	Long: `Shows how many Gemini API requests vger has dispatched today (Pacific Time).
This is a local estimate based on vger's own activity — it does not account
for API calls made by other tools or sessions sharing the same API key/project.
For accurate usage, check https://aistudio.google.com (API keys section).

Configure your billing tier for limit estimates:
  vger config set quota_tier 1   # or: free, 2, 3`,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := quota.DefaultPath()
		if err != nil {
			return fmt.Errorf("locate quota file: %w", err)
		}
		tracker, err := quota.New(path)
		if err != nil {
			return fmt.Errorf("load quota tracker: %w", err)
		}

		today := tracker.Today()
		lim, tierKnown := lookupModelLimits(geminiModel, quotaTier)

		fmt.Println()
		ui.SectionHeader("Quota Status")
		ui.Field("model  ", geminiModel)

		if quotaTier != "" {
			tierLabel := quotaTier
			if quotaTier != "free" {
				tierLabel = "Paid Tier " + quotaTier
			} else {
				tierLabel = "Free"
			}
			ui.Field("tier   ", tierLabel)
		} else {
			ui.Field("tier   ", fmt.Sprintf("not configured — run: vger config set quota_tier 1  (or: free, 2, 3)"))
		}

		ui.Field("today  ", fmt.Sprintf("%d requests dispatched by vger (Pacific date)", today))

		if tierKnown {
			switch lim.RPD {
			case -1:
				ui.Field("RPD    ", "Unlimited / day")
			default:
				remaining := int64(lim.RPD) - today
				if remaining < 0 {
					remaining = 0
				}
				ui.Field("RPD    ", fmt.Sprintf("%s / day", formatInt(lim.RPD)))
				ui.Field("est. remaining", fmt.Sprintf("%d (vger tracking only)", remaining))
			}
			if lim.RPM > 0 {
				ui.Field("RPM    ", fmt.Sprintf("%s req/min", formatInt(lim.RPM)))
			}
			if lim.TPM > 0 {
				tpmNote := ""
				if strings.HasPrefix(geminiModel, "gemini-2.5-flash") || strings.HasPrefix(geminiModel, "gemini-2.5-pro") {
					tpmNote = "  ⚠  practical bottleneck for video analysis (~500K–1M tokens/video)"
				}
				ui.Field("TPM    ", fmt.Sprintf("%s tokens/min%s", formatInt(lim.TPM), tpmNote))
			}
		} else if quotaTier != "" {
			ui.Field("limits ", "not in known table — check https://aistudio.google.com")
		}

		fmt.Println()
		ui.Status("Resets at midnight Pacific Time (PT) — limits are per project, not per API key.")
		ui.Status("Other tools using the same project share this quota pool.")
		ui.Status("Full usage: https://aistudio.google.com (API keys → Rate limits)")
		fmt.Println()
		return nil
	},
}

// formatInt formats an integer with thousands separators for readability.
func formatInt(n int) string {
	s := fmt.Sprintf("%d", n)
	out := make([]byte, 0, len(s)+len(s)/3)
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, byte(c))
	}
	return string(out)
}
