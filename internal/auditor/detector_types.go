package auditor

import "fmt"

type Language string

const (
	LangGo     Language = "go"
	LangJS     Language = "js"
	LangTS     Language = "ts"
	LangPython Language = "py"
	LangShell  Language = "sh"
	LangJava   Language = "java"
	LangPHP    Language = "php"
	LangRuby   Language = "ruby"
	LangCSharp Language = "cs"
	LangAll    Language = "all"
)

type Framework string

const (
	FrameworkNextJS     Framework = "nextjs"
	FrameworkVue        Framework = "vue"
	FrameworkAngular    Framework = "angular"
	FrameworkReact      Framework = "react"
	FrameworkSpringBoot Framework = "springboot"
	FrameworkLaravel    Framework = "laravel"
	FrameworkRails      Framework = "rails"
	FrameworkDotNet     Framework = "dotnet"
	FrameworkBun        Framework = "bun"
	FrameworkNode       Framework = "node"
	FrameworkFastAPI    Framework = "fastapi"
	FrameworkDjango     Framework = "django"
)

type EnvUsage struct {
	Key       string
	File      string
	Lines     []int
	Language  Language
	Framework Framework
	Method    string
}

type SemanticDetector interface {
	Detect(path string, content string) ([]EnvUsage, Language, Framework)
	Extensions() []string
	Language() Language
}

func GetAllDetectors() []SemanticDetector {
	return []SemanticDetector{
		NewGoDetector(),
		NewTypeScriptDetector(),
		NewJavaScriptDetector(),
		NewJavaDetector(),
		NewPHPDetector(),
		NewRubyDetector(),
		NewCSharpDetector(),
		NewPythonDetector(),
		NewShellDetector(),
	}
}

func GetDetectorForFile(path string) (SemanticDetector, error) {
	ext := getExtension(path)

	detectors := GetAllDetectors()
	for _, d := range detectors {
		for _, e := range d.Extensions() {
			if ext == e {
				return d, nil
			}
		}
	}

	return nil, fmt.Errorf("no detector found for file: %s", path)
}

func getExtension(path string) string {
	ext := ""
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			ext = path[i:]
			break
		}
		if path[i] == '/' || path[i] == '\\' {
			break
		}
	}
	return ext
}
