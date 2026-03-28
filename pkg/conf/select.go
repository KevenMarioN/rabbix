package conf

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func (c *Conf) CmdSelect() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "select [nome]",
		Short: "Seleciona uma configuração existente ou cria uma nova",
		Args:  cobra.MaximumNArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			baseDir := c.settings.GetBaseDir()
			files := listConfigFiles(baseDir)

			var res []string

			for _, f := range files {
				if strings.HasPrefix(strings.ToLower(f), strings.ToLower(toComplete)) {
					res = append(res, f)
				}
			}

			return res, cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) {
			baseDir := c.settings.GetBaseDir()
			_ = os.MkdirAll(baseDir, os.ModePerm)

			if len(args) == 0 {
				opts := listConfigFiles(baseDir)
				if len(opts) == 0 {
					fmt.Println("Nenhuma configuração encontrada. Informe um nome para criar uma nova, " +
						"por exemplo: rabbix conf select minha.json")
				} else {
					fmt.Println("Informe o nome da configuração. Disponíveis:")

					for _, o := range opts {
						fmt.Println("- " + o)
					}
				}

				return
			}

			profileName := args[0]

			file := profileName
			if !strings.HasSuffix(strings.ToLower(file), ".json") {
				file += ".json"
			}

			target := filepath.Join(baseDir, file)

			// Create profile directory first (always ensure it exists)
			profileDir := filepath.Join(baseDir, strings.TrimSuffix(file, ".json"))
			_ = os.MkdirAll(profileDir, os.ModePerm)

			// Create the configuration if it does not exist.
			if _, err := os.Stat(target); os.IsNotExist(err) {
				defaultCfg := map[string]string{
					"auth":        "Z3Vlc3Q6Z3Vlc3Q=",
					"host":        "http://localhost:15672",
					"output_dir":  profileDir,
					"envs_file":   filepath.Join(profileDir, "envs.json"),
					"default_env": "local",
				}
				if data, err := json.MarshalIndent(defaultCfg, "", "  "); err == nil {
					_ = os.WriteFile(target, data, 0644)
				}

				// Create default envs.json if it does not exist
				envsPath := filepath.Join(profileDir, "envs.json")
				if _, err := os.Stat(envsPath); os.IsNotExist(err) {
					defaultEnvs := map[string]map[string]string{
						"local": {
							"EXAMPLE_VAR": "example_value",
						},
					}
					if data, err := json.MarshalIndent(defaultEnvs, "", "  "); err == nil {
						_ = os.WriteFile(envsPath, data, 0644)
					}
				}

				// Create default test example file
				testsDir := profileDir
				defaultTest := map[string]any{
					"name":      "example-test",
					"route_key": "example.route.key",
					"json_pool": map[string]any{
						"message": "${EXAMPLE_VAR}",
						"user_id": 123,
					},
					"headers": map[string]any{
						"Content-Type": "application/json",
					},
				}

				testPath := filepath.Join(testsDir, "example-test.json")
				if data, err := json.MarshalIndent(defaultTest, "", "  "); err == nil {
					_ = os.WriteFile(testPath, data, 0644)
				}

				fmt.Println("Criada nova configuração:", file)
			}

			// Updates settings.json with the selected file
			settPath := filepath.Join(baseDir, "settings.json")

			settings := map[string]string{"sett": file}
			if data, err := os.ReadFile(settPath); err == nil {
				_ = json.Unmarshal(data, &settings)
				settings["sett"] = file
			}

			if data, err := json.MarshalIndent(settings, "", "  "); err == nil {
				_ = os.WriteFile(settPath, data, 0644)
			}

			fmt.Println("Configuração ativa atualizada para:", file)
		},
	}

	return cmd
}

func listConfigFiles(baseDir string) []string {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return []string{}
	}

	var out []string

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		n := e.Name()
		if strings.EqualFold(n, "settings.json") || strings.EqualFold(n, "cache.json") {
			continue
		}

		if strings.HasSuffix(strings.ToLower(n), ".json") {
			baseName := strings.TrimSuffix(n, ".json")
			// Ignora arquivos de variáveis de ambiente
			if strings.EqualFold(baseName, "env") || strings.EqualFold(baseName, "envs") {
				continue
			}

			out = append(out, baseName)
		}
	}

	return out
}
