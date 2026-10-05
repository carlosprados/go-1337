// Package tui is the live leet editor: type on one side, read the conversion
// on the other.
package tui

import (
	"math/rand/v2"
	"time"

	"github.com/carlosprados/go-1337/internal/anim"
	"github.com/carlosprados/go-1337/internal/leet"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CopyFunc puts text on the system clipboard.
type CopyFunc func(string) error

// Run starts the editor and blocks until the user quits.
func Run(a *leet.Alphabet, copyFn CopyFunc) error {
	_, err := tea.NewProgram(New(a, copyFn), tea.WithAltScreen()).Run()
	return err
}

type keyMap struct {
	Mode, Level, Random, Shuffle, Copy, Quit key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Mode, k.Level, k.Random, k.Shuffle, k.Copy, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding { return [][]key.Binding{k.ShortHelp()} }

var keys = keyMap{
	Mode:    key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "encode/decode")),
	Level:   key.NewBinding(key.WithKeys("ctrl+l"), key.WithHelp("ctrl+l", "level")),
	Random:  key.NewBinding(key.WithKeys("ctrl+r"), key.WithHelp("ctrl+r", "random")),
	Shuffle: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "reshuffle")),
	Copy:    key.NewBinding(key.WithKeys("ctrl+y"), key.WithHelp("ctrl+y", "copy output")),
	Quit:    key.NewBinding(key.WithKeys("esc", "ctrl+c"), key.WithHelp("esc", "quit")),
}

// Model is the Bubble Tea model of the editor.
type Model struct {
	alphabet *leet.Alphabet
	copyFn   CopyFunc
	input    textarea.Model
	help     help.Model
	decode   bool
	level    leet.Level
	random   bool
	seed     uint64
	status   string
	width    int
	height   int

	animStart time.Time // zero when no animation is running
	animSeed  uint64
}

type (
	clearStatusMsg struct{}
	animTickMsg    struct{}
)

const animDuration = 600 * time.Millisecond

func animTick() tea.Cmd {
	return tea.Tick(time.Second/30, func(time.Time) tea.Msg { return animTickMsg{} })
}

// startAnim replays the decrypting effect on the output panel.
func (m *Model) startAnim() tea.Cmd {
	m.animStart = time.Now()
	m.animSeed++
	return animTick()
}

// animProgress is 1 when idle, otherwise how far the animation has run.
func (m Model) animProgress() float64 {
	if m.animStart.IsZero() {
		return 1
	}
	return min(float64(time.Since(m.animStart))/float64(animDuration), 1)
}

// New builds the editor model.
func New(a *leet.Alphabet, copyFn CopyFunc) Model {
	ta := textarea.New()
	ta.Placeholder = "Type something..."
	ta.ShowLineNumbers = false
	ta.Prompt = ""
	ta.CharLimit = 0
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.Focus()
	return Model{
		alphabet: a,
		copyFn:   copyFn,
		input:    ta,
		help:     help.New(),
		level:    leet.Elite,
		seed:     uint64(time.Now().UnixNano()),
		width:    80,
		height:   24,
	}
}

func (m Model) Init() tea.Cmd { return textarea.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.help.Width = msg.Width
		m.input.SetWidth(m.panelWidth() - 2)
		m.input.SetHeight(m.panelHeight() - 1)
		return m, nil
	case clearStatusMsg:
		m.status = ""
		return m, nil
	case animTickMsg:
		if m.animProgress() < 1 {
			return m, animTick()
		}
		m.animStart = time.Time{}
		return m, nil
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, keys.Mode):
			m.decode = !m.decode
			return m, m.startAnim()
		case key.Matches(msg, keys.Level):
			m.level = (m.level + 1) % (leet.Elite + 1)
			return m, m.startAnim()
		case key.Matches(msg, keys.Random):
			m.random = !m.random
			return m, m.startAnim()
		case key.Matches(msg, keys.Shuffle):
			m.seed++
			return m, m.startAnim()
		case key.Matches(msg, keys.Copy):
			m.status = "copied to clipboard"
			if err := m.copyFn(m.Output()); err != nil {
				m.status = err.Error()
			}
			return m, tea.Tick(2*time.Second, func(time.Time) tea.Msg { return clearStatusMsg{} })
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// Output is the conversion of the current input.
func (m Model) Output() string {
	text := m.input.Value()
	if m.decode {
		return m.alphabet.Decode(text)
	}
	opts := leet.EncodeOptions{Level: m.level}
	if m.random {
		// A fixed seed per text keeps the output stable while typing.
		opts.Rand = rand.New(rand.NewPCG(m.seed, m.seed))
	}
	return m.alphabet.Encode(text, opts)
}

var (
	accent    = lipgloss.Color("10")
	dim       = lipgloss.Color("8")
	logoStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(accent).Padding(0, 1)
	pillOn    = lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(accent).Padding(0, 1)
	pillOff   = lipgloss.NewStyle().Foreground(dim).Padding(0, 1)
	panel     = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(dim).Padding(0, 1)
	active    = panel.BorderForeground(accent)
	titleOf   = lipgloss.NewStyle().Bold(true).Foreground(accent)
	outStyle  = lipgloss.NewStyle().Foreground(accent)
	statusOf  = lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Padding(0, 1)
)

func (m Model) sideBySide() bool { return m.width >= 90 }

// panelWidth and panelHeight are the inner sizes (padding included, border
// excluded) of each panel; the border adds 2 to each dimension.
func (m Model) panelWidth() int {
	if m.sideBySide() {
		return (m.width-1)/2 - 2
	}
	return m.width - 2
}

func (m Model) panelHeight() int {
	h := m.height - 2 // header and help lines
	if !m.sideBySide() {
		h /= 2
	}
	return max(h-2, 4)
}

func pill(label string, on bool) string {
	if on {
		return pillOn.Render(label)
	}
	return pillOff.Render(label)
}

func (m Model) View() string {
	inTitle, outTitle := "PLAIN", "1337"
	if m.decode {
		inTitle, outTitle = "1337", "PLAIN"
	}
	header := lipgloss.JoinHorizontal(lipgloss.Center,
		logoStyle.Render("1337"), " ",
		pill("encode", !m.decode), pill("decode", m.decode), "  ",
		pill(m.level.String(), !m.decode), " ", pill("random", m.random && !m.decode),
		statusOf.Render(m.status),
	)

	output := m.Output()
	if p := m.animProgress(); p < 1 {
		output = anim.Frame(output, p, m.animSeed)
	}

	w, h := m.panelWidth(), m.panelHeight()
	in := active.Width(w).Height(h).Render(titleOf.Render(inTitle) + "\n" + m.input.View())
	out := panel.Width(w).Height(h).Render(titleOf.Render(outTitle) + "\n" +
		outStyle.Width(w-2).MaxHeight(h-1).Render(output))

	var body string
	if m.sideBySide() {
		body = lipgloss.JoinHorizontal(lipgloss.Top, in, " ", out)
	} else {
		body = lipgloss.JoinVertical(lipgloss.Left, in, out)
	}
	return lipgloss.JoinVertical(lipgloss.Left, header, body, m.help.View(keys))
}
