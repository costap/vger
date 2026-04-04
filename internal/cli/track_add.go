package cli

import (
"bufio"
"fmt"
"os"
"os/exec"
"strings"
"time"

"github.com/costap/vger/internal/adapters/gemini"
"github.com/costap/vger/internal/adapters/signals"
"github.com/costap/vger/internal/cli/ui"
"github.com/costap/vger/internal/domain"
"github.com/spf13/cobra"
)

var trackAddAIPrompt string
var trackAddEdit bool

var trackAddCmd = &cobra.Command{
Use:   "add",
Short: "Capture a new technology signal",
Long: `Interactively capture a new technology or idea to track.

Without --ai: prompts for title, URL, source, category, and a brief note.
With --ai: describe the signal in natural language; Gemini extracts the fields.

When TECHDR_DIR is set (tech-signals repo), the signal is saved as a Markdown
file and auto-committed to git. Pass --edit to open $EDITOR for review.

Examples:
  vger track add
  vger track add --ai "saw a tweet about eBPF replacing sidecars https://..."
  vger track add --ai "..." --edit`,
RunE: func(cmd *cobra.Command, args []string) error {
store, err := resolveSignalStore()
if err != nil {
ui.RedAlert(err)
return err
}

id, err := store.NextID(cmd.Context())
if err != nil {
ui.RedAlert(err)
return err
}

now := time.Now().UTC()
var sig *domain.Signal

if trackAddAIPrompt != "" {
if geminiAPIKey == "" {
err := fmt.Errorf("GEMINI_API_KEY is required for --ai — set it as an env var or pass --gemini-key")
ui.RedAlert(err)
return err
}

gmClient := gemini.New(geminiAPIKey, geminiModel)

ui.SectionHeader(fmt.Sprintf("new signal — %s (ai)", id))
fmt.Println(ui.DimStyle().Render("  calling Gemini…"))

sig, err = gmClient.ParseSignalFromPrompt(cmd.Context(), trackAddAIPrompt)
if err != nil {
ui.RedAlert(err)
return err
}
} else {
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

sig = &domain.Signal{
Title:    title,
URL:      url,
Source:   source,
Category: category,
Note:     note,
}
}

sig.ID = id
sig.Date = now.Format("2006-01-02")
sig.Status = domain.SignalStatusSpotted
sig.CreatedAt = now
sig.UpdatedAt = now

if err := store.Save(cmd.Context(), sig); err != nil {
ui.RedAlert(err)
return err
}

// Bump .next-id counter for MarkdownStore (techdr convention).
if ms, ok := store.(*signals.MarkdownStore); ok {
_ = ms.BumpID()

// Open $EDITOR for review when --edit is set or defaulted for AI-assisted adds.
openEdit := trackAddEdit || (trackAddAIPrompt != "" && !cmd.Flags().Changed("edit"))
if openEdit {
openInEditor(ms.FilePath(id))
}
}

fmt.Println()
ui.Field("ID", sig.ID)
ui.Field("Title", sig.Title)
ui.Field("Source", sig.Source)
ui.Field("Category", sig.Category)
ui.Field("Status", sig.Status)
if len(sig.Tags) > 0 {
ui.Field("Tags", strings.Join(sig.Tags, ", "))
}
ui.Complete(fmt.Sprintf("signal %s captured", sig.ID))
return nil
},
}

func init() {
trackAddCmd.Flags().StringVar(&trackAddAIPrompt, "ai", "", "Describe the signal in natural language; Gemini extracts the fields")
trackAddCmd.Flags().BoolVar(&trackAddEdit, "edit", false, "Open $EDITOR after capture for review (default true with --ai when TECHDR_DIR is set)")
}

// openInEditor opens the given file path in $VISUAL or $EDITOR.
func openInEditor(path string) {
if path == "" {
return
}
editor := os.Getenv("VISUAL")
if editor == "" {
editor = os.Getenv("EDITOR")
}
if editor == "" {
return
}
cmd := exec.Command(editor, path)
cmd.Stdin = os.Stdin
cmd.Stdout = os.Stdout
cmd.Stderr = os.Stderr
_ = cmd.Run()
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
