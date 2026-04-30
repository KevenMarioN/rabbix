package run

import (
	"reflect"
	"sort"
	"testing"
)

func TestReplaceEnvs(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		envs     map[string]string
		expected []byte
	}{
		{
			name:     "nil envs returns original data",
			data:     []byte(`{"message": "${VAR}"}`),
			envs:     nil,
			expected: []byte(`{"message": "${VAR}"}`),
		},
		{
			name:     "empty envs returns original data",
			data:     []byte(`{"message": "${VAR}"}`),
			envs:     map[string]string{},
			expected: []byte(`{"message": "${VAR}"}`),
		},
		{
			name:     "single variable replacement",
			data:     []byte(`{"message": "${VAR}"}`),
			envs:     map[string]string{"VAR": "value"},
			expected: []byte(`{"message": "value"}`),
		},
		{
			name:     "multiple variables replacement",
			data:     []byte(`{"message": "${VAR1}", "user": "${VAR2}"}`),
			envs:     map[string]string{"VAR1": "hello", "VAR2": "world"},
			expected: []byte(`{"message": "hello", "user": "world"}`),
		},
		{
			name:     "same variable multiple times",
			data:     []byte(`{"a": "${VAR}", "b": "${VAR}"}`),
			envs:     map[string]string{"VAR": "same"},
			expected: []byte(`{"a": "same", "b": "same"}`),
		},
		{
			name:     "variable not in envs remains unchanged",
			data:     []byte(`{"message": "${MISSING}"}`),
			envs:     map[string]string{"OTHER": "value"},
			expected: []byte(`{"message": "${MISSING}"}`),
		},
		{
			name:     "mixed replaced and not replaced",
			data:     []byte(`{"a": "${VAR1}", "b": "${MISSING}"}`),
			envs:     map[string]string{"VAR1": "replaced"},
			expected: []byte(`{"a": "replaced", "b": "${MISSING}"}`),
		},
		{
			name:     "variable with underscore and numbers",
			data:     []byte(`{"url": "${API_URL_V2}"}`),
			envs:     map[string]string{"API_URL_V2": "http://api.example.com/v2"},
			expected: []byte(`{"url": "http://api.example.com/v2"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceEnvs(tt.data, tt.envs)
			if string(result) != string(tt.expected) {
				t.Errorf("ReplaceEnvs() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestFindMissingEnvs(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		envs     map[string]string
		expected []string
	}{
		{
			name:     "nil envs returns nil",
			data:     []byte(`{"message": "${VAR}"}`),
			envs:     nil,
			expected: nil,
		},
		{
			name:     "no variables in data",
			data:     []byte(`{"message": "static"}`),
			envs:     map[string]string{"VAR": "value"},
			expected: []string{},
		},
		{
			name:     "all variables present",
			data:     []byte(`{"message": "${VAR}"}`),
			envs:     map[string]string{"VAR": "value"},
			expected: []string{},
		},
		{
			name:     "single missing variable",
			data:     []byte(`{"message": "${MISSING}"}`),
			envs:     map[string]string{"OTHER": "value"},
			expected: []string{"MISSING"},
		},
		{
			name:     "multiple missing variables",
			data:     []byte(`{"a": "${VAR1}", "b": "${VAR2}"}`),
			envs:     map[string]string{"OTHER": "value"},
			expected: []string{"VAR1", "VAR2"},
		},
		{
			name:     "mixed present and missing",
			data:     []byte(`{"a": "${VAR1}", "b": "${VAR2}", "c": "${VAR3}"}`),
			envs:     map[string]string{"VAR1": "value1", "VAR3": "value3"},
			expected: []string{"VAR2"},
		},
		{
			name:     "same missing variable multiple times",
			data:     []byte(`{"a": "${VAR}", "b": "${VAR}"}`),
			envs:     map[string]string{},
			expected: []string{"VAR"},
		},
		{
			name:     "variable with underscore and numbers",
			data:     []byte(`{"url": "${API_URL_V2}"}`),
			envs:     map[string]string{},
			expected: []string{"API_URL_V2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FindMissingEnvs(tt.data, tt.envs)

			// Sort both slices for consistent comparison
			sort.Strings(result)
			sort.Strings(tt.expected)

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("FindMissingEnvs() = %v, want %v", result, tt.expected)
			}
		})
	}
}
