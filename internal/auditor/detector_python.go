package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PythonDetector struct {
	lang Language
}

func NewPythonDetector() *PythonDetector {
	return &PythonDetector{lang: LangPython}
}

func (d *PythonDetector) Language() Language   { return d.lang }
func (d *PythonDetector) Extensions() []string { return []string{".py"} }

var (
	pythonOsEnviron    = regexp.MustCompile(`os\.environ(?:ment)?\[["']([A-Z_][A-Z0-9_]*)["']\]`)
	pythonOsGetenv     = regexp.MustCompile(`os\.getenv\(["']([A-Z_][A-Z0-9_]*)["']`)
	pythonOsEnvironGet = regexp.MustCompile(`os\.environ(?:ment)?\.get\(["']([A-Z_][A-Z0-9_]*)["']`)
	pythonOsEnvGet     = regexp.MustCompile(`os\.env\.get\(["']([A-Z_][A-Z0-9_]*)["']`)
)

func (d *PythonDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectPythonFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		for _, re := range []*regexp.Regexp{pythonOsEnviron, pythonOsGetenv, pythonOsEnvironGet, pythonOsEnvGet} {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangPython,
						Framework: framework,
						Method:    "os.environ/getenv",
					})
				}
			}
		}
	}

	return vars, LangPython, framework
}

func detectPythonFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "requirements.txt")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "requirements.txt"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "fastapi") {
				return FrameworkFastAPI
			}
			if strings.Contains(lower, "django") {
				return FrameworkDjango
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "pyproject.toml")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "fastapi") {
				return FrameworkFastAPI
			}
			if strings.Contains(lower, "django") {
				return FrameworkDjango
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "manage.py")); err == nil {
			return FrameworkDjango
		}
		if dir == "." {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
