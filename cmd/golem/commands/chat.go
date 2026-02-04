package commands

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/MEKXH/golem/internal/agent"
	"github.com/MEKXH/golem/internal/bus"
	"github.com/MEKXH/golem/internal/config"
	"github.com/MEKXH/golem/internal/provider"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

func NewChatCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "chat [message]",
		Short: "Chat with Golem",
		RunE:  runChat,
	}
}

type (
	errMsg error
)

type model struct {
	viewport      viewport.Model
	textarea      textarea.Model
	senderStyle   lipgloss.Style
	aiStyle       lipgloss.Style
	thinkingStyle lipgloss.Style
	toolStyle     lipgloss.Style
	renderer      *glamour.TermRenderer
	history       *strings.Builder
	err           error
	loop          *agent.Loop
	ctx           context.Context
}

func initialModel(ctx context.Context, loop *agent.Loop) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)

	history := &strings.Builder{}
	history.WriteString(`Welcome to Golem Chat!
Type a message and press Enter to send.`)
	vp.SetContent(history.String())

	ta.KeyMap.InsertNewline.SetEnabled(false)

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithStandardStyle("dark"),
		glamour.WithWordWrap(30),
	)

	return model{
		textarea:      ta,
		viewport:      vp,
		senderStyle:   lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		aiStyle:       lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		thinkingStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true),
		toolStyle:     lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
		renderer:      renderer,
		history:       history,
		loop:          loop,
		ctx:           ctx,
		err:           nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

type responseMsg string

type toolStartMsg struct {
	name string
	args string
}

type toolFinishMsg struct {
	name   string
	result string
	err    error
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		// Height - textarea height - 1 line for separation
		m.viewport.Height = msg.Height - m.textarea.Height() - 1
		m.textarea.SetWidth(msg.Width)

		// Update renderer width
		if m.renderer != nil {
			m.renderer, _ = glamour.NewTermRenderer(
				glamour.WithStandardStyle("dark"),
				glamour.WithWordWrap(msg.Width),
			)
		}

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			if m.textarea.Value() == "" {
				return m, nil
			}
			input := m.textarea.Value()
			m.textarea.Reset()

			m.history.WriteString("\n\n" + m.senderStyle.Render("You: ") + input)
			m.viewport.SetContent(m.history.String())
			m.viewport.GotoBottom()

			return m, func() tea.Msg {
				resp, err := m.loop.ProcessDirect(m.ctx, input)
				if err != nil {
					return errMsg(err)
				}
				return responseMsg(resp)
			}
		}

	case responseMsg:
		content := string(msg)
		re := regexp.MustCompile(`(?s)<think>(.*?)</think>`)
		matches := re.FindStringSubmatch(content)

		var viewContent string
		if len(matches) > 1 {
			thinkContent := strings.TrimSpace(matches[1])
			mainContent := strings.TrimSpace(re.ReplaceAllString(content, ""))

			// Indent thinking content slightly for better visual separation
			thinkContent = strings.ReplaceAll(thinkContent, "\n", "\n  ")

			viewContent = "\n\n" + m.thinkingStyle.Render("💭 Thinking:\n  "+thinkContent) +
				"\n\n" + m.aiStyle.Render("Golem: ") + mainContent
		} else {
			mainContent := content
			rendered, err := m.renderer.Render(mainContent)
			if err == nil {
				mainContent = rendered
			} else {
				mainContent += fmt.Sprintf("\n(Markdown render error: %v)", err)
			}
			viewContent = "\n\n" + m.aiStyle.Render("Golem: ") + mainContent
		}

		m.history.WriteString(viewContent)
		m.viewport.SetContent(m.history.String())
		m.viewport.GotoBottom()

	case toolStartMsg:
		content := fmt.Sprintf("🛠️  Executing tool: %s\n", msg.name)
		m.history.WriteString("\n" + m.toolStyle.Render(content))
		m.viewport.SetContent(m.history.String())
		m.viewport.GotoBottom()

	case toolFinishMsg:
		var content string
		if msg.err != nil {
			content = fmt.Sprintf("❌ Tool failed: %v", msg.err)
		} else {
			// Truncate result if too long
			result := msg.result
			if len(result) > 100 {
				result = result[:100] + "..."
			}
			content = fmt.Sprintf("✅ Tool finished: %s", result)
		}
		m.history.WriteString("\n" + m.toolStyle.Render(content))
		m.viewport.SetContent(m.history.String())
		m.viewport.GotoBottom()

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n%s",
		m.viewport.View(),
		m.textarea.View(),
	)
}

func runChat(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Disable logging for TUI
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	modelProvider, err := provider.NewChatModel(ctx, cfg)
	if err != nil {
		fmt.Printf("Warning: %v\nRunning without LLM (tools only mode)\n", err)
		modelProvider = nil
	}

	msgBus := bus.NewMessageBus(10)
	loop, err := agent.NewLoop(cfg, msgBus, modelProvider)
	if err != nil {
		return fmt.Errorf("invalid workspace: %w", err)
	}
	if err := loop.RegisterDefaultTools(cfg); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}

	if len(args) > 0 {
		message := strings.Join(args, " ")
		resp, err := loop.ProcessDirect(ctx, message)
		if err != nil {
			return err
		}
		fmt.Println(resp)
		return nil
	}

	p := tea.NewProgram(initialModel(ctx, loop), tea.WithAltScreen())

	// Set callbacks
	loop.OnToolStart = func(name, args string) {
		p.Send(toolStartMsg{name: name, args: args})
	}
	loop.OnToolFinish = func(name, result string, err error) {
		p.Send(toolFinishMsg{name: name, result: result, err: err})
	}

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}
