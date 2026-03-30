package schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor = lipgloss.Color("14")
	grayColor    = lipgloss.Color("8")
	greenColor   = lipgloss.Color("10")
	whiteColor   = lipgloss.Color("15")

	titleStyle = lipgloss.Style{}.
			Foreground(primaryColor).
			Bold(true)
	subtitleStyle = lipgloss.Style{}.
			Foreground(grayColor)
	headerStyle = lipgloss.Style{}.
			Foreground(grayColor)
	helpStyle = lipgloss.Style{}.
			Foreground(grayColor)

	selectedStyle = lipgloss.Style{}.
			Background(primaryColor).
			Foreground(whiteColor)
	modifiedTypeStyle = lipgloss.Style{}.
				Foreground(greenColor).
				Bold(true)
	normalTypeStyle = lipgloss.Style{}.
			Foreground(whiteColor)
	grayStyle = lipgloss.Style{}.
			Foreground(grayColor)
)

var editableTypes = []VarType{
	TypeString,
	TypeInt,
	TypeFloat,
	TypeBool,
	TypeURL,
	TypeEmail,
}

type SchemaEntry struct {
	Name         string
	InferredType VarType
	CurrentType  VarType
	Value        string
	IsModified   bool
}

type SchemaEditorModel struct {
	entries []SchemaEntry
	cursor  int
	saved   bool
	schema  *Schema
}

func (m SchemaEditorModel) Init() tea.Cmd {
	return nil
}

func (m SchemaEditorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			return m, tea.Quit
		case "s":
			m.saved = true
			for _, e := range m.entries {
				if m.schema.Vars == nil {
					m.schema.Vars = make(map[string]VarSchema)
				}
				m.schema.Vars[e.Name] = VarSchema{
					Name: e.Name,
					Type: e.CurrentType,
				}
			}
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
		case "left":
			if len(m.entries) > 0 {
				current := &m.entries[m.cursor]
				currentIdx := typeIndex(current.CurrentType)
				currentIdx = (currentIdx - 1 + len(editableTypes)) % len(editableTypes)
				current.CurrentType = editableTypes[currentIdx]
				current.IsModified = current.CurrentType != current.InferredType
			}
		case "right":
			if len(m.entries) > 0 {
				current := &m.entries[m.cursor]
				currentIdx := typeIndex(current.CurrentType)
				currentIdx = (currentIdx + 1) % len(editableTypes)
				current.CurrentType = editableTypes[currentIdx]
				current.IsModified = current.CurrentType != current.InferredType
			}
		}
	}
	return m, nil
}

func typeIndex(vt VarType) int {
	for i, t := range editableTypes {
		if t == vt {
			return i
		}
	}
	return 0
}

func (m SchemaEditorModel) View() string {
	var lines []string

	lines = append(lines, titleStyle.Render("  envforge — schema editor"))
	lines = append(lines, subtitleStyle.Render("  Inferred types from your .env — use ←/→ to change, 's' to save"))
	lines = append(lines, "")

	header := fmt.Sprintf("  %-28s | %-10s | %-10s | %-20s",
		"VARIABLE", "INFERRED", "TYPE", "SAMPLE VALUE")
	lines = append(lines, headerStyle.Render(header))
	lines = append(lines, headerStyle.Render(strings.Repeat("─", 78)))

	for i, e := range m.entries {
		displayValue := e.Value
		if len(displayValue) > 20 {
			displayValue = displayValue[:17] + "..."
		}

		typeDisplay := string(e.CurrentType)
		if e.IsModified {
			typeDisplay = modifiedTypeStyle.Render(typeDisplay)
		} else {
			typeDisplay = normalTypeStyle.Render(typeDisplay)
		}

		row := fmt.Sprintf("  %-28s | %-10s | %-10s | %-20s",
			e.Name,
			grayStyle.Render(string(e.InferredType)),
			typeDisplay,
			grayStyle.Render(displayValue),
		)

		if i == m.cursor {
			lines = append(lines, selectedStyle.Render(row))
		} else {
			lines = append(lines, row)
		}
	}

	lines = append(lines, "")
	lines = append(lines, helpStyle.Render("  [j/k] navigate   [←/→] change type   [s] save   [q] quit"))

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func RunSchemaEditor(schema *Schema, envVars map[string]string, schemaPath string) (*Schema, error) {
	entries := make([]SchemaEntry, 0)

	keys := make([]string, 0)
	for name := range envVars {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		value := envVars[name]
		inferredType := TypeString

		if schema != nil && schema.Vars != nil {
			if vs, ok := schema.Vars[name]; ok {
				inferredType = vs.Type
			}
		}

		entries = append(entries, SchemaEntry{
			Name:         name,
			InferredType: inferredType,
			CurrentType:  inferredType,
			Value:        value,
			IsModified:   false,
		})
	}

	model := SchemaEditorModel{
		entries: entries,
		cursor:  0,
		saved:   false,
		schema:  &Schema{Vars: make(map[string]VarSchema)},
	}

	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	result := finalModel.(SchemaEditorModel)

	if result.saved {
		err = result.schema.WriteTo(schemaPath)
		if err != nil {
			return nil, err
		}
		return result.schema, nil
	}

	return nil, nil
}
