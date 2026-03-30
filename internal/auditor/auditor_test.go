package auditor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGoDetector(t *testing.T) {
	detector := NewGoDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "go", "main.go")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangGo {
		t.Errorf("expected language go, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY", "DATABASE_URL"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestTypeScriptDetector(t *testing.T) {
	detector := NewTypeScriptDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "ts", "app.ts")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangTS {
		t.Errorf("expected language ts, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestJavaScriptDetector(t *testing.T) {
	detector := NewJavaScriptDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "js", "app.js")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangJS {
		t.Errorf("expected language js, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestPythonDetector(t *testing.T) {
	detector := NewPythonDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "py", "app.py")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangPython {
		t.Errorf("expected language py, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestShellDetector(t *testing.T) {
	detector := NewShellDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "sh", "script.sh")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangShell {
		t.Errorf("expected language sh, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestJavaDetector(t *testing.T) {
	detector := NewJavaDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "java", "Config.java")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, framework := detector.Detect(path, string(data))
	if lang != LangJava {
		t.Errorf("expected language java, got %s", lang)
	}

	_ = framework

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestPHPDetector(t *testing.T) {
	detector := NewPHPDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "php", "config.php")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangPHP {
		t.Errorf("expected language php, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestRubyDetector(t *testing.T) {
	detector := NewRubyDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "ruby", "config.rb")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangRuby {
		t.Errorf("expected language ruby, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestCSharpDetector(t *testing.T) {
	detector := NewCSharpDetector()
	path := filepath.Join("..", "..", "testdata", "fixtures", "cs", "Config.cs")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	results, lang, _ := detector.Detect(path, string(data))
	if lang != LangCSharp {
		t.Errorf("expected language cs, got %s", lang)
	}

	found := make(map[string]bool)
	for _, r := range results {
		found[r.Key] = true
	}

	expected := []string{"DB_HOST", "DB_PORT", "API_KEY"}
	for _, key := range expected {
		if !found[key] {
			t.Errorf("expected to find key %q", key)
		}
	}
}

func TestGetDetectorForFile(t *testing.T) {
	tests := []struct {
		path     string
		expected SemanticDetector
	}{
		{"test.go", &GoDetector{}},
		{"test.ts", &TypeScriptDetector{}},
		{"test.js", &JavaScriptDetector{}},
		{"test.py", &PythonDetector{}},
		{"test.sh", &ShellDetector{}},
		{"test.java", &JavaDetector{}},
		{"test.php", &PHPDetector{}},
		{"test.rb", &RubyDetector{}},
		{"test.cs", &CSharpDetector{}},
	}

	for _, tt := range tests {
		detector, err := GetDetectorForFile(tt.path)
		if err != nil {
			t.Errorf("GetDetectorForFile(%s) error = %v", tt.path, err)
			continue
		}
		if detector == nil {
			t.Errorf("GetDetectorForFile(%s) returned nil", tt.path)
		}
		_ = tt.expected
	}
}

func TestAuditDir(t *testing.T) {
	result, err := AuditDir(".", "", []Language{LangGo}, nil, false)
	if err != nil {
		t.Fatalf("AuditDir() error = %v", err)
	}

	if result == nil {
		t.Fatal("AuditDir() returned nil result")
	}
}
