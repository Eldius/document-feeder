package ui

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/eldius/document-feeder/internal/adapter"
	"golang.org/x/term"
)

var (
	_ tea.Model = &addScreenModel{}

	addTitleStyle = func() lipgloss.Style {
		b := lipgloss.RoundedBorder()
		b.Right = "├"
		return lipgloss.NewStyle().BorderStyle(b).Padding(0, 1)
	}()

	addVpStyle                                         = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).       // Use RoundedBorder or NormalBorder
			BorderForeground(lipgloss.Color("63")). // Set the border color
			Padding(1, 2)                           // Add some screenPadding inside the border

	addHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

	addSuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	addErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	addProcessingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#0a4d94")).
				Bold(true)
	addNotProcessedYetStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#777777")).
				Bold(true)
)

type addScreenModel struct {
	a            *adapter.FeedAdapter
	viewport     viewport.Model
	progress     progress.Model
	ctx          context.Context
	cancel       context.CancelFunc
	feedsList    []string
	feedsURLList []string
	title        string
	idx          int
	feedCount    int
}

func addFeedMsg() tea.Cmd {
	return func() tea.Msg {
		return addNextFeed{}
	}
}

func (m *addScreenModel) Init() tea.Cmd {
	return tea.Batch(addFeedMsg(), tickCmd(time.Second*1))
}

func (m *addScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.idx == m.feedCount {
			return m, m.progress.SetPercent(1)
		}
		return m, tea.Batch(m.progress.SetPercent(float64(m.idx)/float64(m.feedCount)), tickCmd(time.Second*1))

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - screenPadding*2 - 4
		m.title = addScreenTitleContent(m.title, msg.Width)
		m.viewport.Width = msg.Width - screenPadding*2 - 4
		return m, nil

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case addNextFeed:
		cmd := m.progress.SetPercent(float64(m.idx) / float64(m.feedCount))
		if m.idx == m.feedCount {
			return m, cmd
		}
		return m, tea.Batch(cmd, m.process())
	default:
		return m, nil
	}
}

func (m *addScreenModel) process() tea.Cmd {
	return func() tea.Msg {
		feedURL := m.feedsURLList[m.idx]
		m.feedsList[m.idx] = addProcessingStyle.Render(fmt.Sprintf("⏳  %s", feedURL))
		feed, err := m.a.Parse(m.ctx, feedURL)
		if err != nil {
			m.feedsList[m.idx] = addErrorStyle.Render(fmt.Sprintf("✗  %s", feedURL)) + "\n"
		} else {
			m.feedsList[m.idx] = addSuccessStyle.Render(fmt.Sprintf("✓  %s (%d articles)", feed.Title, len(feed.Items))) + "\n"
		}
		m.idx += 1
		if m.idx == m.feedCount {
			m.viewport.SetContent(strings.Join(m.feedsList, "\n"))
			return tea.Batch(tickCmd(time.Second*1), m.progress.SetPercent(1), func() tea.Msg {
				return progress.FrameMsg{}
			})
		}
		return addNextFeed{}
	}
}

func (m *addScreenModel) View() string {
	pad := strings.Repeat(" ", screenPadding)
	if m.idx == m.feedCount {
		return m.viewport.View() + "\nFinished processing feeds!" + "\n\n" +
			pad + m.progress.View() + "\n\n" +
			pad + addHelpStyle("Press q / ctrl+c to quit")
	}

	m.viewport.SetContent(strings.Join(m.feedsList, "\n"))

	return addTitleStyle.Render(m.title) + "\n\n" + m.viewport.View() + "\n" +
		pad + m.progress.View() + "\n" +
		pad + addHelpStyle("Press q / ctrl+c to quit")
}

type addNextFeed struct{}

func newAddScreenModel(ctx context.Context, a *adapter.FeedAdapter, feedURLs []string) (*addScreenModel, error) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return nil, fmt.Errorf("getting terminal size: %w", err)
	}
	fmt.Printf("Initial terminal size: %dx%d\n", width, height)
	vp := viewport.New(screenMaxUIWidth(width), addScreenViewportHeight(height))
	vp.Style = addVpStyle
	feedsList := make([]string, len(feedURLs))
	for i, feed := range feedURLs {
		feedsList[i] = addNotProcessedYetStyle.Render("   " + feed)
	}

	ctx, cancel := context.WithCancel(ctx)
	return &addScreenModel{
		progress:     progress.New(progress.WithDefaultGradient(), progress.WithWidth(screenMaxUIWidth(width))),
		ctx:          ctx,
		cancel:       cancel,
		feedsList:    feedsList,
		feedsURLList: feedURLs,
		feedCount:    len(feedsList),
		idx:          0,
		a:            a,
		viewport:     vp,
		title:        addScreenTitleContent("Adding feeds", width),
	}, nil
}

func addScreenTitleContent(title string, screenWidth int) string {
	maxViewportWidth := screenMaxUIWidth(screenWidth)
	slog.With("title_length", len(title), "max_width", maxViewportWidth).Debug(
		"Calculating title padding size",
	)
	slog.With("title_length", len(title), "max_width", maxViewportWidth).Debug(
		"Calculating title padding size",
	)
	title = strings.TrimSpace(
		strings.ReplaceAll(title, "\n", " "),
	)
	slog.With("title_length", len(title), "max_width", maxViewportWidth).Debug("Trimmed title to:")
	titlePaddingSize := (maxViewportWidth - (len(title) + 2)) / 2
	return strings.Repeat(" ", titlePaddingSize) + title + strings.Repeat(" ", titlePaddingSize-1)
}

func addScreenViewportHeight(screenHeight int) int {
	return screenHeight - 7
}

func AddScreen(ctx context.Context, a *adapter.FeedAdapter, feedURLs []string) error {
	m, err := newAddScreenModel(ctx, a, feedURLs)
	if err != nil {
		return fmt.Errorf("creating add screen model: %w", err)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		err = fmt.Errorf("error running add screen: %w", err)
		fmt.Println(err)
		return err
	}

	return nil
}
