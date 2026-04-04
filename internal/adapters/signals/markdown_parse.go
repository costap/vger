package signals

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/costap/vger/internal/domain"
)

// ── Parse ─────────────────────────────────────────────────────────────────────

// parseSignalFromMarkdown reads a TechDR Markdown file and converts it to a Signal.
// Fields not present in the Markdown (CreatedAt, UpdatedAt) are estimated from the
// file's modification time.
func parseSignalFromMarkdown(path string) (*domain.Signal, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	info, _ := os.Stat(path)
	modTime := time.Now().UTC()
	if info != nil {
		modTime = info.ModTime().UTC()
	}

	base := filepath.Base(path)
	id := ""
	if len(base) >= 4 {
		id = base[:4]
	}

	lines := strings.Split(string(data), "\n")
	sig := &domain.Signal{
		ID:        id,
		CreatedAt: modTime,
		UpdatedAt: modTime,
	}

	// Split into header (before AI ENRICHMENT) and enrichment sections.
	headerLines, enrichLines := splitAtEnrichmentMarker(lines)

	parseHeader(headerLines, sig)
	parseEnrichment(enrichLines, sig)

	return sig, nil
}

// splitAtEnrichmentMarker splits lines at the AI ENRICHMENT comment marker.
func splitAtEnrichmentMarker(lines []string) (header, enrich []string) {
	for i, l := range lines {
		if strings.Contains(l, "AI ENRICHMENT") {
			return lines[:i], lines[i+1:]
		}
	}
	return lines, nil
}

// parseHeader extracts Signal fields from the header section of a TechDR file.
func parseHeader(lines []string, sig *domain.Signal) {
	var noteLines []string
	inNote := false

	for _, raw := range lines {
		l := strings.TrimRight(raw, " \r")

		// Title line: # NNNN — Title
		if strings.HasPrefix(l, "# ") {
			sig.Title = extractTitleFromHeading(l)
			inNote = false
			continue
		}

		// Bold key: value fields.
		if strings.HasPrefix(l, "**") {
			inNote = false
			key, val := parseBoldField(l)
			switch strings.ToLower(key) {
			case "date":
				sig.Date = strings.TrimSpace(val)
			case "source":
				sig.Source = strings.TrimSpace(val)
			case "url":
				sig.URL = strings.TrimSpace(val)
			case "category":
				sig.Category = strings.TrimSpace(val)
			case "status":
				sig.Status = strings.TrimSpace(val)
			case "tags":
				sig.Tags = splitTags(val)
			}
			continue
		}

		// Section headings.
		if strings.HasPrefix(l, "## ") {
			heading := strings.TrimPrefix(l, "## ")
			if isNoteHeading(heading) {
				inNote = true
				noteLines = nil
			} else if strings.ToLower(heading) == "links" {
				inNote = false
			} else {
				inNote = false
			}
			continue
		}

		// Links section — look for vger-video: entries.
		if strings.HasPrefix(l, "- vger-video:") {
			vid := strings.TrimSpace(strings.TrimPrefix(l, "- vger-video:"))
			if vid != "" {
				sig.LinkedVideoIDs = append(sig.LinkedVideoIDs, vid)
			}
			continue
		}

		// Note body.
		if inNote && l != "" && !strings.HasPrefix(l, "---") {
			noteLines = append(noteLines, l)
		}
	}

	sig.Note = strings.TrimSpace(strings.Join(noteLines, "\n"))
}

// parseEnrichment extracts SignalEnrichment from the AI ENRICHMENT section.
func parseEnrichment(lines []string, sig *domain.Signal) {
	if len(lines) == 0 {
		return
	}

	sections := extractSections(lines)

	whatItIs := sections["what it is"]
	maturity := sections["maturity & risk"]
	alternativesRaw := sections["alternatives"]
	stackFit := sections["how it could fit our stack"]
	nextStepsRaw := sections["suggested next steps"]

	// Only attach enrichment if at least one section is non-empty.
	if whatItIs == "" && maturity == "" && alternativesRaw == "" && stackFit == "" && nextStepsRaw == "" {
		return
	}

	sig.Enrichment = &domain.SignalEnrichment{
		WhatItIs:     whatItIs,
		Maturity:     maturity,
		Alternatives: parseBulletList(alternativesRaw),
		StackFit:     stackFit,
		NextSteps:    parseCheckboxList(nextStepsRaw),
	}
}

// extractSections returns a map of heading (lowercased) → body text from a set
// of Markdown lines.
func extractSections(lines []string) map[string]string {
	result := make(map[string]string)
	current := ""
	var body []string

	flush := func() {
		if current != "" {
			result[current] = strings.TrimSpace(strings.Join(body, "\n"))
		}
	}

	for _, raw := range lines {
		l := strings.TrimRight(raw, " \r")
		if strings.HasPrefix(l, "## ") {
			flush()
			current = strings.ToLower(strings.TrimPrefix(l, "## "))
			body = nil
			continue
		}
		if strings.HasPrefix(l, "---") {
			continue
		}
		if current != "" {
			body = append(body, l)
		}
	}
	flush()

	return result
}

// ── Render ────────────────────────────────────────────────────────────────────

const techDRTemplate = `# %s — %s

**Date:** %s
**Source:** %s
**URL:** %s
**Category:** %s
**Status:** %s%s

## Why I captured this

%s

## Links

%s---
<!-- ── AI ENRICHMENT ─────────────────────────────────────────────────────── -->
<!-- Run: vger track enrich %s  to populate the sections below via Gemini    -->

## What it is

%s
## Maturity & Risk

%s
## Alternatives

%s
## How it could fit our stack

%s
## Suggested next steps

%s`

// renderSignalToMarkdown converts a Signal to TechDR Markdown.
func renderSignalToMarkdown(sig *domain.Signal) string {
	tagsLine := ""
	if len(sig.Tags) > 0 {
		tagsLine = "\n**Tags:** " + strings.Join(sig.Tags, ", ")
	}

	var linksBuilder strings.Builder
	if sig.URL != "" {
		linksBuilder.WriteString("- " + sig.URL + "\n")
	}
	for _, vid := range sig.LinkedVideoIDs {
		linksBuilder.WriteString("- vger-video: " + vid + "\n")
	}
	links := linksBuilder.String()
	if links == "" {
		links = "- \n"
	}

	whatItIs, maturity, alternativesStr, stackFit, nextStepsStr := "", "", "- \n", "", "- [ ] \n"
	if sig.Enrichment != nil {
		e := sig.Enrichment
		whatItIs = e.WhatItIs
		maturity = e.Maturity
		stackFit = e.StackFit
		if len(e.Alternatives) > 0 {
			var sb strings.Builder
			for _, a := range e.Alternatives {
				sb.WriteString("- " + a + "\n")
			}
			alternativesStr = sb.String()
		}
		if len(e.NextSteps) > 0 {
			var sb strings.Builder
			for _, s := range e.NextSteps {
				s = strings.TrimPrefix(s, "[ ] ")
				s = strings.TrimPrefix(s, "[x] ")
				sb.WriteString("- [ ] " + s + "\n")
			}
			nextStepsStr = sb.String()
		}
	}

	url := sig.URL
	if url == "" {
		url = "<primary link>"
	}
	source := sig.Source
	if source == "" {
		source = "Unknown"
	}
	status := sig.Status
	if status == "" {
		status = domain.SignalStatusSpotted
	}
	category := sig.Category
	if category == "" {
		category = "other"
	}

	return fmt.Sprintf(techDRTemplate,
		sig.ID, sig.Title,
		sig.Date,
		source,
		url,
		category,
		status,
		tagsLine,
		sig.Note,
		links,
		sig.ID,
		whatItIs,
		maturity,
		alternativesStr,
		stackFit,
		nextStepsStr,
	)
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// extractTitleFromHeading parses "# NNNN — Title" → "Title".
// Also handles "# NNNN" and "# Title" forms.
func extractTitleFromHeading(line string) string {
	s := strings.TrimPrefix(line, "# ")
	// Remove ID prefix "NNNN — " or "NNNN – "
	re := regexp.MustCompile(`^\d{4}\s*[—–-]+\s*`)
	return strings.TrimSpace(re.ReplaceAllString(s, ""))
}

// parseBoldField extracts key and value from "**Key:** value" lines.
func parseBoldField(line string) (key, val string) {
	// Strip leading **
	s := strings.TrimPrefix(line, "**")
	// Find closing **:
	idx := strings.Index(s, ":**")
	if idx < 0 {
		return "", ""
	}
	key = s[:idx]
	val = strings.TrimSpace(s[idx+3:]) // skip :** and space
	return
}

// isNoteHeading returns true for the "Why I captured this" heading variants.
func isNoteHeading(heading string) bool {
	h := strings.ToLower(heading)
	return strings.Contains(h, "why") || strings.Contains(h, "captured")
}

// splitTags splits a comma-separated tag string.
func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	var out []string
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// parseBulletList extracts items from a Markdown bullet list body.
func parseBulletList(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		if strings.HasPrefix(l, "- ") {
			item := strings.TrimPrefix(l, "- ")
			if item != "" && item != "<primary link>" {
				out = append(out, item)
			}
		}
	}
	return out
}

// parseCheckboxList extracts step text from "- [ ] step" lines.
func parseCheckboxList(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		l := strings.TrimSpace(line)
		l = strings.TrimPrefix(l, "- [ ] ")
		l = strings.TrimPrefix(l, "- [x] ")
		l = strings.TrimPrefix(l, "- [X] ")
		if l != "" && l != "<primary link>" {
			out = append(out, l)
		}
	}
	return out
}

// slugify converts a title to a kebab-case slug safe for use in file names.
func slugify(title string) string {
	var sb strings.Builder
	prevDash := false
	for _, r := range strings.ToLower(title) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
			prevDash = false
		} else if !prevDash && sb.Len() > 0 {
			sb.WriteRune('-')
			prevDash = true
		}
	}
	s := strings.TrimRight(sb.String(), "-")
	if len(s) > 60 {
		s = s[:60]
		s = strings.TrimRight(s, "-")
	}
	return s
}
