package health

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/maxwelbm/rabbix/pkg/sett"
	"github.com/spf13/cobra"
)

func CmdHealth(settings sett.SettItf) *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Verifica o status de saúde da API do RabbitMQ",
		Long: `Faz uma requisição para o endpoint /api/overview para verificar se a API do "+
"RabbitMQ está funcionando corretamente.`,
		Run: func(cmd *cobra.Command, args []string) {
			settings := settings.LoadSettings()

			var auth = settings["auth"]
			if auth == "" {
				fmt.Printf("necessario configurar user e password com o comando " +
					"'rabbix conf set --user <user> --password <password>'\n")

				return
			}

			auth = "Basic " + auth

			var host = "http://localhost:15672" // host default
			if settings["host"] != "" {
				host = settings["host"]
			}

			url := strings.TrimRight(host, "/") + "/api/overview"

			fmt.Printf("🔍 Verificando saúde da API...\n")
			fmt.Printf("📡 URL: %s\n", url)

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				fmt.Printf("❌ Erro ao criar requisição: %v\n", err)
				return
			}

			req.Header.Add("Authorization", auth)

			client := &http.Client{}

			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("❌ Erro ao fazer requisição: %v\n", err)
				return
			}

			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("❌ Erro ao fechar corpo da resposta: %v\n", err)
				}
			}()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("❌ Erro ao ler resposta: %v\n", err)
				return
			}

			fmt.Printf("📊 Status: %s\n", resp.Status)

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				fmt.Printf("✅ API está funcionando corretamente!\n")
			} else {
				fmt.Printf("⚠️  API retornou status de erro\n")
			}

			fmt.Printf("📄 Resposta:\n%s\n", string(body))
		},
	}
}
