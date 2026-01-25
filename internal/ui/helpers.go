package ui

import (
	"charm.land/lipgloss/v2"
	"fmt"
	"github.com/eldius/document-feed-embedder/internal/client/ollama"
)

func DisplayModels(ms *ollama.OllamaModelsResponse) {
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
