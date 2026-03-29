package run

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/maxwelbm/rabbix/pkg/cache"
	"github.com/maxwelbm/rabbix/pkg/rabbix"
	"github.com/maxwelbm/rabbix/pkg/request"
	"github.com/maxwelbm/rabbix/pkg/sett"
	"github.com/spf13/cobra"
)

type Run struct {
	settings sett.SettItf
	Cache    cache.CacheItf
	request  request.RequestItf
}

func New(
	settings sett.SettItf,
	cache cache.CacheItf,
	request request.RequestItf,
) *Run {
	return &Run{
		settings: settings,
		Cache:    cache,
		request:  request,
	}
}

func (r *Run) CmdRun() *cobra.Command {
	var (
		quantity   int
		mockSpec   string
		envAmbient string
	)

	var cmd = &cobra.Command{
		Use:   "run [test-name]",
		Short: "Executa um caso de teste específico",
		Long: `Executa um caso de teste específico salvamento previamente.
Exemplo: rabbix run meu-teste`,
		Args: cobra.ExactArgs(1),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// Sincroniza cache antes de fornecer sugestões
			r.Cache.SyncCacheWithFileSystem()

			// Obtém lista de testes do cache
			cachedTests := r.Cache.GetCachedTests()

			return cachedTests, cobra.ShellCompDirectiveNoFileComp
		},
		Run: func(cmd *cobra.Command, args []string) {
			testName := args[0]

			// Validação: --env e --mock são mutuamente exclusivos
			if envAmbient != "" && strings.TrimSpace(mockSpec) != "" {
				fmt.Println("❌ Erro: As flags --env e --mock não podem ser usadas juntas")
				fmt.Println("💡 Use --env para substituir variáveis do ambiente ou --mock para gerar dados dinâmicos")

				return
			}

			// Carrega configuração para obter diretório de saída
			settings := r.settings.LoadSettings()

			outputDir := settings["output_dir"]
			if outputDir == "" {
				home, _ := os.UserHomeDir()
				outputDir = filepath.Join(home, ".rabbix", "tests")
			}

			// Lê o arquivo do teste
			testPath := filepath.Join(outputDir, testName+".json")

			data, err := os.ReadFile(testPath)
			if err != nil {
				fmt.Printf("❌ Erro: Teste '%s' não encontrado em %s\n", testName, testPath)
				fmt.Println("💡 Use 'rabbix list' para ver os testes disponíveis")

				return
			}

			// Substituição de variáveis de ambiente
			// Prioridade: --env > default_env da configuração
			envToUse := envAmbient
			if envToUse == "" {
				envToUse = settings["default_env"]
			}

			if envToUse != "" {
				envs, err := r.settings.LoadEnvs(envToUse)
				if err != nil {
					fmt.Printf("❌ Erro ao carregar ambiente '%s': %v\n", envToUse, err)

					ambients, listErr := r.settings.ListAmbients()
					if listErr == nil && len(ambients) > 0 {
						fmt.Println("📋 Ambientes disponíveis:")

						for _, a := range ambients {
							fmt.Printf("  - %s\n", a)
						}
					}

					return
				}

				if envs != nil {
					// Verifica variáveis faltantes antes de substituir
					missing := FindMissingEnvs(data, envs)
					if len(missing) > 0 {
						fmt.Printf("❌ Variáveis não encontradas no ambiente '%s': %v\n", envToUse, missing)
						fmt.Println("💡 Verifique o arquivo envs.json e adicione as variáveis necessárias")

						return
					}

					data = ReplaceEnvs(data, envs)

					fmt.Printf("🔧 Ambiente: %s\n", envToUse)
				}
			}

			var tc rabbix.TestCase
			if err := json.Unmarshal(data, &tc); err != nil {
				fmt.Printf("❌ Erro ao carregar JSON do teste '%s': %v\n", testName, err)
				return
			}

			// Garante que JSONPool exista
			if tc.JSONPool == nil {
				tc.JSONPool = map[string]any{}
			}

			// Parser do mockSpec -> []string de "campo:tipo"
			var mockPairs []string

			if strings.TrimSpace(mockSpec) != "" {
				// permite JSON array ou lista separada por vírgula
				trim := strings.TrimSpace(mockSpec)
				if strings.HasPrefix(trim, "[") {
					if err := json.Unmarshal([]byte(trim), &mockPairs); err != nil {
						fmt.Printf("⚠️  Não foi possível interpretar --mock como JSON array: %v\n", err)
						// tenta fallback por vírgulas removendo colchetes
						trim = strings.Trim(trim, "[]")
						if trim != "" {
							mockPairs = strings.Split(trim, ",")
						}
					}
				} else {
					mockPairs = strings.Split(trim, ",")
				}
				// limpeza de espaços e aspas
				for i := range mockPairs {
					mockPairs[i] = strings.Trim(mockPairs[i], " \"\n\t")
				}
			}

			if quantity <= 0 {
				quantity = 1
			}

			fmt.Printf("🚀 Executando teste: %s\n", tc.Name)
			fmt.Printf("📤 Route Key: %s\n", tc.RouteKey)

			if quantity > 1 {
				fmt.Printf("🔁 Quantidade: %d\n", quantity)
			}

			if len(mockPairs) > 0 {
				fmt.Printf("🧪 Mock: %v\n", mockPairs)
			}

			for i := 1; i <= quantity; i++ {
				// aplica mocks por iteração
				if len(mockPairs) > 0 {
					seed := time.Now().UnixNano() + int64(i)
					rng := rand.New(rand.NewSource(seed))

					for _, pair := range mockPairs {
						if pair == "" {
							continue
						}

						parts := strings.SplitN(pair, ":", 2)
						if len(parts) != 2 {
							fmt.Printf("⚠️  Par inválido em --mock: '%s' (esperado 'campo:tipo')\n", pair)
							continue
						}

						field := strings.TrimSpace(parts[0])
						typeName := strings.ToLower(strings.TrimSpace(parts[1]))

						var value any

						switch typeName {
						case "int":
							value = rng.Intn(1000000)
						case "float", "float64":
							value = rng.Float64() * 100000
						case "string":
							value = randomString(12, rng)
						case "time", "datetime", "date":
							value = time.Now().Format(time.RFC3339)
						case "bool", "boolean":
							value = rng.Intn(2) == 0
						default:
							fmt.Printf("⚠️  Tipo desconhecido '%s' para campo '%s'. Usando string.\n", typeName, field)

							value = randomString(8, rng)
						}
						// aplica no JSONPool
						tc.JSONPool[field] = value
					}
				}
				// Usa a função reutilizável PublishMessage
				resp, err := r.request.Request(tc)
				if err != nil {
					fmt.Printf("❌ [%d/%d] Erro ao enviar mensagem: %v\n", i, quantity, err)
					continue
				}

				func() {
					defer func() {
						if err := resp.Body.Close(); err != nil {
							fmt.Printf("❌ Erro ao fechar corpo da resposta: %v\n", err)
						}
					}()

					// Lê a resposta
					body, err := io.ReadAll(resp.Body)
					if err != nil {
						fmt.Printf("❌ [%d/%d] Erro ao ler resposta: %v\n", i, quantity, err)
						return
					}

					// Exibe o resultado
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						fmt.Printf("✅ [%d/%d] Mensagem enviada com sucesso! (Status: %d)\n", i, quantity, resp.StatusCode)
					} else {
						fmt.Printf("⚠️  [%d/%d] Resposta com status %d\n", i, quantity, resp.StatusCode)
					}

					fmt.Printf("📥 [%d/%d] Resposta do RabbitMQ:\n%s\n", i, quantity, string(body))
				}()
			}
		},
	}

	cmd.Flags().IntVarP(&quantity, "quantity", "n", 1,
		"Quantidade de vezes que o caso de teste será executado")
	cmd.Flags().StringVar(&mockSpec, "mock", "",
		"Array JSON ou lista separada por vírgulas de pares 'campo:tipo' para gerar dados dinâmicos")
	cmd.Flags().StringVar(&envAmbient, "env", "",
		"Nome do ambiente no arquivo de envs para substituir variáveis ${VAR} no JSON")

	return cmd
}

// randomString gera uma ‘string’ aleatória alfanumérica
func randomString(n int, rng *rand.Rand) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}

	return string(b)
}
