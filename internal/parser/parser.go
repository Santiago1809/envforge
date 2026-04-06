package parser

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"unicode"
)

// builderPool reuses strings.Builder instances to reduce allocations
var builderPool = sync.Pool{
	New: func() interface{} {
		return &strings.Builder{}
	},
}

type EnvFile struct {
	entries map[string]string
	order   []string
}

type ParseError struct {
	Line   int
	Column int
	msg    string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("parse error at line %d, column %d: %s", e.Line, e.Column, e.msg)
}

func NewEnvFile() *EnvFile {
	entries := make(map[string]string, 16)
	order := make([]string, 0, 16)
	return &EnvFile{
		entries: entries,
		order:   order,
	}
}

func Load(path string) (*EnvFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return LoadReader(file)
}

func LoadReader(r io.Reader) (*EnvFile, error) {
	env := NewEnvFile()
	scanner := bufio.NewScanner(r)
	lineNum := 0
	var currentKey string
	currentValue := builderPool.Get().(*strings.Builder)
	currentValue.Reset()
	defer builderPool.Put(currentValue)
	inMultiline := false
	quoteChar := rune(0)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if inMultiline {
			if strings.Contains(line, string(quoteChar)) {
				remaining := strings.SplitN(line, string(quoteChar), 2)
				if currentValue.Len() > 0 {
					currentValue.WriteString("\n")
				}
				currentValue.WriteString(remaining[0])
				env.entries[currentKey] = currentValue.String()
				inMultiline = false
				quoteChar = 0
				currentValue.Reset()
			} else {
				if currentValue.Len() > 0 {
					currentValue.WriteString("\n")
				}
				currentValue.WriteString(line)
			}
			continue
		}

		trimmed := strings.TrimLeftFunc(line, unicode.IsSpace)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		key, value, err := parseLine(line, lineNum)
		if err != nil {
			if _, ok := err.(*ParseError); ok {
				return nil, err
			}
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		if key == "" {
			continue
		}

		if strings.HasPrefix(value, "\"") || strings.HasPrefix(value, "'") || strings.HasPrefix(value, "`") {
			quoteChar = rune(value[0])
			if strings.Count(value, string(quoteChar)) >= 2 && !isQuotedContinue(value) {
				cleanValue := value[1 : len(value)-1]
				env.entries[key] = unescapeValue(cleanValue, quoteChar)
			} else if isQuotedContinue(value) {
				currentKey = key
				currentValue.WriteString(value[1:])
				inMultiline = true
				quoteChar = rune(value[0])
				env.order = append(env.order, key)
			} else {
				cleanValue := value[1 : len(value)-1]
				env.entries[key] = unescapeValue(cleanValue, quoteChar)
			}
		} else {
			value = stripInlineComment(value)
			value = strings.TrimRight(value, " ")
			env.entries[key] = value
		}

		if !inMultiline {
			env.order = append(env.order, key)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	if inMultiline {
		return nil, &ParseError{Line: lineNum, Column: 1, msg: "unclosed multiline value"}
	}

	return env, nil
}

func parseLine(line string, lineNum int) (string, string, error) {
	// Get builders from pool to reduce allocations
	keyBuilder := builderPool.Get().(*strings.Builder)
	keyBuilder.Reset()
	valueBuilder := builderPool.Get().(*strings.Builder)
	valueBuilder.Reset()
	defer func() {
		builderPool.Put(keyBuilder)
		builderPool.Put(valueBuilder)
	}()

	state := 0
	var quoteChar rune

	for i, ch := range line {
		switch state {
		case 0:
			if unicode.IsSpace(ch) {
				continue
			}
			if unicode.IsLetter(ch) || ch == '_' {
				keyBuilder.WriteRune(ch)
				state = 1
			} else if ch == '#' {
				return "", "", nil
			} else {
				return "", "", &ParseError{Line: lineNum, Column: i + 1, msg: fmt.Sprintf("invalid character '%c' in key", ch)}
			}
		case 1:
			if ch == '=' {
				state = 2
			} else if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
				keyBuilder.WriteRune(ch)
			} else {
				return "", "", &ParseError{Line: lineNum, Column: i + 1, msg: fmt.Sprintf("invalid character '%c' in key", ch)}
			}
		case 2:
			if unicode.IsSpace(ch) {
				continue
			}
			if ch == '"' || ch == '\'' || ch == '`' {
				quoteChar = ch
				valueBuilder.WriteRune(ch)
				state = 3
			} else if ch == '#' {
				return keyBuilder.String(), valueBuilder.String(), nil
			} else {
				valueBuilder.WriteRune(ch)
				state = 4
			}
		case 3:
			valueBuilder.WriteRune(ch)
			if ch == quoteChar {
				if i > 0 && line[i-1] == '\\' {
					continue
				}
				state = 4
			}
		case 4:
			if ch == '#' && valueBuilder.Len() > 0 && isLastNonSpace(valueBuilder.String(), i, line) {
				return keyBuilder.String(), strings.TrimRight(valueBuilder.String(), " "), nil
			}
			valueBuilder.WriteRune(ch)
		}
	}

	return keyBuilder.String(), valueBuilder.String(), nil
}

func isQuotedContinue(value string) bool {
	quote := rune(value[0])
	count := 0
	for _, ch := range value {
		if ch == quote {
			count++
		}
	}
	return count%2 != 0
}

func unescapeValue(value string, quoteChar rune) string {
	if quoteChar == '"' {
		value = strings.ReplaceAll(value, "\\n", "\n")
		value = strings.ReplaceAll(value, "\\t", "\t")
		value = strings.ReplaceAll(value, "\\\"", "\"")
		value = strings.ReplaceAll(value, "\\\\", "\\")
	}
	return value
}

func isLastNonSpace(value string, idx int, line string) bool {
	afterHash := strings.TrimLeft(line[idx:], " \t")
	return len(afterHash) == 1 && afterHash[0] == '#'
}

func stripInlineComment(value string) string {
	for i := len(value) - 1; i >= 0; i-- {
		if value[i] == '#' && (i == 0 || value[i-1] == ' ') {
			return strings.TrimRight(value[:i], " ")
		}
	}
	return value
}

func (e *EnvFile) Keys() []string {
	result := make([]string, len(e.order))
	copy(result, e.order)
	return result
}

func (e *EnvFile) Get(key string) (string, bool) {
	val, ok := e.entries[key]
	return val, ok
}

func (e *EnvFile) Set(key, value string) {
	if _, exists := e.entries[key]; !exists {
		e.order = append(e.order, key)
	}
	e.entries[key] = value
}

func (e *EnvFile) Write(path string) error {
	var buf bytes.Buffer
	for _, key := range e.order {
		value := e.entries[key]
		if needsQuoting(value) {
			fmt.Fprintf(&buf, "%s=\"%s\"\n", key, escapeValue(value))
		} else {
			fmt.Fprintf(&buf, "%s=%s\n", key, value)
		}
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

func needsQuoting(value string) bool {
	for _, ch := range value {
		if ch == ' ' || ch == '#' || ch == '\n' || ch == '\t' || ch == '"' || ch == '\'' || ch == '`' {
			return true
		}
	}
	return false
}

func escapeValue(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	value = strings.ReplaceAll(value, "\t", "\\t")
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}

func (e *EnvFile) Expand() {
	for i, key := range e.order {
		value := e.entries[key]
		expanded := expandVariables(value, e.entries)
		e.entries[key] = expanded
		e.order[i] = key
	}
}

func expandVariables(value string, env map[string]string) string {
	result := value
	for key, val := range env {
		placeholder := "${" + key + "}"
		result = strings.ReplaceAll(result, placeholder, val)
	}
	return result
}

func (e *EnvFile) WritePretty(path string) error {
	var buf bytes.Buffer
	buf.WriteString("# Generated by envoy\n\n")
	for _, key := range e.order {
		value := e.entries[key]
		buf.WriteString(fmt.Sprintf("%s=%s\n", key, value))
	}

	return os.WriteFile(path, buf.Bytes(), 0644)
}

var ErrKeyNotFound = errors.New("key not found")

func (e *EnvFile) Delete(key string) error {
	idx := -1
	for i, k := range e.order {
		if k == key {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrKeyNotFound
	}
	e.order = append(e.order[:idx], e.order[idx+1:]...)
	delete(e.entries, key)
	return nil
}
