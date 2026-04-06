package check

import (
	"fmt"
	"os"
	"strings"

	"github.com/Santiago1809/envforge/internal/parser"
	"github.com/Santiago1809/envforge/internal/schema"
)

type CheckResult struct {
	Valid         bool
	MissingKeys   []string
	EmptyKeys     []string
	PresentKeys   []string
	MissingCount  int
	EmptyCount    int
	TotalRequired int
	TypeErrors    []schema.ValidationError `json:"type_errors"`
}

type Options struct {
	Required   []string
	FromFile   string
	AllowEmpty bool
	Prefix     string
}

func Check(opts *Options) (*CheckResult, error) {
	reqCount := len(opts.Required)
	result := &CheckResult{
		Valid:         true,
		MissingKeys:   make([]string, 0, reqCount),
		EmptyKeys:     make([]string, 0, reqCount),
		PresentKeys:   make([]string, 0, reqCount),
		MissingCount:  0,
		EmptyCount:    0,
		TotalRequired: reqCount,
	}

	required := opts.Required

	var envFile *parser.EnvFile
	if opts.FromFile != "" {
		var err error
		envFile, err = parser.Load(opts.FromFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file: %w", err)
		}
		if len(required) == 0 {
			required = envFile.Keys()
		}
		result.TotalRequired = len(required)
	}

	for _, key := range required {
		if opts.Prefix != "" && !strings.HasPrefix(key, opts.Prefix) {
			continue
		}

		var value string
		var exists bool

		if envFile != nil {
			value, exists = envFile.Get(key)
		} else {
			value, exists = os.LookupEnv(key)
		}

		if !exists {
			result.MissingKeys = append(result.MissingKeys, key)
			result.MissingCount++
			result.Valid = false
			continue
		}

		if opts.AllowEmpty || value != "" {
			result.PresentKeys = append(result.PresentKeys, key)
		} else {
			result.EmptyKeys = append(result.EmptyKeys, key)
			result.EmptyCount++
			result.Valid = false
		}
	}

	return result, nil
}

func CheckRequiredKeys(keys []string, allowEmpty bool) (*CheckResult, error) {
	opts := &Options{
		Required:   keys,
		AllowEmpty: allowEmpty,
	}
	return Check(opts)
}

func CheckFromFile(envFile string, allowEmpty bool, prefix string) (*CheckResult, error) {
	opts := &Options{
		FromFile:   envFile,
		AllowEmpty: allowEmpty,
		Prefix:     prefix,
	}
	return Check(opts)
}

func CheckKeys(keys []string) error {
	opts := &Options{
		Required: keys,
	}

	result, err := Check(opts)
	if err != nil {
		return err
	}

	if !result.Valid {
		if len(result.MissingKeys) > 0 {
			fmt.Printf("Error: Missing required environment variables:\n")
			for _, key := range result.MissingKeys {
				fmt.Printf("  - %s\n", key)
			}
		}
		if len(result.EmptyKeys) > 0 {
			fmt.Printf("Error: Required environment variables are empty:\n")
			for _, key := range result.EmptyKeys {
				fmt.Printf("  - %s\n", key)
			}
		}
	}

	return nil
}

func GetEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

func GetEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func GetAllWithPrefix(prefix string) map[string]string {
	envVars := os.Environ()
	// Estimate capacity: typically half of all env vars will match prefix? Use total as upper bound
	result := make(map[string]string, len(envVars))
	for _, env := range envVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		value := parts[1]
		if strings.HasPrefix(key, prefix) {
			result[key] = value
		}
	}
	return result
}

func RunWithSchema(envFile string, s *schema.Schema, required []string, allowEmpty bool, prefix string) (*CheckResult, error) {
	opts := &Options{
		FromFile:   envFile,
		Required:   required,
		AllowEmpty: allowEmpty,
		Prefix:     prefix,
	}

	result, err := Check(opts)
	if err != nil {
		return nil, err
	}

	if s != nil {
		envVars := make(map[string]string)
		if envFile != "" {
			env, err := parser.Load(envFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load env file: %w", err)
			}
			keys := env.Keys()
			// Preallocate map with estimated capacity
			envVars = make(map[string]string, len(keys))
			for _, key := range keys {
				val, _ := env.Get(key)
				envVars[key] = val
			}
		}

		result.TypeErrors = s.Validate(envVars)
	}

	return result, nil
}
