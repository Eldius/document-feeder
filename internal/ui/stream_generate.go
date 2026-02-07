package ui

import (
	"context"
	"fmt"
	"github.com/eldius/document-feeder/internal/adapter"
	"github.com/eldius/initial-config-go/logs"
	"golang.org/x/term"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Mensagem personalizada para texto recebido do canal
type textReceivedMsg string

// Mensagem para indicar que o canal foi fechado
type channelClosedMsg struct{}

// Modelo para a tela de streaming
type streamGenerateScreenModel struct {
	vp       viewport.Model
	content  strings.Builder
	ctx      context.Context
	textChan <-chan string
	ready    bool
}

func NewStreamModel(ctx context.Context, textChan <-chan string) tea.Model {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		fmt.Println("Failed getting terminal size:", err)
		return nil
	}
	fmt.Printf("Initial terminal size: %dx%d\n", width, height)
	vp := viewport.New(screenMaxUIWidth(width), screenMaxUIWidth(height))
	vp.SetContent("")

	return &streamGenerateScreenModel{
		vp:       vp,
		ctx:      ctx,
		textChan: textChan,
		ready:    false,
	}
}

func (m *streamGenerateScreenModel) Init() tea.Cmd {
	return tea.Batch(
		m.waitForText(),
		tea.EnterAltScreen,
	)
}

// Comando para aguardar texto do canal
func (m *streamGenerateScreenModel) waitForText() tea.Cmd {
	return func() tea.Msg {
		select {
		case text, ok := <-m.textChan:
			if !ok {
				return channelClosedMsg{}
			}
			logs.NewLogger(m.ctx, logs.KeyValueData{
				"chunk_size": len(text),
				"chunk_text": text,
			}).Debug("text received")
			return textReceivedMsg(text)
		case <-m.ctx.Done():
			return tea.Quit()
		}
	}
}

func (m *streamGenerateScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		if !m.ready {
			m.vp = viewport.New(screenMaxUIWidth(msg.Width), screenMaxUIWidth(msg.Height))
			m.vp.SetContent(m.content.String())
			m.ready = true
		} else {
			m.vp.Width = msg.Width - 2
			m.vp.Height = msg.Height - 4
		}

	case textReceivedMsg:
		m.content.WriteString(string(msg))
		m.vp.SetContent(m.content.String())
		m.vp.GotoBottom()
		return m, m.waitForText()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		default:
			m.vp, cmd = m.vp.Update(msg)
		}
	case channelClosedMsg:

	}

	return m, cmd
}

func (m *streamGenerateScreenModel) View() string {
	if !m.ready {
		return "Inicializando..."
	}

	// Estilo para o viewport
	viewportStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("62")).
		Padding(1, 2)

	// Estilo para o cabeçalho
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("205")).
		Bold(true).
		Align(lipgloss.Center)

	// Estilo para o rodapé
	footerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Align(lipgloss.Center)

	header := headerStyle.Render("Stream de Texto")
	vp := viewportStyle.Render(m.vp.View())
	footer := footerStyle.Render("Pressione 'q' ou 'Ctrl+C' para sair • Use ↑↓ para navegar")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		vp,
		footer,
	)
}

func RunStreamApp(ctx context.Context, question string) error {
	ch := make(chan string)
	var wg sync.WaitGroup
	wg.Go(func() {
		model := NewStreamModel(ctx, ch)
		p := tea.NewProgram(
			model,
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		_, err := p.Run()
		if err != nil {
			fmt.Println("Error running program:", err)
			wg.Done()
		}
	})

	a, err := adapter.NewDefaultAdapter()
	if err != nil {
		return fmt.Errorf("creating adapter: %w", err)
	}
	wg.Go(func() {
		if err := a.AskAQuestionStream(ctx, question, ch); err != nil {
			fmt.Println("Error asking question:", err)
			return
		}
	})
	wg.Wait()
	return nil
}
