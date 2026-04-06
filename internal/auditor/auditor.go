package auditor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/Santiago1809/envforge/internal/audittypes"
)

type AuditResult struct {
	UsedNotDeclared []EnvUsage
	DeclaredNotUsed []string
	DeclaredAndUsed []string
	Directory       string
	Language        string
}

type Auditor struct {
	rootDir   string
	languages []Language
	exclude   []string
	envFile   string
	results   []EnvUsage
	mu        sync.Mutex
	declared  map[string]bool
	wg        sync.WaitGroup
	ctx       context.Context
}

var DefaultExclusions = []string{
	"testdata", "vendor", "node_modules", ".git", "dist", "build", "bin",
	".agents", ".claude", ".skills", "skills",
}

func New(rootDir string) *Auditor {
	return &Auditor{
		rootDir:   rootDir,
		languages: []Language{LangGo, LangJS, LangTS, LangPython, LangShell, LangJava, LangPHP, LangRuby, LangCSharp},
		exclude:   DefaultExclusions,
		declared:  make(map[string]bool, 50), // typical env vars
	}
}

func (a *Auditor) SetLanguages(langs []Language) {
	a.languages = langs
}

func (a *Auditor) SetExclude(exclude []string) {
	seen := make(map[string]bool, len(a.exclude))
	for _, e := range a.exclude {
		seen[e] = true
	}
	for _, e := range exclude {
		if !seen[e] {
			a.exclude = append(a.exclude, e)
		}
	}
}

func (a *Auditor) SetEnvFile(envFile string) {
	a.envFile = envFile
}

func (a *Auditor) SetContext(ctx context.Context) {
	a.ctx = ctx
}

func (a *Auditor) Run() (*AuditResult, error) {
	if a.envFile != "" {
		if err := a.loadDeclaredVars(); err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
	}

	files, err := a.collectFiles()
	if err != nil {
		return nil, err
	}

	a.wg.Add(len(files))
	for _, file := range files {
		go a.auditFile(file)
	}

	a.wg.Wait()

	return a.buildResult()
}

func (a *Auditor) loadDeclaredVars() error {
	data, err := os.ReadFile(a.envFile)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "="); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			if key != "" {
				a.declared[key] = true
			}
		}
	}
	return nil
}

func (a *Auditor) collectFiles() ([]string, error) {
	files := make([]string, 0, 100) // preallocate with capacity hint

	err := filepath.WalkDir(a.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if a.ctx != nil && a.ctx.Err() != nil {
			return a.ctx.Err()
		}

		if d.IsDir() {
			name := d.Name()
			for _, ex := range a.exclude {
				if name == ex {
					return filepath.SkipDir
				}
			}
			return nil
		}

		detector, err := GetDetectorForFile(path)
		if err != nil {
			return nil
		}

		lang := detector.Language()
		for _, l := range a.languages {
			if l == LangAll || l == lang {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	return files, err
}

func (a *Auditor) auditFile(path string) {
	defer a.wg.Done()

	if a.ctx != nil && a.ctx.Err() != nil {
		return
	}

	detector, err := GetDetectorForFile(path)
	if err != nil {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	vars, _, _ := detector.Detect(path, string(data))

	a.mu.Lock()
	defer a.mu.Unlock()
	a.results = append(a.results, vars...)
}

func (a *Auditor) buildResult() (*AuditResult, error) {
	result := &AuditResult{
		UsedNotDeclared: []EnvUsage{},
		DeclaredNotUsed: []string{},
		DeclaredAndUsed: []string{},
	}

	used := make(map[string]bool)
	for _, r := range a.results {
		used[r.Key] = true
	}

	for key := range a.declared {
		if !used[key] {
			result.DeclaredNotUsed = append(result.DeclaredNotUsed, key)
		} else {
			result.DeclaredAndUsed = append(result.DeclaredAndUsed, key)
		}
	}

	keyToUsage := make(map[string]EnvUsage, len(a.results))
	for _, r := range a.results {
		if !a.declared[r.Key] {
			mapKey := r.Key + "|" + r.File
			if existing, ok := keyToUsage[mapKey]; ok {
				seen := make(map[int]bool)
				for _, line := range existing.Lines {
					seen[line] = true
				}
				for _, line := range r.Lines {
					if !seen[line] {
						existing.Lines = append(existing.Lines, line)
						seen[line] = true
					}
				}
				keyToUsage[mapKey] = existing
			} else {
				keyToUsage[mapKey] = r
			}
		}
	}

	for _, r := range keyToUsage {
		result.UsedNotDeclared = append(result.UsedNotDeclared, r)
	}

	result.Directory = a.rootDir
	if len(a.languages) == 0 {
		result.Language = ""
	} else {
		langStrs := make([]string, len(a.languages))
		for i, l := range a.languages {
			langStrs[i] = string(l)
		}
		result.Language = strings.Join(langStrs, ",")
	}

	return result, nil
}

func AuditDir(rootDir string, envFile string, languages []Language, exclude []string, verbose bool) (*AuditResult, error) {
	auditor := New(rootDir)
	if envFile != "" {
		auditor.SetEnvFile(envFile)
	}
	if len(languages) > 0 {
		auditor.SetLanguages(languages)
	}
	if len(exclude) > 0 {
		auditor.SetExclude(exclude)
	}

	return auditor.Run()
}

func GroupEnvUsagesByKey(usages []EnvUsage) []audittypes.VarRef {
	keyMap := make(map[string]*audittypes.VarRef, len(usages))

	for _, u := range usages {
		ref, ok := keyMap[u.Key]
		if !ok {
			ref = &audittypes.VarRef{Key: u.Key}
			keyMap[u.Key] = ref
		}
		ref.References = append(ref.References, audittypes.VarOccurrence{
			File:     u.File,
			Lines:    u.Lines,
			Language: string(u.Language),
		})
	}

	result := make([]audittypes.VarRef, 0, len(keyMap))
	for _, ref := range keyMap {
		result = append(result, *ref)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}
