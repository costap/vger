package ui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
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

// Field prints a key/value pair in LCARS style.
func Field(key, value string) {
	fmt.Printf("  %s  %s\n",
		labelStyle.Render(fmt.Sprintf("%-20s", strings.ToUpper(key))),
		bodyStyle.Render(value),
	)
}

// ListingRow prints a numbered video listing row: index, date, title, url.
func ListingRow(index int, publishedAt, title, url string) {
	date := publishedAt
	if len(date) >= 10 {
		date = date[:10]
	}
	fmt.Printf("  %s  %s  %s\n    %s\n",
		labelStyle.Render(fmt.Sprintf("%3d", index)),
		dimStyle.Render(date),
		bodyStyle.Render(title),
		blueStyle.Render(url),
	)
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
