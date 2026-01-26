package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// tickMsg is sent periodically to animate the dots.
type tickMsg time.Time

// doneMsg is sent when the context is cancelled.
type doneMsg struct{}

type processingScreenModel struct {
	ctx       context.Context
	dots      int
	width     int
	height    int
	quitting  bool
	msg       string
	startTime time.Time
}

// waitForContext monitors the context and sends a message when Done.
func waitForContext(ctx context.Context) tea.Cmd {
	return func() tea.Msg {
		<-ctx.Done()
		return doneMsg{}
	}
}

// tick returns a command that sends a message every 500ms.
func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m processingScreenModel) Init() tea.Cmd {
	return tea.Batch(tick(), waitForContext(m.ctx))
}

func (m processingScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tickMsg:
		m.dots = (m.dots + 1) % 4
		return m, tick()
	case doneMsg:
		m.quitting = true
		return m, tea.Quit
	}
	return m, nil
}

func (m processingScreenModel) View() string {
	if m.quitting {
		return "Done!\n"
	}

	// Get current time and calculate elapsed time
	now := time.Now()
	elapsed := now.Sub(m.startTime)

	// Format timestamp and elapsed time
	timestamp := now.Format("15:04:05")
	elapsedStr := fmt.Sprintf("%02d:%02d", int(elapsed.Minutes()), int(elapsed.Seconds())%60)

	// Create the processing string with dynamic dots, timestamp and elapsed time
	processStr := fmt.Sprintf("%s\nProcessing%s\n\nTimestamp: %s\nElapsed: %s",
		m.msg,
		strings.Repeat(".", m.dots),
		timestamp,
		elapsedStr)

	// Define the box style
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")). // Purple-ish
		Padding(1, 3).
		Bold(true)

	// Render the box
	box := style.Render(processStr)

	// Center the box in the terminal window
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		box,
	)
}

func displayProcessingScreen(ctx context.Context, msg string) error {
	p := tea.NewProgram(processingScreenModel{
		ctx:       ctx,
		msg:       msg,
		startTime: time.Now(),
	})
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v", err)
		return err
	}
	return nil
}

func ProcessingScreen(ctx context.Context, msg string) context.CancelFunc {
	ctx, cancel := context.WithCancel(ctx)

	go func(ctx context.Context) {
		if err := displayProcessingScreen(ctx, msg); err != nil {
			fmt.Println("error processing screen display:", err)
		}
	}(ctx)
	return func() {
		cancel()
	}
}
