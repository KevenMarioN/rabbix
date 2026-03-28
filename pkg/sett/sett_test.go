package sett

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestSett_LoadEnvs(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "rabbix-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create envs.json file
	envsData := map[string]map[string]string{
		"local": {
			"API_URL": "http://localhost:8080",
			"DB_NAME": "test_db",
		},
		"dev": {
			"API_URL": "http://dev.example.com",
			"DB_NAME": "dev_db",
		},
	}

	envsPath := filepath.Join(tmpDir, "envs.json")
	envsJSON, _ := json.MarshalIndent(envsData, "", "  ")
	if err := os.WriteFile(envsPath, envsJSON, 0644); err != nil {
		t.Fatalf("Failed to write envs.json: %v", err)
	}

	// Create settings file pointing to envs.json
	settingsPath := filepath.Join(tmpDir, "settings.json")
	settingsData := map[string]string{
		"envs_file": envsPath,
	}
	settingsJSON, _ := json.MarshalIndent(settingsData, "", "  ")
	if err := os.WriteFile(settingsPath, settingsJSON, 0644); err != nil {
		t.Fatalf("Failed to write settings.json: %v", err)
	}

	tests := []struct {
		name        string
		ambient     string
		expected    map[string]string
		expectError bool
	}{
		{
			name:     "load local environment",
			ambient:  "local",
			expected: envsData["local"],
		},
		{
			name:     "load dev environment",
			ambient:  "dev",
			expected: envsData["dev"],
		},
		{
			name:        "non-existent environment returns error",
			ambient:     "prod",
			expectError: true,
		},
		{
			name:     "empty ambient returns nil",
			ambient:  "",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sett{path: settingsPath}
			result, err := s.LoadEnvs(tt.ambient)

			if tt.expectError {
				if err == nil {
					t.Error("LoadEnvs() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("LoadEnvs() unexpected error: %v", err)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("LoadEnvs() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSett_ListAmbients(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "rabbix-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create envs.json file
	envsData := map[string]map[string]string{
		"local": {"VAR": "value1"},
		"dev":   {"VAR": "value2"},
		"prod":  {"VAR": "value3"},
	}

	envsPath := filepath.Join(tmpDir, "envs.json")
	envsJSON, _ := json.MarshalIndent(envsData, "", "  ")
	if err := os.WriteFile(envsPath, envsJSON, 0644); err != nil {
		t.Fatalf("Failed to write envs.json: %v", err)
	}

	// Create settings file pointing to envs.json
	settingsPath := filepath.Join(tmpDir, "settings.json")
	settingsData := map[string]string{
		"envs_file": envsPath,
	}
	settingsJSON, _ := json.MarshalIndent(settingsData, "", "  ")
	if err := os.WriteFile(settingsPath, settingsJSON, 0644); err != nil {
		t.Fatalf("Failed to write settings.json: %v", err)
	}

	s := &Sett{path: settingsPath}
	result, err := s.ListAmbients()

	if err != nil {
		t.Errorf("ListAmbients() unexpected error: %v", err)
		return
	}

	expected := []string{"local", "dev", "prod"}
	sort.Strings(result)
	sort.Strings(expected)

	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ListAmbients() = %v, want %v", result, expected)
	}
}

func TestSett_ListAmbients_NoEnvsFile(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "rabbix-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create settings file without envs_file
	settingsPath := filepath.Join(tmpDir, "settings.json")
	settingsData := map[string]string{
		"host": "http://localhost:15672",
	}
	settingsJSON, _ := json.MarshalIndent(settingsData, "", "  ")
	if err := os.WriteFile(settingsPath, settingsJSON, 0644); err != nil {
		t.Fatalf("Failed to write settings.json: %v", err)
	}

	s := &Sett{path: settingsPath}
	result, err := s.ListAmbients()

	if err != nil {
		t.Errorf("ListAmbients() unexpected error: %v", err)
		return
	}

	if result != nil {
		t.Errorf("ListAmbients() = %v, want nil", result)
	}
}

func TestSett_LoadEnvs_AbsolutePath(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "rabbix-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create envs.json file
	envsData := map[string]map[string]string{
		"local": {"VAR": "value"},
	}

	envsPath := filepath.Join(tmpDir, "envs.json")
	envsJSON, _ := json.MarshalIndent(envsData, "", "  ")
	if err := os.WriteFile(envsPath, envsJSON, 0644); err != nil {
		t.Fatalf("Failed to write envs.json: %v", err)
	}

	// Create settings file with absolute path
	settingsPath := filepath.Join(tmpDir, "settings.json")
	settingsData := map[string]string{
		"envs_file": envsPath, // absolute path
	}
	settingsJSON, _ := json.MarshalIndent(settingsData, "", "  ")
	if err := os.WriteFile(settingsPath, settingsJSON, 0644); err != nil {
		t.Fatalf("Failed to write settings.json: %v", err)
	}

	s := &Sett{path: settingsPath}
	result, err := s.LoadEnvs("local")

	if err != nil {
		t.Errorf("LoadEnvs() unexpected error: %v", err)
		return
	}

	expected := envsData["local"]
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("LoadEnvs() = %v, want %v", result, expected)
	}
}
