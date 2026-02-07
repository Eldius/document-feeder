package ui

import (
	"charm.land/lipgloss/v2"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eldius/document-feeder/internal/client/ollama"
	"log/slog"
	"strings"
	"time"
)

func DisplayModels(ms *ollama.ModelsResponse) {
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("255")).
		Bold(true)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("39")).
		Bold(false)
	for _, m := range ms.Models {
		fmt.Println("+---------------------------------------------")
		fmt.Println("|", labelStyle.Render("- name:    "), valueStyle.Render(m.Name))
		fmt.Println("|", labelStyle.Render("  size:    "), valueStyle.Render(fmt.Sprintf("%d", m.Size)))
		fmt.Println("|", labelStyle.Render("  modified:"), valueStyle.Render(m.ModifiedAt.Format("2006-01-02 15:04:05")))
		fmt.Println("|", labelStyle.Render("  details:"))
		fmt.Println("|", labelStyle.Render("     format:"), valueStyle.Render(m.Details.Format))
		fmt.Println("|", labelStyle.Render("     families:"))
		for _, f := range m.Details.Families {
			fmt.Println("|", labelStyle.Render("        -"), valueStyle.Render(f))
		}
		fmt.Println("+---------------------------------------------")
	}

}

// tickMsg is sent periodically to animate the dots.
type tickMsg time.Time

// tickCmd returns a command that sends a tickMsg every second.
func tickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

const (
	screenPadding = 2
)

func screenMaxUIWidth(screenWidth int) int {
	calculatedWidth := screenWidth - (screenPadding*2 + 4)
	slog.With("screen_width", screenWidth, "max_ui_width", calculatedWidth).Debug("computing max ui width")
	return calculatedWidth
}

func lineWrap(s string, k int) string {
	var result []string
	// Use strings.Fields to handle all kinds of whitespace and get clean words.
	words := strings.Fields(s)

	if len(words) == 0 {
		return ""
	}

	currentLine := words[0]

	for i := 1; i < len(words); i++ {
		word := words[i]
		// Check if adding the next word (plus a space) exceeds the limit.
		if len(currentLine)+1+len(word) <= k {
			currentLine += " " + word
		} else {
			// Start a new line if the word doesn't fit.
			result = append(result, currentLine)
			currentLine = word
		}
	}

	// Add the last accumulated line.
	result = append(result, currentLine)

	return strings.Join(result, "\n")
}
