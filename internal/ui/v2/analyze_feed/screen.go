package analyze_feed

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
	_ tea.Model = &analyzeFeedScreen{}

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
	titleStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("#00D7FF")).Bold(true).Margin(1, 2)
)

const (
	waitingIcon    = "⏳"
	failedIcon     = "✗ "
	processingIcon = "->"
	successIcon    = "✓ "
)

type analyzeFeedScreen struct {
	vp               *viewport.Model
	progressBar      *progress.Model
	ctx              context.Context
	screenSize       screen.Size
	a                *adapter.FeedAdapter
	feed             *model.Feed
	articles         []model.Article
	processingStatus []string
	idx              int
	skipCache        bool
}

func (a *analyzeFeedScreen) Init() tea.Cmd {
	return func() tea.Msg {
		return nextArticleItem{
			idx: 0,
		}
	}
}

func (a *analyzeFeedScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
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

	case nextArticleItem:
		if msg.idx >= len(a.articles) {
			return a, tea.Quit
		}
		a.idx = msg.idx
		a.processingStatus[a.idx] = processingStyle.Render(processingIcon + "\t" + a.articles[a.idx].Title)
		a.progressBar.SetPercent(float64(a.idx) / float64(len(a.articles)))
		a.vp.SetContent(strings.Join(a.processingStatus, "\n"))
		a.vp.GotoBottom()
		return a, a.Next()
	}

	var vp viewport.Model
	vp, cmd = a.vp.Update(msg)
	a.vp = &vp
	return a, cmd
}

func (a *analyzeFeedScreen) View() tea.View {
	return tea.NewView(titleStyle.Render("Analyzing Feed: "+a.feed.Title) + "\n" +
		viewPortStyle.Render(a.vp.View()) + "\n" +
		progressBarStyle.Render(a.progressBar.View()) + "\n")
}

func newScreen(ctx context.Context, a *adapter.FeedAdapter, feedLink string, skipCache bool) (*analyzeFeedScreen, error) {
	sc, err := screen.GetScreenSize()
	if err != nil {
		return nil, fmt.Errorf("getting screen size: %w", err)
	}

	vp := viewport.New(viewport.WithWidth(viewportWidth(*sc)), viewport.WithHeight(viewportHeight(*sc)))
	progressBar := progress.New(progress.WithWidth(sc.Width))

	feeds, err := a.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting feeds: %w", err)
	}

	var targetFeed *model.Feed
	for _, f := range feeds {
		if f.FeedLink == feedLink {
			targetFeed = f
			break
		}
	}

	if targetFeed == nil {
		return nil, fmt.Errorf("feed not found: %s", feedLink)
	}

	var articlesProcessing []string
	for _, article := range targetFeed.Items {
		articlesProcessing = append(articlesProcessing, waitingStyle.Render(waitingIcon+"\t"+article.Title))
	}

	return &analyzeFeedScreen{
		vp:               &vp,
		progressBar:      &progressBar,
		ctx:              ctx,
		screenSize:       *sc,
		a:                a,
		feed:             targetFeed,
		articles:         targetFeed.Items,
		processingStatus: articlesProcessing,
		idx:              0,
		skipCache:        skipCache,
	}, nil
}

type nextArticleItem struct {
	idx int
}

func (a *analyzeFeedScreen) Next() tea.Cmd {
	return func() tea.Msg {
		article := a.articles[a.idx]
		if _, err := a.a.AnalyzeArticle(a.ctx, article, a.skipCache); err != nil {
			a.processingStatus[a.idx] = failedStyle.Render(fmt.Sprintf("%s\t%s (%s)", failedIcon, article.Title, err.Error()))
		} else {
			a.processingStatus[a.idx] = successStyle.Render(fmt.Sprintf("%s\t%s", successIcon, article.Title))
		}
		return nextArticleItem{
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
	return screenSize.Height - 2*screenPadding - 8
}

func Start(ctx context.Context, a *adapter.FeedAdapter, feedLink string, skipCache bool) error {
	m, err := newScreen(ctx, a, feedLink, skipCache)
	if err != nil {
		return fmt.Errorf("creating screen: %w", err)
	}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		return fmt.Errorf("running program: %w", err)
	}
	return nil
}
