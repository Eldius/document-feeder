package ui

import (
	"context"
	"fmt"
	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eldius/document-feeder/internal/model"
	"golang.org/x/term"
	"os"
	"strconv"
	"strings"
)

var (
	_ tea.Model = &contentReaderModel{}

	contentReaderTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(1, 2)
	}()
	contentReaderViewportStyle = lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)
	contentReaderMenuStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(1, 2)
	contentReaderLinkStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")). // Cyan color
		Underline(true)
)

type contentReaderModel struct {
	vp    viewport.Model
	title string
	link  string
	pages []model.Article
	idx   int
	ctx   context.Context
}

func (m *contentReaderModel) Init() tea.Cmd {
	return nil
}

func (m *contentReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case tea.KeyLeft.String():
			if m.idx > 0 {
				m.idx--
			}
		case tea.KeyRight.String():
			if m.idx < len(m.pages)-1 {
				m.idx++
			}
		}
	case tea.WindowSizeMsg:
		m.vp.Width = screenMaxUIWidth(msg.Width)
		m.vp.Height = contentReaderViewportMaxHeight(msg.Height)
	}
	m.vp.SetContent(m.Content())
	m.link = m.pages[m.idx].Link
	m.title = m.Title()
	var cmdVP tea.Cmd
	m.vp, cmdVP = m.vp.Update(msg)
	return m, tea.Batch(cmdVP, cmd)
}

func (m *contentReaderModel) View() string {
	return contentReaderTitleStyle.Render(m.title) +
		"\n\n" +
		contentReaderViewportStyle.Render(m.vp.View()) +
		"\n" +
		"Article source: " + contentReaderLinkStyle.Render(m.link) + "\n" +
		"\n" +
		contentReaderMenuStyle.Render("Press q / ctrl+c to exit, ←/→ to navigate through articles, ↑/↓ to scroll")
}

func (m *contentReaderModel) Title() string {
	page := " (" + strconv.Itoa(m.idx+1) + " of " + strconv.Itoa(len(m.pages)) + ")"
	title := m.pages[m.idx].Title
	pad := strings.Repeat(" ", m.vp.Width-(len(title)+len(page)))
	return title + pad + page
}

func (m *contentReaderModel) Content() string {
	article, err := htmltomarkdown.ConvertString("<!DOCTYPE html><html><body>" + m.pages[m.idx].Content + "</body></html>")
	if err != nil {
		article = m.pages[m.idx].Content
	}
	return lineWrap(article, m.vp.Width)

}

func newContentReaderModel(ctx context.Context, articles []model.Article) *contentReaderModel {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Failed getting terminal size:", err)
		return nil
	}
	fmt.Printf("Initial terminal size: %dx%d\n", width, height)
	vp := viewport.New(screenMaxUIWidth(width), contentReaderViewportMaxHeight(height))
	vp.SetContent("")
	return &contentReaderModel{
		ctx:   ctx,
		pages: articles,
	}
}

func contentReaderViewportMaxHeight(height int) int {
	return height - 15
}

func ContentReader(ctx context.Context, articles []model.Article) error {
	m := newContentReaderModel(ctx, articles)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
