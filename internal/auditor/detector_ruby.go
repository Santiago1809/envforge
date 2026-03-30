package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type RubyDetector struct {
	lang Language
}

func NewRubyDetector() *RubyDetector {
	return &RubyDetector{lang: LangRuby}
}

func (d *RubyDetector) Language() Language   { return d.lang }
func (d *RubyDetector) Extensions() []string { return []string{".rb"} }

var (
	rubyEnvFetch   = regexp.MustCompile(`ENV\.fetch\(['"]([A-Z_][A-Z0-9_]*)['"]`)
	rubyEnvBracket = regexp.MustCompile(`ENV\[["']([A-Z_][A-Z0-9_]*)["']\]`)
	rubyEnvDig     = regexp.MustCompile(`ENV\.dig\(["']([A-Z_][A-Z0-9_]*)['"]`)
	rubyRailsEnv   = regexp.MustCompile(`Rails\.application\.credentials\.[a-z_]+\[["']([A-Z_][A-Z0-9_]*)["']`)
	rubyRailsEnv2  = regexp.MustCompile(`Rails\.env\[["']([A-Z_][A-Z0-9_]*)["']`)
)

func (d *RubyDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectRubyFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := rubyEnvFetch.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangRuby,
					Framework: framework,
					Method:    "ENV.fetch",
				})
			}
		}

		matches = rubyEnvBracket.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangRuby,
					Framework: framework,
					Method:    "ENV[]",
				})
			}
		}

		matches = rubyEnvDig.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangRuby,
					Framework: framework,
					Method:    "ENV.dig",
				})
			}
		}

		if framework == FrameworkRails {
			matches = rubyRailsEnv.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangRuby,
						Framework: FrameworkRails,
						Method:    "Rails.credentials",
					})
				}
			}

			matches = rubyRailsEnv2.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangRuby,
						Framework: FrameworkRails,
						Method:    "Rails.env",
					})
				}
			}
		}
	}

	return vars, LangRuby, framework
}

func detectRubyFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "Gemfile")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "Gemfile"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "rails") {
				return FrameworkRails
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "config.ru")); err == nil {
			return FrameworkDotNet
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
