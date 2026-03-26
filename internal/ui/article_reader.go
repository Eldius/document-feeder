package ui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/eldius/document-feeder/internal/model"
)

type readerDoneMsg struct{}

type articleReaderModel struct {
	ctx      context.Context
	articles []model.Article
	index    int
	offset   int
	width    int
	height   int
	quitting bool
}

func waitForReaderContext(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return readerDoneMsg{}
	}
}

func (m articleReaderModel) Init() tea.Cmd {
	return waitForReaderContext(m.ctx)
}

func (m articleReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "left", "h", "p":
			if len(m.articles) > 0 {
				m.index = (m.index - 1 + len(m.articles)) % len(m.articles)
				m.offset = 0
			}
		case "right", "l", "n":
			if len(m.articles) > 0 {
				m.index = (m.index + 1) % len(m.articles)
				m.offset = 0
			}
		case "up", "k":
			m.offset--
		case "down", "j":
			m.offset++
		case "pgup", "b":
			m.offset -= m.bodyHeight()
		case "pgdown", "f":
			m.offset += m.bodyHeight()
		case "home", "g":
			m.offset = 0
		case "end", "G":
			m.offset = m.maxOffset()
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case readerDoneMsg:
		m.quitting = true
		return m, tea.Quit
	}

	m.offset = clamp(m.offset, 0, m.maxOffset())
	return m, nil
}

func (m articleReaderModel) View() string {
	if m.quitting {
		return "Done.\n"
	}
	if m.width == 0 || m.height == 0 {
		return "Loading...\n"
	}
	if len(m.articles) == 0 {
		return "No articles.\n"
	}

	headerStyle := lipgloss.NewStyle().Bold(true)
	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("242"))

	article := m.articles[m.index]
	title := strings.TrimSpace(article.Title)
	if title == "" {
		title = "(untitled)"
	}

	header := headerStyle.Render(fmt.Sprintf("Article %d/%d — %s", m.index+1, len(m.articles), title))
	footer := footerStyle.Render("j/k scroll • n/p next/prev • q quit")

	lines := m.articleLines(m.width)
	bodyHeight := m.bodyHeight()
	start := clamp(m.offset, 0, m.maxOffset())
	end := min(start+bodyHeight, len(lines))
	body := strings.Join(lines[start:end], "\n")

	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m articleReaderModel) bodyHeight() int {
	if m.height <= 2 {
		return 1
	}
	return m.height - 2
}

func (m articleReaderModel) maxOffset() int {
	lines := m.articleLines(m.width)
	bodyHeight := m.bodyHeight()
	if len(lines) <= bodyHeight {
		return 0
	}
	return len(lines) - bodyHeight
}

func (m articleReaderModel) articleLines(width int) []string {
	if len(m.articles) == 0 || width <= 0 {
		return []string{""}
	}

	article := m.articles[m.index]
	sectionTitleStyle := lipgloss.NewStyle().Bold(true)
	wrapWidth := max(width, 10)

	content, err := htmltomarkdown.ConvertString(article.Content)
	if err != nil {
		content = article.Content
	}
	var lines []string
	lines = append(lines, wrapLines(sectionTitleStyle.Render("Title"), wrapWidth)...)
	lines = append(lines, wrapLines(article.Title, wrapWidth)...)
	lines = append(lines, "")
	lines = append(lines, wrapLines(sectionTitleStyle.Render("Summary"), wrapWidth)...)
	lines = append(lines, wrapLines(article.Description, wrapWidth)...)
	lines = append(lines, "")
	lines = append(lines, wrapLines(sectionTitleStyle.Render("Content"), wrapWidth)...)
	lines = append(lines, wrapLines(content, wrapWidth)...)

	return lines
}

func wrapLines(text string, width int) []string {
	if strings.TrimSpace(text) == "" {
		return []string{""}
	}
	wrapped := lipgloss.NewStyle().Width(width).Render(text)
	return strings.Split(wrapped, "\n")
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

// ArticleReaderScreen runs a reader UI for the provided articles until the user exits or the context is done.
func ArticleReaderScreen(ctx context.Context, articles []model.Article) error {
	p := tea.NewProgram(articleReaderModel{
		ctx:      ctx,
		articles: articles,
	})
	_, err := p.Run()
	return err
}
