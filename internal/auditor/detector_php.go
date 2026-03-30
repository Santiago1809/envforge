package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type PHPDetector struct {
	lang Language
}

func NewPHPDetector() *PHPDetector {
	return &PHPDetector{lang: LangPHP}
}

func (d *PHPDetector) Language() Language   { return d.lang }
func (d *PHPDetector) Extensions() []string { return []string{".php"} }

var (
	phpGetEnv        = regexp.MustCompile(`\$_ENV\[["']([A-Z_][A-Z0-9_]*)["']`)
	phpGetServer     = regexp.MustCompile(`\$_SERVER\[["']([A-Z_][A-Z0-9_]*)["']`)
	phpEnvFunction   = regexp.MustCompile(`env\(['"]([A-Z_][A-Z0-9_]*)['"]`)
	phpLaravelConfig = regexp.MustCompile(`config\(['"]([A-Za-z_][A-Za-z0-9_\.]*)['"]|config\([^)]*->get\(['"]([A-Za-z_][A-Za-z0-9_\.]*)['"]`)
)

func (d *PHPDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectPHPFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := phpGetEnv.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangPHP,
					Framework: framework,
					Method:    "$_ENV",
				})
			}
		}

		matches = phpGetServer.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangPHP,
					Framework: framework,
					Method:    "$_SERVER",
				})
			}
		}

		matches = phpEnvFunction.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangPHP,
					Framework: framework,
					Method:    "env()",
				})
			}
		}

		if framework == FrameworkLaravel {
			matches = phpLaravelConfig.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				key := match[1]
				if key == "" {
					key = match[2]
				}
				if key != "" && strings.Contains(key, ".") {
					parts := strings.Split(key, ".")
					for _, part := range parts {
						if strings.ToUpper(part) == part && strings.Contains(part, "_") {
							vars = append(vars, EnvUsage{
								Key:       part,
								File:      path,
								Lines:     []int{lineNum},
								Language:  LangPHP,
								Framework: FrameworkLaravel,
								Method:    "config()",
							})
						}
					}
				}
			}
		}
	}

	return vars, LangPHP, framework
}

func detectPHPFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "artisan")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "artisan"))
			if strings.Contains(string(content), "Laravel") || strings.Contains(string(content), "Illuminate") {
				return FrameworkLaravel
			}
			return FrameworkLaravel
		}
		if _, err := os.Stat(filepath.Join(dir, "composer.json")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "composer.json"))
			if strings.Contains(string(content), "laravel") {
				return FrameworkLaravel
			}
			if strings.Contains(string(content), "symfony") {
				return FrameworkDotNet
			}
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
