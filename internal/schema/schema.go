package schema

import (
	"errors"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type VarType string

const (
	TypeString VarType = "string"
	TypeInt    VarType = "int"
	TypeFloat  VarType = "float"
	TypeBool   VarType = "bool"
	TypeURL    VarType = "url"
	TypeEmail  VarType = "email"
	TypeEnum   VarType = "enum"
	TypeRegex  VarType = "regex"
)

type VarSchema struct {
	Name    string
	Type    VarType
	Options []string
	Pattern string
}

type Schema struct {
	Vars map[string]VarSchema
}

type ValidationError struct {
	Name    string  `json:"name"`
	Value   string  `json:"value"`
	Type    VarType `json:"type"`
	Message string  `json:"message"`
}

func Parse(path string) (*Schema, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	schema := &Schema{Vars: make(map[string]VarSchema, len(lines)/2)} // estimate capacity

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		name := strings.TrimSpace(parts[0])
		typeStr := strings.TrimSpace(parts[1])

		varSchema := VarSchema{Name: name, Type: VarType(typeStr)}

		if strings.HasPrefix(typeStr, "enum:") {
			varSchema.Type = TypeEnum
			values := strings.TrimPrefix(typeStr, "enum:")
			varSchema.Options = strings.Split(values, ",")
		} else if strings.HasPrefix(typeStr, "regex:") {
			varSchema.Type = TypeRegex
			pattern := strings.TrimPrefix(typeStr, "regex:")
			_, err := regexp.Compile(pattern)
			if err != nil {
				return nil, errors.New("invalid regex pattern for " + name + ": " + err.Error())
			}
			varSchema.Pattern = pattern
		}

		schema.Vars[name] = varSchema
	}

	return schema, nil
}

func Infer(vars map[string]string) *Schema {
	schema := &Schema{Vars: make(map[string]VarSchema, len(vars))}

	for name, value := range vars {
		varType := inferType(value)
		schema.Vars[name] = VarSchema{Name: name, Type: varType}
	}

	return schema
}

func inferType(value string) VarType {
	if value == "" {
		return TypeString
	}

	if _, err := strconv.ParseBool(value); err == nil {
		return TypeBool
	}

	if _, err := strconv.ParseInt(value, 0, 64); err == nil {
		return TypeInt
	}

	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return TypeFloat
	}

	if strings.Contains(value, "://") {
		u, err := url.Parse(value)
		if err == nil && u.Scheme != "" && u.Host != "" {
			return TypeURL
		}
	}

	if strings.Contains(value, "@") {
		atIndex := strings.Index(value, "@")
		domain := value[atIndex+1:]
		if strings.Contains(domain, ".") && domain != "" {
			return TypeEmail
		}
	}

	return TypeString
}

func (s *Schema) Validate(vars map[string]string) []ValidationError {
	errors := make([]ValidationError, 0, len(s.Vars))

	for name, varSchema := range s.Vars {
		value, exists := vars[name]
		if !exists {
			continue
		}

		switch varSchema.Type {
		case TypeBool:
			if !validateBool(value) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be bool, got '" + value + "'",
				})
			}
		case TypeURL:
			if !validateURL(value) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be url, got '" + value + "'",
				})
			}
		case TypeEmail:
			if !validateEmail(value) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be email, got '" + value + "'",
				})
			}
		case TypeInt:
			if !validateInt(value) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be int, got '" + value + "'",
				})
			}
		case TypeFloat:
			if !validateFloat(value) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be float, got '" + value + "'",
				})
			}
		case TypeEnum:
			if !validateEnum(value, varSchema.Options) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must be one of " + strings.Join(varSchema.Options, ", ") + ", got '" + value + "'",
				})
			}
		case TypeRegex:
			if !validateRegex(value, varSchema.Pattern) {
				errors = append(errors, ValidationError{
					Name:    name,
					Value:   value,
					Type:    varSchema.Type,
					Message: "must match pattern " + varSchema.Pattern + ", got '" + value + "'",
				})
			}
		}
	}

	return errors
}

func validateBool(value string) bool {
	lower := strings.ToLower(value)
	return lower == "true" || lower == "false" || lower == "1" || lower == "0" || lower == "yes" || lower == "no"
}

func validateURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validateEmail(value string) bool {
	atIndex := strings.Index(value, "@")
	if atIndex == -1 {
		return false
	}
	domain := value[atIndex+1:]
	return strings.Contains(domain, ".") && domain != ""
}

func validateInt(value string) bool {
	_, err := strconv.ParseInt(value, 0, 64)
	return err == nil
}

func validateFloat(value string) bool {
	_, err := strconv.ParseFloat(value, 64)
	return err == nil
}

func validateEnum(value string, options []string) bool {
	for _, opt := range options {
		if value == opt {
			return true
		}
	}
	return false
}

func validateRegex(value string, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(value)
}

func (s *Schema) WriteTo(path string) error {
	var lines []string

	keys := make([]string, 0, len(s.Vars))
	for name := range s.Vars {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, name := range keys {
		varSchema := s.Vars[name]
		var typeStr string

		switch varSchema.Type {
		case TypeEnum:
			typeStr = "enum:" + strings.Join(varSchema.Options, ",")
		case TypeRegex:
			typeStr = "regex:" + varSchema.Pattern
		default:
			typeStr = string(varSchema.Type)
		}

		lines = append(lines, name+"="+typeStr)
	}

	content := strings.Join(lines, "\n")
	if len(lines) > 0 {
		content += "\n"
	}

	return os.WriteFile(path, []byte(content), 0644)
}
