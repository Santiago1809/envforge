package auditor

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type CSharpDetector struct {
	lang Language
}

func NewCSharpDetector() *CSharpDetector {
	return &CSharpDetector{lang: LangCSharp}
}

func (d *CSharpDetector) Language() Language   { return d.lang }
func (d *CSharpDetector) Extensions() []string { return []string{".cs"} }

var (
	csEnvGetEnvironmentVariable = regexp.MustCompile(`Environment\.GetEnvironmentVariable\(["']([A-Z_][A-Z0-9_]*)["']`)
	csEnvBracket                = regexp.MustCompile(`Environment\[["']([A-Z_][A-Z0-9_]*)["']\]`)
	csIConfiguration            = regexp.MustCompile(`_configuration\[["']([A-Za-z_][A-Za-z0-9_:]*)["']\]`)
	csIOptions                  = regexp.MustCompile(`IOptions<.*?>\).*?Value\.([A-Za-z_][A-Za-z0-9_]*)`)
)

func (d *CSharpDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage
	lines := strings.Split(content, "\n")
	framework := detectCSharpFramework(path, filepath.Dir(path))

	for i, line := range lines {
		lineNum := i + 1

		matches := csEnvGetEnvironmentVariable.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangCSharp,
					Framework: framework,
					Method:    "Environment.GetEnvironmentVariable",
				})
			}
		}

		matches = csEnvBracket.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) > 1 {
				vars = append(vars, EnvUsage{
					Key:       match[1],
					File:      path,
					Lines:     []int{lineNum},
					Language:  LangCSharp,
					Framework: framework,
					Method:    "Environment[]",
				})
			}
		}

		if framework == FrameworkDotNet {
			matches = csIConfiguration.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangCSharp,
						Framework: FrameworkDotNet,
						Method:    "IConfiguration[]",
					})
				}
			}

			matches = csIOptions.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					vars = append(vars, EnvUsage{
						Key:       match[1],
						File:      path,
						Lines:     []int{lineNum},
						Language:  LangCSharp,
						Framework: FrameworkDotNet,
						Method:    "IOptions.Value",
					})
				}
			}
		}
	}

	return vars, LangCSharp, framework
}

func detectCSharpFramework(path string, dir string) Framework {
	for {
		files, err := filepath.Glob(filepath.Join(dir, "*.csproj"))
		if err == nil && len(files) > 0 {
			if len(files) > 0 {
				content, _ := os.ReadFile(files[0])
				lower := strings.ToLower(string(content))
				if strings.Contains(lower, "webapi") || strings.Contains(lower, "mvc") || strings.Contains(lower, "razor") {
					return FrameworkDotNet
				}
			}
			return FrameworkDotNet
		}
		if _, err := os.Stat(filepath.Join(dir, "*.sln")); err == nil {
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
