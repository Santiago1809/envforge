package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type JavaDetector struct {
	lang Language
}

func NewJavaDetector() *JavaDetector {
	return &JavaDetector{lang: LangJava}
}

func (d *JavaDetector) Language() Language   { return d.lang }
func (d *JavaDetector) Extensions() []string { return []string{".java"} }

var (
	javaSystemGetEnv      = regexp.MustCompile(`System\.getenv\(["']([A-Z_][A-Z0-9_]*)["']`)
	javaSystemGetProperty = regexp.MustCompile(`System\.getProperty\(["']([A-Z_][A-Z0-9_]*)["']`)
	javaEnvAnnotation     = regexp.MustCompile(`@Value\(["']([A-Z_][A-Z0-9_]*)["']`)
	javaEnvVariable       = regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)
)

func (d *JavaDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectJavaFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := javaSystemGetEnv.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangJava,
					Framework: framework,
					Method:    "System.getenv",
				})
			}
		}

		matches = javaSystemGetProperty.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangJava,
					Framework: framework,
					Method:    "System.getProperty",
				})
			}
		}

		matches = javaEnvAnnotation.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangJava,
					Framework: framework,
					Method:    "@Value annotation",
				})
			}
		}

		matches = javaEnvVariable.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangJava,
					Framework: framework,
					Method:    "${} placeholder",
				})
			}
		}
	}

	return vars, LangJava, framework
}

func detectJavaFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "pom.xml")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "pom.xml"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "spring-boot") {
				return FrameworkSpringBoot
			}
			return FrameworkNode
		}
		if _, err := os.Stat(filepath.Join(dir, "build.gradle")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "build.gradle"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "spring.boot") {
				return FrameworkSpringBoot
			}
			return FrameworkNode
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
