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
	contentReaderLinkStyle                                   = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10")). // Cyan color
				Underline(true)
	contentReaderQuestionStyle                                   = lipgloss.NewStyle().
					Foreground(lipgloss.Color("63")). // Cy
					Bold(true).
					Padding(1, 2)
)

type contentReaderModel struct {
	vp       viewport.Model
	pages    []*model.SearchResult
	ctx      context.Context
	question string
	title    string
	link     string
	score    float32
	idx      int
}

func (m *contentReaderModel) Init() tea.Cmd {
	return nil
}

func (m *contentReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmdVP tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case tea.KeyLeft.String():
			m.vp, cmdVP = m.vp.Update(tea.KeyLeft)
			if m.idx > 0 {
				m.idx--
			}
		case tea.KeyRight.String():
			m.vp, cmdVP = m.vp.Update(tea.KeyRight)
			if m.idx < len(m.pages)-1 {
				m.idx++
			}
		}
	case tea.WindowSizeMsg:
		m.vp.Width = screenMaxUIWidth(msg.Width)
		m.vp.Height = contentReaderViewportMaxHeight(msg.Height)
	}

	m.refreshContent()
	m.refreshTitle()
	m.refreshLink()
	m.refreshSimilarityScore()
	if cmdVP != nil {
		var cmdVPAux tea.Cmd
		m.vp, cmdVPAux = m.vp.Update(msg)
		cmdVP = tea.Batch(cmdVP, cmdVPAux)
	} else {
		m.vp, cmdVP = m.vp.Update(msg)
	}
	return m, tea.Batch(cmdVP, cmd)
}

func (m *contentReaderModel) View() string {
	return contentReaderQuestionStyle.Render("Question: "+m.question) +
		"\n" +
		contentReaderTitleStyle.Render(m.title) +
		"\n" +
		contentReaderViewportStyle.Render(m.vp.View()) +
		"\n" +
		"Article source: " + contentReaderLinkStyle.Render(m.link) +
		"\n" +
		"Similarity score: " + contentReaderLinkStyle.Render(fmt.Sprintf("%.2f", m.score)) + "/1" +
		"\n" +
		contentReaderMenuStyle.Render("Press q / ctrl+c to exit, ←/→ to navigate through articles, ↑/↓ to scroll")
}

func (m *contentReaderModel) refreshTitle() {
	page := " (" + strconv.Itoa(m.idx+1) + " of " + strconv.Itoa(len(m.pages)) + ")"
	title := m.pages[m.idx].Article.Title
	pad := strings.Repeat(" ", m.vp.Width-(len(title)+len(page)))
	m.title = title + pad + page
}

func (m *contentReaderModel) refreshContent() {
	article, err := htmltomarkdown.ConvertString("<!DOCTYPE html><html><body>" + m.pages[m.idx].Article.Content + "</body></html>")
	if err != nil {
		article = m.pages[m.idx].Article.Content
	}
	m.vp.SetContent(lineWrap(article, m.vp.Width))
}

func (m *contentReaderModel) refreshLink() {
	m.link = m.pages[m.idx].Article.Link
}

func (m *contentReaderModel) refreshSimilarityScore() {
	m.score = m.pages[m.idx].Similarity
}

func newContentReaderModel(ctx context.Context, articles []*model.SearchResult, question string) *contentReaderModel {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Failed getting terminal size:", err)
		return nil
	}
	fmt.Printf("Initial terminal size: %dx%d\n", width, height)
	vp := viewport.New(screenMaxUIWidth(width), contentReaderViewportMaxHeight(height))
	vp.SetContent("")
	return &contentReaderModel{
		ctx:      ctx,
		pages:    articles,
		question: question,
	}
}

func contentReaderViewportMaxHeight(height int) int {
	return height - 16
}

func ContentReader(ctx context.Context, articles []*model.SearchResult, question string) error {
	if len(articles) == 0 {
		return fmt.Errorf("article list is empty")
	}
	m := newContentReaderModel(ctx, articles, question)
	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}
