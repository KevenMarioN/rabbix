package sett

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SettItf interface {
	LoadSettings() map[string]string
	SaveSettings(settings map[string]string)
	GetBaseDir() string
	LoadEnvs(ambient string) (map[string]string, error)
	ListAmbients() ([]string, error)
}

var _ SettItf = (*Sett)(nil)

type Sett struct {
	path string
}

func New() *Sett {
	return &Sett{
		path: getSettingsPath(),
	}
}

func getSettingsPath() string {
	baseDir := getBaseDir()
	pathSett := filepath.Join(baseDir, "settings.json")

	// Guarantees base directory
	_ = os.MkdirAll(baseDir, os.ModePerm)

	// Defaults
	defaultSett := map[string]string{"sett": "local.json"}
	defaultCfg := map[string]string{
		"auth":       "Z3Vlc3Q6Z3Vlc3Q=",
		"host":       "http://localhost:15672",
		"output_dir": baseDir,
	}

	// Create settings.json if it does not exist.
	if _, err := os.Stat(pathSett); os.IsNotExist(err) {
		if data, err := json.MarshalIndent(defaultSett, "", "  "); err == nil {
			_ = os.WriteFile(pathSett, data, 0644)
		}
	}

	// Loads settings.json
	settings := map[string]string{}
	if data, err := os.ReadFile(pathSett); err == nil {
		_ = json.Unmarshal(data, &settings)
	}

	// Guarantees the "sett" key and persists if necessary.
	if settings["sett"] == "" {
		settings["sett"] = defaultSett["sett"]
		if data, err := json.MarshalIndent(settings, "", "  "); err == nil {
			_ = os.WriteFile(pathSett, data, 0644)
		}
	}

	// Target configuration file path
	targetPath := filepath.Join(baseDir, settings["sett"])

	// Creates the target file with defaults if it doesn't exist.
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		if data, err := json.MarshalIndent(defaultCfg, "", "  "); err == nil {
			_ = os.WriteFile(targetPath, data, 0644)
		}
	}

	return targetPath
}

func getBaseDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".rabbix")
}

func (s *Sett) GetBaseDir() string {
	return getBaseDir()
}

func (s *Sett) LoadSettings() map[string]string {
	settings := map[string]string{}

	if data, err := os.ReadFile(s.path); err == nil {
		_ = json.Unmarshal(data, &settings)
	}

	return settings
}

func (s *Sett) SaveSettings(settings map[string]string) {
	_ = os.MkdirAll(filepath.Dir(s.path), os.ModePerm)

	data, _ := json.MarshalIndent(settings, "", "  ")
	_ = os.WriteFile(s.path, data, 0644)
}

func (s *Sett) LoadEnvs(ambient string) (map[string]string, error) {
	settings := s.LoadSettings()
	envsFile := settings["envs_file"]
	if envsFile == "" {
		return nil, nil
	}

	// Suporta caminho absoluto ou relativo ao baseDir
	var envsPath string
	if filepath.IsAbs(envsFile) {
		envsPath = envsFile
	} else {
		envsPath = filepath.Join(s.GetBaseDir(), envsFile)
	}

	data, err := os.ReadFile(envsPath)
	if err != nil {
		return nil, fmt.Errorf("arquivo de envs não encontrado: %s", envsPath)
	}

	var allEnvs map[string]map[string]string
	if err := json.Unmarshal(data, &allEnvs); err != nil {
		return nil, fmt.Errorf("erro ao parsear arquivo de envs: %v", err)
	}

	if ambient == "" {
		return nil, nil
	}

	envs, ok := allEnvs[ambient]
	if !ok {
		return nil, fmt.Errorf("ambiente '%s' não encontrado", ambient)
	}

	return envs, nil
}

func (s *Sett) ListAmbients() ([]string, error) {
	settings := s.LoadSettings()
	envsFile := settings["envs_file"]
	if envsFile == "" {
		return nil, nil
	}

	// Suporta caminho absoluto ou relativo ao baseDir
	var envsPath string
	if filepath.IsAbs(envsFile) {
		envsPath = envsFile
	} else {
		envsPath = filepath.Join(s.GetBaseDir(), envsFile)
	}

	data, err := os.ReadFile(envsPath)
	if err != nil {
		return nil, fmt.Errorf("arquivo de envs não encontrado: %s", envsPath)
	}

	var allEnvs map[string]map[string]string
	if err := json.Unmarshal(data, &allEnvs); err != nil {
		return nil, fmt.Errorf("erro ao parsear arquivo de envs: %v", err)
	}

	ambients := make([]string, 0, len(allEnvs))
	for name := range allEnvs {
		ambients = append(ambients, name)
	}

	return ambients, nil
}
