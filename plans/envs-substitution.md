# Plano: Substituição Automática de Variáveis de Ambiente

## Visão Geral

Implementar substituição automática de variáveis no JSON do teste usando valores de um arquivo de envs JSON com múltiplos ambientes.

## Fluxo

```mermaid
flowchart TD
    A[rabbix run teste --env ambientX] --> B[Carrega settings.json]
    B --> C[Verifica se envs_file está definido]
    C -->|Sim| D[Lê arquivo JSON de envs]
    C -->|Não| F[Lê JSON do teste]
    D --> E[Seleciona ambiente da flag --env]
    E --> G[Lê JSON do teste]
    G --> H[Busca padrões ${VAR} no JSON]
    H --> I[Substitui ${VAR} pelo valor do ambiente selecionado]
    I --> J[Executa teste com valores substituídos]
    F --> J
```

## Formato dos Arquivos

### settings.json
```json
{
  "sett": "local.json",
  "envs_file": "envs.json"
}
```

### envs.json (com múltiplos ambientes)
```json
{
  "ambientX": {
    "RABBITMQ_HOST": "amqp://localhost:5672",
    "API_KEY": "secret-key-123",
    "QUEUE_NAME": "my-queue"
  },
  "ambientey": {
    "RABBITMQ_HOST": "amqp://localhost:5624372",
    "API_KEY": "secret-key-1223",
    "QUEUE_NAME": "my-queu2e"
  }
}
```

### teste.json
```json
{
  "name": "meu-teste",
  "route_key": "test.route",
  "json_pool": {
    "host": "${RABBITMQ_HOST}",
    "api_key": "${API_KEY}"
  }
}
```

## Tarefas de Implementação

### 1. Adicionar campo envs_file em sett.go
- [ ] Adicionar `envs_file` aos defaults em [`getSettingsPath()`](pkg/sett/sett.go:35)
- [ ] Criar método `LoadEnvs(ambient string)` na interface `SettItf`
- [ ] Implementar leitura do arquivo de envs e seleção do ambiente

### 2. Criar função de substituição de variáveis
- [ ] Criar novo arquivo `pkg/run/envs.go`
- [ ] Implementar função `ReplaceEnvs(data []byte, envs map[string]string) []byte`
- [ ] Usar regex para encontrar padrões `${VAR}`
- [ ] Substituir apenas se a variável existir no envs

### 3. Integrar no fluxo de execução
- [ ] Adicionar flag `--env` em [`CmdRun()`](pkg/run/cmd.go:38)
- [ ] Carregar envs se `envs_file` estiver definido e flag `--env` fornecida
- [ ] Aplicar substituição no JSON antes do `json.Unmarshal`
- [ ] Logar ambiente e variáveis substituídas

### 4. Tratamento de erros
- [ ] Arquivo de envs não existe: warning e continua
- [ ] Ambiente não encontrado: erro e lista ambientes disponíveis
- [ ] Variável não encontrada: warning e mantém `${VAR}` original
- [ ] JSON de envs inválido: erro e para execução

## Exemplo de Uso

```bash
# Executar teste com ambiente específico
rabbix run meu-teste --env ambientX

# Executar sem ambiente (sem substituição)
rabbix run meu-teste
```

## Decisões Tomadas

| Decisão | Escolha |
|---------|---------|
| Sintaxe de referência | `${VAR}` |
| Flag para selecionar ambiente | `--env ambientX` |
| Formato do arquivo de envs | JSON com mapas por ambiente |
| Campo em settings | `envs_file` |
| Variável não encontrada | Mantém `${VAR}` original com warning |
| Ambiente não encontrado | Erro e lista ambientes disponíveis |
