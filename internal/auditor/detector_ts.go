package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type TypeScriptDetector struct {
	lang Language
}

func NewTypeScriptDetector() *TypeScriptDetector {
	return &TypeScriptDetector{lang: LangTS}
}

func (d *TypeScriptDetector) Language() Language   { return d.lang }
func (d *TypeScriptDetector) Extensions() []string { return []string{".ts", ".tsx"} }

var (
	tsProcessEnvDot     = regexp.MustCompile(`process\.env\.([A-Z_][A-Z0-9_]*)`)
	tsDotEnv            = regexp.MustCompile(`\.env\.([A-Z_][A-Z0-9_]*)`)
	tsProcessEnvBracket = regexp.MustCompile(`process\.env\[["']([A-Z_][A-Z0-9_]*)["']\]`)
	tsImportMetaEnv     = regexp.MustCompile(`import\.meta\.env\.([A-Z_][A-Z0-9_]*)`)
	tsDefineConfig      = regexp.MustCompile(`defineConfig\([^)]*env\.([A-Z_][A-Z0-9_]*)`)
	tsNextPublic        = regexp.MustCompile(`process\.env\.NEXT_PUBLIC_([A-Z_][A-Z0-9_]*)`)
	tsNgForRoot         = regexp.MustCompile(`InjectionToken.*?ENV.*?forRoot`)
	tsBunEnv            = regexp.MustCompile(`Bun\.env\.([A-Z_][A-Z0-9_]*)`)
)

func (d *TypeScriptDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectTSFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := tsBunEnv.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangTS,
					Framework: FrameworkBun,
					Method:    "Bun.env",
				})
			}
		}

		for _, re := range []*regexp.Regexp{tsProcessEnvDot, tsDotEnv, tsProcessEnvBracket, tsImportMetaEnv, tsDefineConfig} {
			matches := re.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				key := match[1]
				if key == "" || strings.HasPrefix(key, "NEXT_PUBLIC_") {
					continue
				}
				method := "process.env"
				if re == tsImportMetaEnv {
					method = "import.meta.env"
				} else if re == tsDefineConfig {
					method = "defineConfig"
				}
				vars = append(vars, EnvUsage{
					Key:       key,
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangTS,
					Framework: framework,
					Method:    method,
				})
			}
		}

		matches = tsNextPublic.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 && match[1] != "" {
				vars = append(vars, EnvUsage{
					Key:       "NEXT_PUBLIC_" + match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangTS,
					Framework: FrameworkNextJS,
					Method:    "process.env.NEXT_PUBLIC",
				})
			}
		}
	}

	if framework == FrameworkAngular {
		for i, line := range lines {
			matches := tsNgForRoot.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 0 {
					vars = append(vars, EnvUsage{
						Key:       "ENV",
						File:      path,
						Lines:     []int{i + 1},
						Language:  LangTS,
						Framework: FrameworkAngular,
						Method:    "InjectionToken ENV forRoot",
					})
				}
			}
		}
	}

	if framework == "" {
		framework = detectBunInTS(path)
	}

	return vars, LangTS, framework
}

func detectTSFramework(path string, dir string) Framework {
	for {
		if _, err := os.Stat(filepath.Join(dir, "package.json")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "package.json"))
			lower := strings.ToLower(string(content))
			if strings.Contains(lower, "next") {
				return FrameworkNextJS
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
		}
		if _, err := os.Stat(filepath.Join(dir, "nuxt.config.ts")); err == nil {
			return FrameworkVue
		}
		if _, err := os.Stat(filepath.Join(dir, "svelte.config.js")); err == nil {
			return FrameworkReact
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

func detectBunInTS(path string) Framework {
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
