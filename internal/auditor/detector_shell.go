package auditor

import (
	"regexp"
	"strings"
)

type ShellDetector struct {
	lang Language
}

func NewShellDetector() *ShellDetector {
	return &ShellDetector{lang: LangShell}
}

func (d *ShellDetector) Language() Language   { return d.lang }
func (d *ShellDetector) Extensions() []string { return []string{".sh", ".bash", ".zsh"} }

var shellVarPattern = regexp.MustCompile(`\$([A-Z_][A-Z0-9_]*)|\$\{([A-Z_][A-Z0-9_]*)\}`)

func (d *ShellDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		matches := shellVarPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			key := match[1]
			if key == "" {
				key = match[2]
			}
			if key != "" {
				vars = append(vars, EnvUsage{
					Key:       key,
					File:      path,
					Lines:     []int{i + 1},
					Language:  LangShell,
					Framework: "",
					Method:    "$VAR or ${VAR}",
				})
			}
		}
	}

	return vars, LangShell, ""
}
