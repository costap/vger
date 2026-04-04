package cli

import "fmt"

// Lens is a named analytical preset for the ask command. It injects a role
// context and a default question so the user doesn't need to retype the same
// verbose prompt every time they want a standard type of analysis.
type Lens struct {
	Name            string
	ShortDesc       string
	RoleContext     string
	DefaultQuestion string
}

// builtinLenses is the ordered list of lenses available via --lens.
var builtinLenses = []Lens{
	{
		Name:      "architect",
		ShortDesc: "Solutions architect analysis: approach, decisions, trade-offs, novelty",
		RoleContext: "Act as a solutions architect reviewing this conference talk. " +
			"Use precise technical language appropriate for architects.",
		DefaultQuestion: "Analyse the technical approach and architectural decisions made in this talk. " +
			"What problem does it solve and how? " +
			"What are the novelty points compared to conventional approaches? " +
			"What trade-offs or risks were acknowledged? " +
			"What should a solutions architect take away from this?",
	},
	{
		Name:      "engineer",
		ShortDesc: "Hands-on deep-dive: implementation patterns, config details, getting started",
		RoleContext: "Act as a senior software engineer evaluating this talk for hands-on adoption. " +
			"Focus on practical, actionable detail.",
		DefaultQuestion: "What are the key implementation patterns, configuration details, or code examples shown? " +
			"What would an engineer need to know to get started with the main technology discussed? " +
			"What pitfalls, gotchas, or operational concerns were mentioned? " +
			"What specific versions, APIs, or tools were called out?",
	},
	{
		Name:      "radar",
		ShortDesc: "Tech radar: Adopt / Trial / Assess / Hold recommendation per technology",
		RoleContext: "Act as a technology radar analyst assessing technologies for enterprise adoption.",
		DefaultQuestion: "For each significant technology mentioned in this talk, provide a technology radar " +
			"recommendation: Adopt, Trial, Assess, or Hold. " +
			"For each one give a one-sentence justification based on signals from the talk: " +
			"maturity, production usage, community, operational complexity, and vendor risk.",
	},
	{
		Name:      "brief",
		ShortDesc: "Team briefing: 3–5 bullet summary of problem, approach, takeaways, action",
		RoleContext: "Act as a technical lead writing a concise briefing for your engineering team.",
		DefaultQuestion: "Summarise this talk in 3–5 bullet points covering: " +
			"what problem it addresses, the approach taken, the key technical takeaways, " +
			"and whether the team should evaluate any of the technologies discussed — and why.",
	},
}

// lookupLens returns the Lens with the given name (case-insensitive).
// Returns the zero Lens and false if not found.
func lookupLens(name string) (Lens, bool) {
	for _, l := range builtinLenses {
		if l.Name == name {
			return l, true
		}
	}
	return Lens{}, false
}

// lensNames returns a comma-separated list of available lens names.
func lensNames() string {
	names := make([]string, len(builtinLenses))
	for i, l := range builtinLenses {
		names[i] = l.Name
	}
	out := ""
	for i, n := range names {
		if i > 0 {
			out += ", "
		}
		out += n
	}
	return out
}

// buildLensPrompt merges the lens role context with the user question.
// If userQuestion is empty, the lens default question is used.
func buildLensPrompt(lens Lens, userQuestion string) string {
	q := userQuestion
	if q == "" {
		q = lens.DefaultQuestion
	}
	return fmt.Sprintf("%s\n\n%s", lens.RoleContext, q)
}
