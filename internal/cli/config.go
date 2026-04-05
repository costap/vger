package cli

import (
	"fmt"
	"strings"

	"github.com/costap/vger/internal/adapters/config"
	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage vger configuration",
	Long: `Manage persistent vger configuration stored at ~/.vger/config.yaml.

Configuration is automatically applied to all Gemini prompts (ask, research, digest, enrich)
so you don't need to repeat your stack context in every command.

Examples:
  vger config show
  vger config set user_context "We run AWS EKS 1.29 with Cilium CNI and Istio service mesh."
  vger config clear user_context`,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			ui.RedAlert(err)
			return err
		}
		if strings.TrimSpace(cfg.UserContext) == "" {
			fmt.Println(ui.DimStyle().Render("user_context: (not set)"))
		} else {
			ui.Field("user_context", "")
			for _, line := range strings.Split(strings.TrimSpace(cfg.UserContext), "\n") {
				fmt.Println("  " + line)
			}
		}
		configPath, _ := config.DefaultPath()
		fmt.Println(ui.DimStyle().Render("\nConfig file: " + configPath))
		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <field> <value>",
	Short: "Set a configuration field",
	Long: `Set a configuration field in ~/.vger/config.yaml.

Supported fields:
  user_context   Persistent context injected into all Gemini prompts.
                 Describe your stack, team, and goals so answers are tailored.

Example:
  vger config set user_context "We run AWS EKS 1.29, Go microservices, Cilium CNI."`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		field := strings.ToLower(args[0])
		value := args[1]

		cfg, err := config.Load()
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		switch field {
		case "user_context":
			cfg.UserContext = value
		default:
			err := fmt.Errorf("unknown config field %q — supported: user_context", field)
			ui.RedAlert(err)
			return err
		}

		if err := config.Save(cfg); err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Complete(fmt.Sprintf("%s updated", field))
		return nil
	},
}

var configClearCmd = &cobra.Command{
	Use:   "clear <field>",
	Short: "Clear a configuration field",
	Long: `Clear a configuration field, removing it from ~/.vger/config.yaml.

Example:
  vger config clear user_context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		field := strings.ToLower(args[0])

		cfg, err := config.Load()
		if err != nil {
			ui.RedAlert(err)
			return err
		}

		switch field {
		case "user_context":
			cfg.UserContext = ""
		default:
			err := fmt.Errorf("unknown config field %q — supported: user_context", field)
			ui.RedAlert(err)
			return err
		}

		if err := config.Save(cfg); err != nil {
			ui.RedAlert(err)
			return err
		}

		ui.Complete(fmt.Sprintf("%s cleared", field))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configClearCmd)
}
