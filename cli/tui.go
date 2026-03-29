package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/Santiago1809/envforge/internal/auditor"
	"github.com/Santiago1809/envforge/internal/parser"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	tabBorder   = lipgloss.RoundedBorder()
	inactiveTab = lipgloss.Style{}.
			Foreground(lipgloss.Color("8")).
			Border(tabBorder)
	activeTab = lipgloss.Style{}.
			Foreground(lipgloss.Color("15")).
			Bold(true).
			Border(tabBorder)
	tabContent = lipgloss.Style{}.
			Foreground(lipgloss.Color("15"))
	helpStyle = lipgloss.Style{}.
			Foreground(lipgloss.Color("8"))
	greenStyle  = lipgloss.Style{}.Foreground(lipgloss.Color("10"))
	redStyle    = lipgloss.Style{}.Foreground(lipgloss.Color("9"))
	yellowStyle = lipgloss.Style{}.Foreground(lipgloss.Color("11"))
	cyanStyle   = lipgloss.Style{}.Foreground(lipgloss.Color("14"))
	boxStyle    = lipgloss.Style{}.
			Border(lipgloss.RoundedBorder()).
			Padding(1)
)

type model struct {
	currentTab int
	scroll     int

	envKeys     []string
	exampleKeys []string
	missingKeys []string
	extraKeys   []string

	auditUsedNotDeclared []auditEntry
	auditDeclaredNotUsed []string

	healthKeys     []string
	healthMissing  []string
	fileSize       int64
	exampleSize    int64
	lastModified   time.Time
	lastExampleMod time.Time

	loaded bool
	err    error
}

type auditEntry struct {
	key   string
	file  string
	lines string
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.currentTab = (m.currentTab + 1) % 3
			m.scroll = 0
		case "shift+tab":
			m.currentTab = (m.currentTab - 1 + 3) % 3
			m.scroll = 0
		case "up", "k":
			if m.scroll > 0 {
				m.scroll--
			}
		case "down", "j":
			maxScroll := m.maxScroll()
			if m.scroll < maxScroll {
				m.scroll++
			}
		}
	}
	return m, nil
}

func (m model) maxScroll() int {
	switch m.currentTab {
	case 0:
		return max(0, len(m.envKeys)+len(m.missingKeys)+len(m.extraKeys)-10)
	case 1:
		return max(0, len(m.auditUsedNotDeclared)+len(m.auditDeclaredNotUsed)-10)
	case 2:
		return max(0, len(m.healthKeys)-10)
	}
	return 0
}

func (m model) View() string {
	if m.err != nil {
		return redStyle.Render(fmt.Sprintf("Error: %v\n", m.err))
	}

	if !m.loaded {
		return "Loading..."
	}

	tabs := []string{" Overview ", " Audit ", " Health "}
	for i, t := range tabs {
		if i == m.currentTab {
			tabs[i] = activeTab.Render(t)
		} else {
			tabs[i] = inactiveTab.Render(t)
		}
	}

	header := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)

	var content string
	switch m.currentTab {
	case 0:
		content = m.overviewView()
	case 1:
		content = m.auditView()
	case 2:
		content = m.healthView()
	}

	help := helpStyle.Render(" Tab/Shift+Tab: switch | ↑↓/jk: scroll | q/Ctrl+C: quit ")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		tabContent.Render(content),
		help,
	)
}

func (m model) overviewView() string {
	var lines []string

	if len(m.missingKeys) > 0 {
		lines = append(lines, redStyle.Render("MISSING from .env (in .env.example):"))
		for _, k := range m.missingKeys {
			lines = append(lines, "  "+redStyle.Render("• "+k))
		}
		lines = append(lines, "")
	}

	if len(m.extraKeys) > 0 {
		lines = append(lines, yellowStyle.Render("EXTRA in .env (not in .env.example):"))
		for _, k := range m.extraKeys {
			lines = append(lines, "  "+yellowStyle.Render("• "+k))
		}
		lines = append(lines, "")
	}

	lines = append(lines, greenStyle.Render("PRESENT in both:"))
	for _, k := range m.envKeys {
		lines = append(lines, "  "+greenStyle.Render("• "+k+" = ****"))
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) auditView() string {
	var lines []string

	if len(m.auditUsedNotDeclared) > 0 {
		lines = append(lines, redStyle.Render("USED but NOT DECLARED:"))
		for _, e := range m.auditUsedNotDeclared {
			lines = append(lines, "  "+redStyle.Render("• "+e.key+" ("+e.file+":"+e.lines+")"))
		}
		lines = append(lines, "")
	}

	if len(m.auditDeclaredNotUsed) > 0 {
		lines = append(lines, yellowStyle.Render("DECLARED but NOT USED:"))
		for _, k := range m.auditDeclaredNotUsed {
			lines = append(lines, "  "+yellowStyle.Render("• "+k))
		}
	}

	if len(lines) == 0 {
		lines = append(lines, greenStyle.Render("No issues found!"))
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func (m model) healthView() string {
	var lines []string

	lines = append(lines, cyanStyle.Render("Required Variables:"))
	for _, k := range m.healthKeys {
		isMissing := false
		for _, m := range m.healthMissing {
			if m == k {
				isMissing = true
				break
			}
		}
		if isMissing {
			lines = append(lines, "  "+redStyle.Render("✗ "+k))
		} else {
			lines = append(lines, "  "+greenStyle.Render("✓ "+k))
		}
	}

	lines = append(lines, "")
	lines = append(lines, cyanStyle.Render("File Info:"))
	lines = append(lines, fmt.Sprintf("  .env: %d keys, %d bytes", len(m.envKeys), m.fileSize))
	lines = append(lines, fmt.Sprintf("  .env.example: %d keys, %d bytes", len(m.exampleKeys), m.exampleSize))

	if !m.lastModified.IsZero() {
		lines = append(lines, fmt.Sprintf("  .env modified: %s", m.lastModified.Format("2006-01-02 15:04")))
	}
	if !m.lastExampleMod.IsZero() {
		lines = append(lines, fmt.Sprintf("  .env.example modified: %s", m.lastExampleMod.Format("2006-01-02 15:04")))
	}

	return boxStyle.Render(lipgloss.JoinVertical(lipgloss.Left, lines...))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func RunTUI() error {
	m := model{}

	env, err := parser.Load(".env")
	if err != nil && !os.IsNotExist(err) {
		m.err = fmt.Errorf("failed to load .env: %w", err)
	} else if err == nil {
		m.envKeys = env.Keys()
		if stat, err := os.Stat(".env"); err == nil {
			m.fileSize = stat.Size()
			m.lastModified = stat.ModTime()
		}
	}

	example, err := parser.Load(".env.example")
	if err != nil && !os.IsNotExist(err) {
		m.err = fmt.Errorf("failed to load .env.example: %w", err)
	} else if err == nil {
		m.exampleKeys = example.Keys()
		if stat, err := os.Stat(".env.example"); err == nil {
			m.exampleSize = stat.Size()
			m.lastExampleMod = stat.ModTime()
		}
	}

	if len(m.exampleKeys) > 0 {
		envKeySet := make(map[string]bool)
		for _, k := range m.envKeys {
			envKeySet[k] = true
		}
		for _, k := range m.exampleKeys {
			if !envKeySet[k] {
				m.missingKeys = append(m.missingKeys, k)
			}
		}
	}

	if len(m.envKeys) > 0 {
		exampleKeySet := make(map[string]bool)
		for _, k := range m.exampleKeys {
			exampleKeySet[k] = true
		}
		for _, k := range m.envKeys {
			if !exampleKeySet[k] {
				m.extraKeys = append(m.extraKeys, k)
			}
		}
	}

	result, err := auditor.AuditDir(".", "", []auditor.Language{auditor.LangAll}, nil, false)
	if err == nil {
		for _, u := range result.UsedNotDeclared {
			m.auditUsedNotDeclared = append(m.auditUsedNotDeclared, auditEntry{
				key:   u.Key,
				file:  u.File,
				lines: fmt.Sprintf("%v", u.Lines),
			})
		}
		m.auditDeclaredNotUsed = result.DeclaredNotUsed
	}

	m.healthKeys = m.exampleKeys
	m.healthMissing = m.missingKeys

	m.loaded = true

	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}
	return nil
}
