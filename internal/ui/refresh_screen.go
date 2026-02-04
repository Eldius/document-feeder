package ui

import (
	"context"
	"fmt"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/model"
	"strings"
)

const (
	padding  = 2
	maxWidth = 80
)

var (
	_ tea.Model = &refreshScreenModel{}

	titleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	infoStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Left = "┤"
		return titleStyle.BorderStyle(b)
	}()

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

	refreshSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)
	refreshErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))
)

type refreshScreenModel struct {
	a         *adapter.FeedAdapter
	viewport  viewport.Model
	progress  progress.Model
	ctx       context.Context
	cancel    context.CancelFunc
	feeds     []*model.Feed
	content   string
	idx       int
	feedCount int
	ready     bool
}

func refreshFeedMsg() tea.Cmd {
	return func() tea.Msg {
		return refreshNextFeed{}
	}
}

func (m *refreshScreenModel) Init() tea.Cmd {
	return refreshFeedMsg()
}

func (m *refreshScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.cancel()
			return m, tea.Quit

		default:
			vp, cmd := m.viewport.Update(msg)
			m.viewport = vp
			return m, cmd
		}

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		if m.progress.Width > maxWidth {
			m.progress.Width = maxWidth
		}
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case refreshNextFeed:
		cmd := m.progress.IncrPercent(1 / float64(m.feedCount))
		return m, tea.Batch(cmd, m.process())
	default:
		return m, nil
	}
}

func (m *refreshScreenModel) process() tea.Cmd {
	return func() tea.Msg {
		feed := m.feeds[m.idx]
		err := m.a.RefreshFeed(m.ctx, feed)
		if err != nil {
			m.content += refreshErrorStyle.Render("✗ %s", feed.Title) + "\n"
		} else {
			m.content += refreshSuccessStyle.Render(fmt.Sprintf("✓ %s (%d articles)", feed.Title, len(feed.Items))) + "\n"
		}
		m.idx++
		if m.idx == m.feedCount {
			return tea.Quit
		}
		return refreshNextFeed{}
	}
}

func (m *refreshScreenModel) View() string {
	m.viewport.SetContent(m.content)
	pad := strings.Repeat(" ", padding)
	return m.viewport.View() + "\n" + fmt.Sprintf("Processing feed %s", m.feeds[m.idx].Title) + "\n\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press q / ctrl+c to quit")
}

type refreshNextFeed struct{}

func RefreshScreen(ctx context.Context, a *adapter.FeedAdapter) error {
	feeds, err := a.All(ctx)
	if err != nil {
		return fmt.Errorf("error getting feeds: %w", err)
	}
	vp := viewport.New(20, 10)

	ctx, cancel := context.WithCancel(ctx)
	m := &refreshScreenModel{
		progress:  progress.New(progress.WithDefaultGradient()),
		ctx:       ctx,
		cancel:    cancel,
		feeds:     feeds,
		feedCount: len(feeds),
		idx:       0,
		a:         a,
		viewport:  vp,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		err = fmt.Errorf("error running refresh screen: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}
