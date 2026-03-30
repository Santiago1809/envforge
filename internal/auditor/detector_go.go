package auditor

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type GoDetector struct {
	lang Language
}

func NewGoDetector() *GoDetector {
	return &GoDetector{lang: LangGo}
}

func (d *GoDetector) Language() Language   { return d.lang }
func (d *GoDetector) Extensions() []string { return []string{".go"} }

func (d *GoDetector) Detect(path string, content string) ([]EnvUsage, Language, Framework) {
	var vars []EnvUsage

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, LangGo, ""
	}

	framework := detectGoFramework(path)

	ast.Inspect(node, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv") {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					key := strings.Trim(lit.Value, `"`)
					pos := fset.Position(call.Pos())
					vars = append(vars, EnvUsage{
						Key:       key,
						File:      path,
						Lines:     []int{pos.Line},
						Language:  LangGo,
						Framework: framework,
						Method:    "os.Getenv/LookupEnv",
					})
				}
			}
		}

		if ident.Name == "viper" && isViperGet(sel.Sel.Name) {
			if len(call.Args) > 0 {
				if lit, ok := call.Args[0].(*ast.BasicLit); ok && lit.Kind == token.STRING {
					key := strings.Trim(lit.Value, `"`)
					pos := fset.Position(call.Pos())
					vars = append(vars, EnvUsage{
						Key:       key,
						File:      path,
						Lines:     []int{pos.Line},
						Language:  LangGo,
						Framework: framework,
						Method:    "viper.Get",
					})
				}
			}
		}

		return true
	})

	return vars, LangGo, framework
}

func isViperGet(name string) bool {
	return name == "Get" || name == "GetString" || name == "GetInt" || name == "GetBool" ||
		name == "GetInt64" || name == "GetFloat64"
}

func detectGoFramework(path string) Framework {
	dir := filepath.Dir(path)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			content, _ := os.ReadFile(filepath.Join(dir, "go.mod"))
			lower := string(content)
			if strings.Contains(lower, "gin-gonic/gin") || strings.Contains(lower, "gin") {
				return FrameworkNode
			}
			if strings.Contains(lower, "gorm.io") || strings.Contains(lower, "go-gorm") {
				return FrameworkNode
			}
			if strings.Contains(lower, "fiber") {
				return FrameworkNode
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
