package ui

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/costap/vger/internal/domain"
)

// LCARS colour palette
var (
	amber   = lipgloss.Color("#FF9900")
	blue    = lipgloss.Color("#99CCFF")
	red     = lipgloss.Color("#CC4444")
	white   = lipgloss.Color("#FFFFFF")
	dimGrey = lipgloss.Color("#666666")
)

var (
	labelStyle = lipgloss.NewStyle().Foreground(amber).Bold(false)
	bodyStyle  = lipgloss.NewStyle().Foreground(white)
	dimStyle   = lipgloss.NewStyle().Foreground(dimGrey)
	blueStyle  = lipgloss.NewStyle().Foreground(blue)
	redStyle   = lipgloss.NewStyle().Foreground(red).Bold(false)
)

const banner = `
  ██╗   ██╗ ██████╗ ███████╗██████╗ 
  ██║   ██║██╔════╝ ██╔════╝██╔══██╗
  ██║   ██║██║  ███╗█████╗  ██████╔╝
  ╚██╗ ██╔╝██║   ██║██╔══╝  ██╔══██╗
   ╚████╔╝ ╚██████╔╝███████╗██║  ██║
    ╚═══╝   ╚═════╝ ╚══════╝╚═╝  ╚═╝
  ════════════════════════════════════
  KNOWLEDGE  ASSIMILATION  UNIT  001
`

// Stardate returns a TNG-style stardate string derived from the current UTC time.
func Stardate() string {
	now := time.Now().UTC()
	dayOfYear := float64(now.YearDay())
	daysInYear := 365.0
	if isLeap(now.Year()) {
		daysInYear = 366.0
	}
	fraction := dayOfYear / daysInYear
	stardate := float64(now.Year()-1900)*10 + fraction*10
	return fmt.Sprintf("%.1f", math.Round(stardate*10)/10)
}

func isLeap(y int) bool {
	return y%4 == 0 && (y%100 != 0 || y%400 == 0)
}

// Header prints the V'Ger ASCII banner and stardate.
func Header() {
	fmt.Println(labelStyle.Render(banner))
	fmt.Println(dimStyle.Render(strings.Repeat("─", 60)))
	fmt.Printf("  %s  %s\n\n",
		labelStyle.Render("STARDATE"),
		blueStyle.Render(Stardate()),
	)
}

// Status prints a labelled status line with a stardate prefix.
func Status(msg string) {
	fmt.Printf("  %s  %s\n",
		dimStyle.Render("["+Stardate()+"]"),
		bodyStyle.Render(msg),
	)
}

// Complete prints a completion message in blue.
func Complete(msg string) {
	fmt.Printf("\n  %s\n\n",
		blueStyle.Render("── "+strings.ToUpper(msg)+" ──"),
	)
}

// RedAlert prints an error framed as a RED ALERT banner.
func RedAlert(err error) {
	border := strings.Repeat("█", 60)
	fmt.Println()
	fmt.Println(redStyle.Render(border))
	fmt.Printf("  %s  %s\n", redStyle.Render("RED ALERT"), bodyStyle.Render(err.Error()))
	fmt.Println(redStyle.Render(border))
	fmt.Println()
}

// SectionHeader prints a named section divider in amber.
func SectionHeader(title string) {
	fmt.Printf("\n  %s\n  %s\n",
		labelStyle.Render("▐▌ "+strings.ToUpper(title)),
		dimStyle.Render(strings.Repeat("─", 56)),
	)
}

// LabelStyle returns the amber label style for use outside this package.
func LabelStyle() lipgloss.Style { return labelStyle }

// DimStyle returns the dim grey style for use outside this package.
func DimStyle() lipgloss.Style { return dimStyle }

// BlueStyle returns the blue style for use outside this package.
func BlueStyle() lipgloss.Style { return blueStyle }

// RedStyle returns the red style for use outside this package.
func RedStyle() lipgloss.Style { return redStyle }

// GreenStyle returns a green style for use outside this package.
func GreenStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#44CC88"))
}

// Field prints a key/value pair in LCARS style.
func Field(key, value string) {
	fmt.Printf("  %s  %s\n",
		labelStyle.Render(fmt.Sprintf("%-20s", strings.ToUpper(key))),
		bodyStyle.Render(value),
	)
}

// ListingRow prints a numbered video listing row with cache indicator, duration,
// view count, and (for cached videos) a line of Gemini-derived technology tags.
// Pass a non-nil entry to show the ★ indicator and tech tags; nil renders a dim dot.
func ListingRow(index int, v domain.VideoListing, entry *domain.CachedAnalysis) {
	date := v.PublishedAt
	if len(date) >= 10 {
		date = date[:10]
	}

	var indicator string
	if entry != nil {
		indicator = labelStyle.Render("★")
	} else {
		indicator = dimStyle.Render("·")
	}

	// Build optional metadata tokens (duration, views).
	var meta []string
	if dur := formatISO8601Duration(v.Duration); dur != "" {
		meta = append(meta, dimStyle.Render(dur))
	}
	if v.ViewCount > 0 {
		meta = append(meta, dimStyle.Render(formatViewCount(v.ViewCount)+" views"))
	}
	metaStr := ""
	if len(meta) > 0 {
		metaStr = "  " + strings.Join(meta, "  ")
	}

	fmt.Printf("  %s %s  %s%s  %s\n    %s\n",
		labelStyle.Render(fmt.Sprintf("%3d", index)),
		indicator,
		dimStyle.Render(date),
		metaStr,
		bodyStyle.Render(v.Title),
		blueStyle.Render(v.URL),
	)

	// For cached videos render a compact tag line with Gemini-extracted tech names
	// and speaker names.
	if entry != nil {
		var chips []string
		for _, s := range entry.Speakers() {
			chips = append(chips, dimStyle.Render("("+s+")"))
		}
		for _, t := range entry.Tags() {
			chips = append(chips, blueStyle.Render("["+t+"]"))
		}
		if len(chips) > 0 {
			fmt.Printf("    %s\n", strings.Join(chips, " "))
		}
	}
}

// formatISO8601Duration converts an ISO 8601 duration string (e.g. "PT1H15M32S")
// into a compact human-readable form ("1h15m", "45m", "8m30s").
// Returns "" for empty or unrecognisable input.
func formatISO8601Duration(iso string) string {
	if iso == "" {
		return ""
	}
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	m := re.FindStringSubmatch(iso)
	if m == nil {
		return ""
	}
	hours, _ := strconv.Atoi(m[1])
	mins, _ := strconv.Atoi(m[2])
	secs, _ := strconv.Atoi(m[3])

	switch {
	case hours > 0 && secs == 0:
		return fmt.Sprintf("%dh%dm", hours, mins)
	case hours > 0:
		return fmt.Sprintf("%dh%dm%ds", hours, mins, secs)
	case mins > 0 && secs == 0:
		return fmt.Sprintf("%dm", mins)
	case mins > 0:
		return fmt.Sprintf("%dm%ds", mins, secs)
	default:
		return fmt.Sprintf("%ds", secs)
	}
}

// formatViewCount formats a view count as a compact string: "1.2M", "42K", "999".
func formatViewCount(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.0fK", float64(n)/1_000)
	default:
		return strconv.FormatInt(n, 10)
	}
}

// PlaylistRow prints a numbered playlist row: index, date, video count, title, url.
func PlaylistRow(index int, publishedAt, title, url string, videoCount int64) {
	date := publishedAt
	if len(date) >= 10 {
		date = date[:10]
	}
	fmt.Printf("  %s  %s  %s  %s\n    %s\n",
		labelStyle.Render(fmt.Sprintf("%3d", index)),
		dimStyle.Render(date),
		dimStyle.Render(fmt.Sprintf("[%3d videos]", videoCount)),
		bodyStyle.Render(title),
		blueStyle.Render(url),
	)
}
