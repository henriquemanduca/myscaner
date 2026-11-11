# Guia para Agentes de IA (AGENTS.md) - my-scaner

Este documento contém diretrizes, arquitetura e convenções técnicas para guiar agentes de IA e desenvolvedores ao contribuir ou estender o projeto **my-scaner**.

---

## 📌 Visão Geral do Projeto

**my-scaner** é uma ferramenta de linha de comando (CLI) desenvolvida em **Go**, projetada para alta performance, modularidade e manutenibilidade.

- **Linguagem**: Go 1.21+ (utilizando recursos modernos como `log/slog`).
- **CLI Framework**: [`spf13/cobra`](https://github.com/spf13/cobra).
- **Configuração**: [`spf13/viper`](https://github.com/spf13/viper) com suporte a YAML, variáveis de ambiente e valores padrão.
- **Logging**: `log/slog` nativo (suporte a texto e JSON estruturado).
- **Layout**: Canônico *Standard Go Project Layout* (`cmd/`, `internal/`, `pkg/`).

---

## 🏗️ Arquitetura e Estrutura de Diretórios

```text
my-scaner/
├── cmd/
│   └── my-scaner/
│       └── main.go              # Entrypoint mínimo: apenas chama cli.Execute() e trata código de saída
├── internal/                    # Pacotes privados da aplicação (não importáveis externamente)
│   ├── buildinfo/               # Metadados de versão/build injetados via ldflags
│   ├── cli/                     # Definição e registro de comandos Cobra (root, scan, version)
│   ├── config/                  # Mapeamento e validação de configurações (Viper)
│   └── logger/                  # Configuração do slog.Logger (níveis e formatos text/json)
├── pkg/                         # Código público e reutilizável por terceiros ou outros binários
│   └── scanner/                 # Domínio de varredura: interfaces, structs e testes
├── configs/                     # Modelos de configuração (ex: config.example.yaml)
├── bin/                         # Binários compilados (ignorado no git)
├── Makefile                     # Automação de tarefas (build, test, run, clean)
├── go.mod
├── go.sum
└── .gitignore
```

---

## 📋 Regras e Diretrizes para Agentes

Ao propor ou implementar alterações, siga estritamente as seguintes regras:

### 1. Ponto de Entrada Mínimo (`cmd/`)
- Nunca coloque lógica de negócios ou inicialização complexa em `cmd/my-scaner/main.go`.
- `main.go` deve apenas chamar `cli.Execute()` e executar `os.Exit(1)` em caso de erro fatal.

### 2. Separação Estrita de Responsabilidades (`internal/` vs `pkg/`)
- **`internal/`**: Destinado exclusivamente a código de infraestrutura interna do binário CLI (ex: parsing de flags, binding do Viper, configuração do logger).
- **`pkg/`**: Destinado a domínios de negócio e regras reutilizáveis. Não deve depender de pacotes dentro de `internal/` nem do Cobra/Viper diretamente.

### 3. Inversão de Dependências (DIP) e Interfaces
- Sempre declare interfaces no pacote consumidor ou no pacote de domínio exportável (como em `pkg/scanner/scanner.go`).
- Evite acoplamento direto com recursos de I/O em métodos centrais de negócio; receba `context.Context`, `io.Reader`, `io.Writer` ou dependências via construtores (`New(...)`).

### 4. Tratamento de Erros e Sinais (Graceful Shutdown)
- Propague sempre `context.Context` a partir de `cmd.Context()`.
- O comando raiz já configura `signal.NotifyContext` para `SIGINT` e `SIGTERM`. Todas as operações demoradas (como varredura de portas, requisições de rede, etc.) devem escutar `<-ctx.Done()`.
- Utilize embrulho de erros com `%w` (`fmt.Errorf("falha ao processar: %w", err)`).
- Não utilize `panic()` ou `os.Exit()` dentro de comandos Cobra; retorne erros no `RunE` para permitir encerramento limpo.

### 5. Logging Estruturado com `log/slog`
- Não use `fmt.Println` para emitir logs ou mensagens de debug. Utilize `slog.Info`, `slog.Debug`, `slog.Warn`, `slog.Error`.
- `cmd.OutOrStdout()` e `cmd.ErrOrStderr()` devem ser utilizados para a saída direta destinada ao usuário final (dados de comando).

### 6. Configurações e Hierarquia do Viper
Ao adicionar novas opções de configuração:
1. Adicione o campo correspondente na struct `Config` em `internal/config/config.go`.
2. Configure um valor padrão (`SetDefault`).
3. Mantenha suporte a variáveis de ambiente com o prefixo `MY_SCANER_`.
4. Atualize o arquivo `configs/config.example.yaml`.

---

## 🛠️ Comandos Essenciais

| Comando | Descrição |
|---|---|
| `make build` | Compila o binário aplicando `-ldflags` para `buildinfo` em `./bin/my-scaner` |
| `make test` | Executa todos os testes unitários com race detector (`go test -race -v ./...`) |
| `make run ARGS="..."` | Executa a aplicação diretamente com argumentos (`go run cmd/my-scaner/main.go ...`) |
| `make clean` | Remove o diretório `./bin` e artefatos de compilação |

---

## ➕ Como Adicionar um Novo Subcomando

1. Crie um arquivo em `internal/cli/<comando>.go`.
2. Defina uma função construtora `new<Comando>Cmd() *cobra.Command`.
3. Registre as flags locais utilizando `cmd.Flags()`.
4. Adicione o comando ao `rootCmd` em `internal/cli/root.go`:
   ```go
   rootCmd.AddCommand(new<Comando>Cmd())
   ```
5. Crie testes de unidade para as regras de negócio em `pkg/` ou testes de CLI correspondentes.
6. Atualize o `README.md` se adicionar flags ou comandos relevantes para o usuário final.
