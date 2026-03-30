package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type JavaScriptDetector struct {
	lang Language
}

func NewJavaScriptDetector() *JavaScriptDetector {
	return &JavaScriptDetector{lang: LangJS}
}

func (d *JavaScriptDetector) Language() Language   { return d.lang }
func (d *JavaScriptDetector) Extensions() []string { return []string{".js", ".jsx", ".mjs", ".cjs"} }

var (
	jsProcessEnvDot     = regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`)
	jsDotEnv            = regexp.MustCompile(`\.env\.([A-Z_][A-Z0-9_]*)`)
	jsProcessEnvBracket = regexp.MustCompile(`process\.env\[["']([A-Z_][A-Z0-9_]*)["']\]`)
	jsBunEnv            = regexp.MustCompile(`Bun\.env\.([A-Z_][A-Z0-9_]*)`)
	jsImportMetaEnv     = regexp.MustCompile(`import\.meta\.env\.([A-Z_][A-Z0-9_]*)`)
)

func (d *JavaScriptDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectJSFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := jsBunEnv.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangJS,
					Framework: FrameworkBun,
					Method:    "Bun.env",
				})
			}
		}

		for _, re := range []*regexp.Regexp{jsProcessEnvDot, jsDotEnv, jsProcessEnvBracket, jsImportMetaEnv} {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 && match[1] != "" {
					method := "process.env"
					if re == jsImportMetaEnv {
						method = "import.meta.env"
					}
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangJS,
						Framework: framework,
						Method:    method,
					})
				}
			}
		}
	}

	if framework == "" {
		framework = detectBunRuntime(path)
	}

	return vars, LangJS, framework
}

func detectBunRuntime(path string) Framework {
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, "bun.lockb")); err == nil {
			return FrameworkBun
		}
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "package.json"))
			if strings.Contains(string(content), "bun") {
				return FrameworkBun
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

func detectJSFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "package.json"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "bun") {
				return FrameworkBun
			}
			if strings.Contains(lower, "next") {
				return FrameworkNextJS
			}
			if strings.Contains(lower, "nuxt") {
				return FrameworkVue
			}
			if strings.Contains(lower, "vue") {
				return FrameworkVue
			}
			if strings.Contains(lower, "angular") {
				return FrameworkAngular
			}
			if strings.Contains(lower, "react") {
				return FrameworkReact
			}
			if strings.Contains(lower, "express") || strings.Contains(lower, "fastify") || strings.Contains(lower, "nest") {
				return FrameworkNode
			}
		}
		if _, err := os.Stat(filepath.Join(dir, "nuxt.config.js")); err == nil {
			return FrameworkVue
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
