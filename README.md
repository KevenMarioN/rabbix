# 🐇 rabbix

CLI para testar filas RabbitMQ de forma organizada e reutilizável durante o desenvolvimento.

## 🚀 Início Rápido

### 1. Criar/selecionar perfil de configuração

```bash
rabbix conf select meu-projeto
```

Este comando cria automaticamente um perfil em `~/.rabbix/meu-projeto.json` se não existir, com configurações padrão:
- Host: `http://localhost:15672`
- Auth: `guest:guest` (base64)
- Diretório de testes: `~/.rabbix/meu-projeto/`
- Arquivo de variáveis: `~/.rabbix/meu-projeto/envs.json`

### 2. Configurar credenciais (opcional)

```bash
rabbix conf set --host http://localhost:15672 --user admin --password secret
```

### 3. Verificar conexão

```bash
rabbix health
```

### 4. Criar arquivo de teste

Os testes ficam em `~/.rabbix/<perfil>/<nome>.json`:

```json
{
  "name": "teste-pedido",
  "route_key": "fila-pedidos",
  "json_pool": {
    "id": 123,
    "nome": "teste"
  }
}
```

### 5. Executar teste

```bash
rabbix run teste-pedido
```

## ⚙️ Instalação

Requer Go 1.23 ou superior.

```bash
go install github.com/maxwelbm/rabbix@latest
```

Para autocomplete, veja [AUTOCOMPLETE.md](AUTOCOMPLETE.md).

## 📖 Comandos

### `rabbix conf`

Gerencia configurações e perfis da CLI.

| Subcomando | Descrição |
|------------|-----------|
| `select [nome]` | Seleciona ou cria um perfil de configuração |
| `set` | Define configurações do perfil ativo |
| `get` | Exibe configuração atual |

**`conf select`:**

Cria um novo perfil se não existir, ou ativa um perfil existente.

```bash
# Criar/ativar perfil
rabbix conf select meu-projeto

# Ver perfis disponíveis
rabbix conf select
```

Saída quando não especifica nome:
```
Informe o nome da configuração. Disponíveis:
- meu-projeto.json
- producao.json
```

**Flags do `conf set`:**

| Flag | Descrição | Exemplo |
|------|-----------|---------|
| `--host` | URL do RabbitMQ Management API | `http://localhost:15672` |
| `--user` | Usuário do RabbitMQ | `guest` |
| `--password` | Senha do RabbitMQ | `guest` |
| `--output` | Diretório para salvar testes | `~/.rabbix/tests` |
| `--env` | Ambiente padrão | `local`, `dev`, `prod` |

**Exemplos:**

```bash
# Configurar host e credenciais
rabbix conf set --host http://localhost:15672 --user admin --password secret

# Definir ambiente padrão
rabbix conf set --env local

# Ver configuração atual
rabbix conf get
```

---

### `rabbix health`

Verifica se a API do RabbitMQ está acessível.

```bash
rabbix health
```

Saída esperada:
```
🔍 Verificando saúde da API...
📡 URL: http://localhost:15672/api/overview
📊 Status: 200 OK
✅ API está funcionando corretamente!
```

---

### `rabbix run`

Executa um caso de teste específico.

```bash
rabbix run <nome-do-teste> [flags]
```

**Flags:**

| Flag | Tipo | Descrição |
|------|------|-----------|
| `--quantity` | `int` | Número de execuções (default: 1) |
| `--mock` | `string` | Gera dados dinâmicos (ex: `id:uuid,nome:name`) |
| `--env` | `string` | Ambiente para substituir variáveis |

**Exemplos:**

```bash
# Executar teste simples
rabbix run meu-teste

# Executar 10 vezes
rabbix run meu-teste --quantity 10

# Com dados dinâmicos
rabbix run meu-teste --mock "id:uuid,email:email"

# Com ambiente específico
rabbix run meu-teste --env dev
```

**Mock disponíveis:**

| Tipo | Gera |
|------|------|
| `uuid` | UUID v4 |
| `name` | Nome aleatório |
| `email` | Email aleatório |
| `int` | Número inteiro aleatório |
| `float` | Número decimal aleatório |
| `bool` | Booleano aleatório |

---

### `rabbix batch`

Executa múltiplos testes em paralelo.

```bash
rabbix batch <teste1> <teste2> ... [flags]
```

**Flags:**

| Flag | Tipo | Descrição |
|------|------|-----------|
| `--concurrency` | `int` | Número de workers paralelos (default: 3) |
| `--delay` | `int` | Delay entre execuções em ms (default: 500) |
| `--all` | `bool` | Executa todos os testes disponíveis |

**Exemplos:**

```bash
# Executar testes específicos
rabbix batch teste1 teste2 teste3

# Com concorrência customizada
rabbix batch teste1 teste2 --concurrency 5 --delay 1000

# Executar todos os testes
rabbix batch --all
```

---

### `rabbix list`

Lista todos os casos de teste salvos.

```bash
rabbix list
```

Saída:
```
📄 Casos de teste:
🧪 teste-pedido.json  (routeKey: fila-pedidos)
🧪 teste-pagamento.json  (routeKey: fila-pagamentos)
```

---

### `rabbix cache`

Gerencia o cache de autocomplete.

| Subcomando | Descrição |
|------------|-----------|
| `stats` | Exibe estatísticas do cache |
| `sync` | Sincroniza cache com arquivos |
| `clear` | Limpa o cache |

**Exemplos:**

```bash
rabbix cache stats
rabbix cache sync
rabbix cache clear
```

## 📁 Estrutura de Arquivos

### Diretório padrão

```
~/.rabbix/
├── settings.json          # Configurações globais (perfil ativo)
├── cache.json             # Cache para autocomplete
├── meu-projeto.json       # Configuração do perfil "meu-projeto"
├── producao.json          # Configuração do perfil "producao"
└── meu-projeto/           # Testes do perfil "meu-projeto"
    ├── teste1.json
    ├── teste2.json
    └── envs.json          # Variáveis de ambiente
```

### Formato do arquivo de teste

```json
{
  "name": "nome-do-teste",
  "route_key": "nome-da-fila",
  "json_pool": {
    "campo1": "valor1",
    "campo2": "valor2"
  }
}
```

### Variáveis de ambiente

Crie `~/.rabbix/<perfil>/envs.json` para substituir placeholders:

```json
{
  "local": {
    "API_URL": "http://localhost:3000",
    "DB_NAME": "test_db"
  },
  "dev": {
    "API_URL": "https://dev.api.com",
    "DB_NAME": "dev_db"
  }
}
```

No payload, use `${VAR_NAME}`:

```json
{
  "name": "teste-com-env",
  "route_key": "minha-fila",
  "json_pool": {
    "api_url": "${API_URL}",
    "database": "${DB_NAME}"
  }
}
```

## 🔧 Desenvolvimento

### Build local

```bash
make all
```

### Lint

```bash
make lint
```

## 📜 Licença

[MIT](LICENSE) License © Maxwel Mazur
