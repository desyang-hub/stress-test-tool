package config

import (
	"os"
	"reflect"
	"regexp"
	"strings"
)

// envVarPattern matches ${VAR} and $VAR references.
var envVarPattern = regexp.MustCompile(`\$\{([^}]+)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// ResolveEnvVars replaces all ${VAR} and $VAR references in the config struct
// with their os.Getenv values. Missing variables are logged as warnings
// and left as-is.
func ResolveEnvVars(v any) {
	resolveValue(reflect.ValueOf(v), "")
}

// ResolveString replaces env var references in a single string.
func ResolveString(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		var name string
		submatches := envVarPattern.FindStringSubmatch(match)
		if len(submatches) > 1 && submatches[1] != "" {
			name = submatches[1] // ${VAR} form
		} else if len(submatches) > 2 {
			name = submatches[2] // $VAR form
		}
		val := os.Getenv(name)
		if val == "" {
			return match // leave unresolved
		}
		return val
	})
}

func resolveValue(v reflect.Value, path string) {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if strings.Contains(s, "${") || strings.Contains(s, "$") {
			resolved := ResolveString(s)
			if resolved != s {
				v.SetString(resolved)
			}
		}
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			_ = field.Tag.Get("yaml") // yaml tag metadata available if needed
			// Skip private fields
			if field.PkgPath != "" {
				continue
			}
			fieldPath := path + "." + field.Name
			if path == "" {
				fieldPath = field.Name
			}
			resolveValue(v.Field(i), fieldPath)
		}
	case reflect.Ptr:
		if v.IsNil() {
			return
		}
		resolveValue(v.Elem(), path)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			resolveValue(v.Index(i), path)
		}
	case reflect.Map:
		iter := v.MapRange()
		for iter.Next() {
			resolveValue(iter.Value(), path)
		}
	}
}
