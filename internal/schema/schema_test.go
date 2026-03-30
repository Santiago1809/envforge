package schema

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	dir := t.TempDir()

	tests := []struct {
		name    string
		content string
		want    *Schema
		wantErr bool
	}{
		{
			name:    "file does not exist",
			content: "",
			want:    nil,
			wantErr: false,
		},
		{
			name:    "valid schema with all types",
			content: "PORT=int\nDATABASE_URL=url\nDEBUG=bool\nRATE=float\nEMAIL=email\n",
			want: &Schema{Vars: map[string]VarSchema{
				"PORT":         {Name: "PORT", Type: TypeInt},
				"DATABASE_URL": {Name: "DATABASE_URL", Type: TypeURL},
				"DEBUG":        {Name: "DEBUG", Type: TypeBool},
				"RATE":         {Name: "RATE", Type: TypeFloat},
				"EMAIL":        {Name: "EMAIL", Type: TypeEmail},
			}},
			wantErr: false,
		},
		{
			name:    "schema with enum",
			content: "APP_ENV=enum:development,staging,production\n",
			want: &Schema{Vars: map[string]VarSchema{
				"APP_ENV": {Name: "APP_ENV", Type: TypeEnum, Options: []string{"development", "staging", "production"}},
			}},
			wantErr: false,
		},
		{
			name:    "schema with regex",
			content: "PATTERN=regex:^[a-z]+$\n",
			want: &Schema{Vars: map[string]VarSchema{
				"PATTERN": {Name: "PATTERN", Type: TypeRegex, Pattern: "^[a-z]+$"},
			}},
			wantErr: false,
		},
		{
			name:    "invalid regex",
			content: "PATTERN=regex:[invalid\n",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "with comments and empty lines",
			content: "# comment\nPORT=int\n\nDEBUG=bool\n",
			want: &Schema{Vars: map[string]VarSchema{
				"PORT":  {Name: "PORT", Type: TypeInt},
				"DEBUG": {Name: "DEBUG", Type: TypeBool},
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, "test.schema")
			if tt.name == "file does not exist" {
				path = filepath.Join(dir, "nonexistent.schema")
			} else {
				os.WriteFile(path, []byte(tt.content), 0644)
			}

			got, err := Parse(path)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if got == nil && tt.want == nil {
				return
			}
			if got == nil || tt.want == nil {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
				return
			}
			for name, wantVar := range tt.want.Vars {
				gotVar, ok := got.Vars[name]
				if !ok {
					t.Errorf("Parse() missing key %s", name)
					continue
				}
				if gotVar.Type != wantVar.Type {
					t.Errorf("Parse() type = %v, want %v", gotVar.Type, wantVar.Type)
				}
				if len(gotVar.Options) != len(wantVar.Options) {
					t.Errorf("Parse() options = %v, want %v", gotVar.Options, wantVar.Options)
				}
				if gotVar.Pattern != wantVar.Pattern {
					t.Errorf("Parse() pattern = %v, want %v", gotVar.Pattern, wantVar.Pattern)
				}
			}
		})
	}
}

func TestInfer(t *testing.T) {
	tests := []struct {
		name     string
		vars     map[string]string
		wantType VarType
	}{
		{
			name:     "empty value",
			vars:     map[string]string{"EMPTY": ""},
			wantType: TypeString,
		},
		{
			name:     "bool true",
			vars:     map[string]string{"DEBUG": "true"},
			wantType: TypeBool,
		},
		{
			name:     "bool false",
			vars:     map[string]string{"DEBUG": "false"},
			wantType: TypeBool,
		},
		{
			name:     "bool 1",
			vars:     map[string]string{"DEBUG": "1"},
			wantType: TypeBool,
		},
		{
			name:     "bool 0",
			vars:     map[string]string{"DEBUG": "0"},
			wantType: TypeBool,
		},
		{
			name:     "int",
			vars:     map[string]string{"PORT": "8080"},
			wantType: TypeInt,
		},
		{
			name:     "float",
			vars:     map[string]string{"RATE": "3.14"},
			wantType: TypeFloat,
		},
		{
			name:     "url",
			vars:     map[string]string{"DATABASE_URL": "postgres://localhost:5432/db"},
			wantType: TypeURL,
		},
		{
			name:     "email",
			vars:     map[string]string{"EMAIL": "test@example.com"},
			wantType: TypeEmail,
		},
		{
			name:     "string fallback",
			vars:     map[string]string{"NAME": "hello"},
			wantType: TypeString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Infer(tt.vars)
			for name, varSchema := range got.Vars {
				if varSchema.Type != tt.wantType {
					t.Errorf("Infer() %s = %v, want %v", name, varSchema.Type, tt.wantType)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		schema       *Schema
		vars         map[string]string
		wantErrCount int
	}{
		{
			name: "valid int",
			schema: &Schema{Vars: map[string]VarSchema{
				"PORT": {Name: "PORT", Type: TypeInt},
			}},
			vars:         map[string]string{"PORT": "8080"},
			wantErrCount: 0,
		},
		{
			name: "invalid int",
			schema: &Schema{Vars: map[string]VarSchema{
				"PORT": {Name: "PORT", Type: TypeInt},
			}},
			vars:         map[string]string{"PORT": "abc"},
			wantErrCount: 1,
		},
		{
			name: "valid bool",
			schema: &Schema{Vars: map[string]VarSchema{
				"DEBUG": {Name: "DEBUG", Type: TypeBool},
			}},
			vars:         map[string]string{"DEBUG": "true"},
			wantErrCount: 0,
		},
		{
			name: "valid bool yes/no",
			schema: &Schema{Vars: map[string]VarSchema{
				"DEBUG": {Name: "DEBUG", Type: TypeBool},
			}},
			vars:         map[string]string{"DEBUG": "yes"},
			wantErrCount: 0,
		},
		{
			name: "invalid bool",
			schema: &Schema{Vars: map[string]VarSchema{
				"DEBUG": {Name: "DEBUG", Type: TypeBool},
			}},
			vars:         map[string]string{"DEBUG": "maybe"},
			wantErrCount: 1,
		},
		{
			name: "valid url",
			schema: &Schema{Vars: map[string]VarSchema{
				"URL": {Name: "URL", Type: TypeURL},
			}},
			vars:         map[string]string{"URL": "https://example.com"},
			wantErrCount: 0,
		},
		{
			name: "invalid url no scheme",
			schema: &Schema{Vars: map[string]VarSchema{
				"URL": {Name: "URL", Type: TypeURL},
			}},
			vars:         map[string]string{"URL": "localhost"},
			wantErrCount: 1,
		},
		{
			name: "valid enum",
			schema: &Schema{Vars: map[string]VarSchema{
				"ENV": {Name: "ENV", Type: TypeEnum, Options: []string{"dev", "prod"}},
			}},
			vars:         map[string]string{"ENV": "dev"},
			wantErrCount: 0,
		},
		{
			name: "invalid enum",
			schema: &Schema{Vars: map[string]VarSchema{
				"ENV": {Name: "ENV", Type: TypeEnum, Options: []string{"dev", "prod"}},
			}},
			vars:         map[string]string{"ENV": "invalid"},
			wantErrCount: 1,
		},
		{
			name: "valid regex",
			schema: &Schema{Vars: map[string]VarSchema{
				"PATTERN": {Name: "PATTERN", Type: TypeRegex, Pattern: "^[a-z]+$"},
			}},
			vars:         map[string]string{"PATTERN": "abc"},
			wantErrCount: 0,
		},
		{
			name: "invalid regex",
			schema: &Schema{Vars: map[string]VarSchema{
				"PATTERN": {Name: "PATTERN", Type: TypeRegex, Pattern: "^[a-z]+$"},
			}},
			vars:         map[string]string{"PATTERN": "ABC"},
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.schema.Validate(tt.vars)
			if len(got) != tt.wantErrCount {
				t.Errorf("Validate() got %d errors, want %d", len(got), tt.wantErrCount)
				for _, e := range got {
					t.Logf("  ValidationError: %v", e)
				}
			}
		})
	}
}

func TestWriteTo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "output.schema")

	s := &Schema{Vars: map[string]VarSchema{
		"PORT":    {Name: "PORT", Type: TypeInt},
		"URL":     {Name: "URL", Type: TypeURL},
		"ENV":     {Name: "ENV", Type: TypeEnum, Options: []string{"dev", "prod"}},
		"PATTERN": {Name: "PATTERN", Type: TypeRegex, Pattern: "^[a-z]+$"},
	}}

	err := s.WriteTo(path)
	if err != nil {
		t.Fatalf("WriteTo() error = %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	expected := "ENV=enum:dev,prod\nPATTERN=regex:^[a-z]+$\nPORT=int\nURL=url\n"
	if string(content) != expected {
		t.Errorf("WriteTo() got %q, want %q", string(content), expected)
	}
}
