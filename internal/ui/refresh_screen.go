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
	"golang.org/x/term"
	"os"
	"strings"
	"time"
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

	vpStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()). // Use RoundedBorder or NormalBorder
		BorderForeground(lipgloss.Color("63")). // Set the border color
		Padding(1, 2) // Add some padding inside the border

	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

	refreshSuccessStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#04B575")).
		Bold(true)

	refreshErrorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555"))

	processingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0a4d94")).
		Bold(true)
	notProcessedYetStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#777777")).
		Bold(false)
)

type refreshScreenModel struct {
	a         *adapter.FeedAdapter
	viewport  viewport.Model
	progress  progress.Model
	ctx       context.Context
	cancel    context.CancelFunc
	feeds     []*model.Feed
	feedsList []string
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
	return tea.Batch(refreshFeedMsg(), tickCmd())
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
	case tickMsg:
		return m, m.progress.SetPercent(float64(m.idx) / float64(m.feedCount))

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - padding*2 - 4
		//if m.progress.Width > maxWidth {
		//	m.progress.Width = maxWidth
		//}
		m.viewport.Width = msg.Width - padding*2 - 4
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case refreshNextFeed:
		cmd := m.progress.SetPercent(float64(m.idx) / float64(m.feedCount))
		return m, tea.Batch(cmd, m.process())
	default:
		return m, nil
	}
}

func (m *refreshScreenModel) process() tea.Cmd {
	return func() tea.Msg {
		feed := m.feeds[m.idx]
		m.feedsList[m.idx] = processingStyle.Render(fmt.Sprintf("⏳  %s", feed.Title))
		err := m.a.RefreshFeed(m.ctx, feed)
		if err != nil {
			m.feedsList[m.idx] = refreshErrorStyle.Render(fmt.Sprintf("✗  %s", feed.Title)) + "\n"
		} else {
			m.feedsList[m.idx] = refreshSuccessStyle.Render(fmt.Sprintf("✓  %s (%d articles)", feed.Title, len(feed.Items))) + "\n"
		}
		m.idx++
		if m.idx == m.feedCount {
			return nil
		}
		return refreshNextFeed{}
	}
}

func (m *refreshScreenModel) View() string {
	pad := strings.Repeat(" ", padding)
	if m.idx == m.feedCount {
		return m.viewport.View() + "\nFinished processing feeds!" + "\n\n" +
			pad + m.progress.View() + "\n\n" +
			pad + helpStyle("Press q / ctrl+c to quit")
	}

	m.viewport.SetContent(strings.Join(m.feedsList, "\n"))

	return m.viewport.View() + "\n\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press q / ctrl+c to quit")
}

type refreshNextFeed struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second*1, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func newRefreshScreenModel(ctx context.Context, a *adapter.FeedAdapter) (*refreshScreenModel, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, fmt.Errorf("getting terminal size: %w", err)
	}
	fmt.Printf("Initial terminal size: %dx%d\n", width, height)
	feeds, err := a.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting feeds: %w", err)
	}
	vp := viewport.New(width-padding*2-4, height-4)
	vp.Style = vpStyle
	feedsList := make([]string, len(feeds))
	for i, feed := range feeds {
		feedsList[i] = notProcessedYetStyle.Render("   " + feed.Title)
	}

	ctx, cancel := context.WithCancel(ctx)
	return &refreshScreenModel{
		progress:  progress.New(progress.WithDefaultGradient(), progress.WithWidth(width-padding*2-4)),
		ctx:       ctx,
		cancel:    cancel,
		feeds:     feeds,
		feedsList: feedsList,
		feedCount: len(feeds),
		idx:       0,
		a:         a,
		viewport:  vp,
	}, nil

}

func RefreshScreen(ctx context.Context, a *adapter.FeedAdapter) error {
	m, err := newRefreshScreenModel(ctx, a)
	if err != nil {
		return fmt.Errorf("creating refresh screen model: %w", err)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		err = fmt.Errorf("error running refresh screen: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}
