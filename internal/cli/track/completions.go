package track

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// signalIDCompletionFunc is a Cobra ValidArgsFunction for track sub-commands that
// accept a signal ID as the first positional argument. It reads the active signal
// store and returns "id\ttitle [status]" entries.
func signalIDCompletionFunc(cmd *cobra.Command, args []string, _ string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store, err := resolveSignalStore()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	ctx := context.Background()
	if cmd != nil && cmd.Context() != nil {
		ctx = cmd.Context()
	}

	sigs, err := store.LoadAll(ctx)
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	results := make([]string, 0, len(sigs))
	for _, s := range sigs {
		results = append(results, fmt.Sprintf("%s\t%s [%s]", s.ID, s.Title, s.Status))
	}
	return results, cobra.ShellCompDirectiveNoFileComp
}

// trackStatusCompletionFunc is a Cobra ValidArgsFunction for `vger track status`.
// Arg 0: signal ID — completed from the store.
// Arg 1: status value — completed from a static list.
func trackStatusCompletionFunc(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	switch len(args) {
	case 0:
		return signalIDCompletionFunc(cmd, args, toComplete)
	case 1:
		return []string{
			"evaluating\tCurrently being evaluated",
			"adopted\tDecided to adopt",
			"rejected\tDecided not to adopt",
			"monitoring\tKeeping an eye on it",
			"on-hold\tPaused evaluation",
			"spotted\tJust noticed, not yet evaluated",
		}, cobra.ShellCompDirectiveNoFileComp
	default:
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}
