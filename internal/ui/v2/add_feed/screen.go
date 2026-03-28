package add_feed

import (
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"context"
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/document-feeder/internal/model"
	"github.com/eldius/document-feeder/internal/ui/v2/screen"
	"strings"
)

var (
	_ tea.Model = &addFeedScreen{}

	//waitingStyle = lipgloss.NewStyle().Foreground(lipgloss.Blue).Bold(true)
	//failedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Red).Bold(true)
	//successStyle = lipgloss.NewStyle().Foreground(lipgloss.Green).Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#04B575")).
			Bold(true)

	failedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555"))

	processingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#0a4d94")).
			Bold(true)
	waitingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#777777")).
			Bold(true)

	viewPortStyle    = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Margin(1, 2)
	progressBarStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2).Margin(1, 2)
)

const (
	//waitingIcon    = "⏳"
	//failedIcon     = "\\u274C"
	//processingIcon = "\\U0001F52"

	waitingIcon    = "⏳"
	failedIcon     = "✗ "
	processingIcon = "->"
	successIcon    = "✓ "
)

type addFeedScreen struct {
	vp               *viewport.Model
	progressBar      *progress.Model
	ctx              context.Context
	screenSize       screen.Size
	a                *adapter.FeedAdapter
	feeds            []model.Feed
	processingStatus []string
	idx              int
}

func (a *addFeedScreen) Init() tea.Cmd {
	return func() tea.Msg {
		return nextFeedItem{
			idx: 0,
		}
	}
}

func (a *addFeedScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		var vp viewport.Model
		vp, cmd = a.vp.Update(msg)
		a.vp = &vp
		switch msg.String() {
		case "q", string(tea.KeyEsc), "ctrl+c":
			return a, tea.Quit
		}

	case tea.WindowSizeMsg:
		s := screen.Size{
			Width:  msg.Width,
			Height: msg.Height,
		}

		width := viewportWidth(s)
		a.vp.SetWidth(width)
		height := viewportHeight(s)
		a.vp.SetHeight(height)
		a.progressBar.SetWidth(width)
		viewPortStyle.Width(width + 2)
		viewPortStyle.Height(height + 2)
		progressBarStyle.Width(width + 2)
	case nextFeedItem:
		a.idx = msg.idx
		a.processingStatus[a.idx] = processingStyle.Render(processingIcon + "\t" + a.feeds[a.idx].Title)
		a.progressBar.SetPercent(float64(a.idx) / float64(len(a.feeds)))
		a.vp.SetContent(strings.Join(a.processingStatus, "\n"))
		cmd = a.Next()
	}
	var vp viewport.Model
	if cmd != nil {
		_vp, _cmd := a.vp.Update(msg)
		vp = _vp
		cmd = tea.Batch(cmd, _cmd)
	} else {
		vp, cmd = a.vp.Update(msg)
	}

	a.vp = &vp
	return a, cmd
}

func (a *addFeedScreen) View() tea.View {
	return tea.NewView(viewPortStyle.Render(a.vp.View()) + progressBarStyle.Render(a.progressBar.View()) + "%\n")
}

func newScreen(ctx context.Context, a *adapter.FeedAdapter) (*addFeedScreen, error) {
	sc, err := screen.GetScreenSize()
	if err != nil {
		return nil, fmt.Errorf("getting screen size: %w", err)
	}

	vp := viewport.New(viewport.WithHeight(viewportWidth(*sc)), viewport.WithWidth(viewportHeight(*sc)))
	progressBar := progress.New(progress.WithWidth(sc.Width))

	feeds, err := a.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting feeds: %w", err)
	}

	var feedItems []model.Feed
	var feedsProcessing []string
	for _, feed := range feeds {
		feedsProcessing = append(feedsProcessing, waitingStyle.Render(waitingIcon+"\t"+feed.Title))
		feedItems = append(feedItems, *feed)
	}

	return &addFeedScreen{
		vp:               &vp,
		progressBar:      &progressBar,
		ctx:              ctx,
		screenSize:       *sc,
		a:                a,
		feeds:            feedItems,
		processingStatus: feedsProcessing,
		idx:              0,
	}, nil
}

type nextFeedItem struct {
	idx int
}

func (a *addFeedScreen) Next() tea.Cmd {
	return func() tea.Msg {
		feed := a.feeds[a.idx]
		if err := a.a.RefreshFeed(a.ctx, &feed); err != nil {
			a.processingStatus[a.idx] = failedStyle.Render(fmt.Sprintf("%s\t%s (%s)", failedIcon, feed.Title, err.Error()))
		} else {
			a.processingStatus[a.idx] = successStyle.Render(fmt.Sprintf("%s\t%s (%d articles)", successIcon, feed.Title, len(feed.Items)))
		}
		return &nextFeedItem{
			idx: a.idx + 1,
		}
	}
}

const (
	screenPadding = 4
)

func viewportWidth(screenSize screen.Size) int {
	return screenSize.Width - 2*screenPadding
}

func viewportHeight(screenSize screen.Size) int {
	return screenSize.Height - 2*screenPadding - 4
}

func Start(ctx context.Context, a *adapter.FeedAdapter) error {
	m, err := newScreen(ctx, a)
	if err != nil {
		return fmt.Errorf("creating screen: %w", err)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("running program: %w", err)
	}
	return nil
}
