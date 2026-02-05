package ui

import (
	"charm.land/lipgloss/v2"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/eldius/document-feeder/internal/client/ollama"
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
