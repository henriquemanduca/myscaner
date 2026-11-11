# Guia para Agentes de IA (AGENTS.md) - my-scaner

Este documento contém diretrizes, arquitetura e convenções técnicas para guiar agentes de IA e desenvolvedores ao contribuir ou estender o projeto **my-scaner**.

---

## 📌 Visão Geral do Projeto

**my-scaner** é uma ferramenta de linha de comando (CLI) desenvolvida em **Go** para extração de texto de imagens (OCR) estruturada em Markdown via **OpenRouter**, projetada com foco em alta performance, modularidade e manutenibilidade.

- **Linguagem**: Go 1.21+ (utilizando recursos modernos como `log/slog`).
- **CLI Framework**: [`spf13/cobra`](https://github.com/spf13/cobra).
- **Configuração**: [`spf13/viper`](https://github.com/spf13/viper) com suporte a YAML, variáveis de ambiente (`OPENROUTER_API_KEY`, `MY_SCANER_*`) e valores padrão.
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
├── pkg/                         # Código público e reutilizável
│   └── openrouter/              # Serviço de integração com OpenRouter e manipulação de imagem
│       ├── client.go            # Cliente HTTP, prompt de OCR e interface Extractor
│       ├── client_test.go       # Testes com httptest.Server
│       └── image.go             # Validação de extensões, encoding data URL e cálculo de path .md
├── configs/                     # Modelos de configuração (ex: config.example.yaml)
├── bin/                         # Binários compilados (ignorado no git)
├── install.sh                   # Script de instalação/desinstalação no Linux
├── Makefile                     # Automação de tarefas (build, test, run, clean, install, uninstall)
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
- **`pkg/`**: Destinado a domínios de negócio e regras reutilizáveis (ex: `pkg/openrouter`). Não deve depender de pacotes dentro de `internal/` nem do Cobra/Viper diretamente.

### 3. Inversão de Dependências (DIP) e Interfaces
- Sempre declare interfaces no pacote consumidor ou no pacote de domínio exportável (`openrouter.Extractor`).
- O cliente HTTP deve ser injetável via opções (`WithHTTPClient`), possibilitando mocks com `httptest.Server` nos testes.

### 4. Tratamento de Erros e Sinais (Graceful Shutdown)
- Propague sempre `context.Context` a partir de `cmd.Context()` até a camada HTTP do OpenRouter.
- Sinais `SIGINT` e `SIGTERM` cancelam requisições ativas de forma graciosa.
- Utilize embrulho de erros com `%w` (`fmt.Errorf("falha ao processar: %w", err)`).
- Não utilize `panic()` ou `os.Exit()` dentro de comandos Cobra; retorne erros no `RunE`.

### 5. Validação de Arquivos e Extensões
- Antes de disparar chamadas externas de rede, valide a existência do arquivo e restrinja as extensões aceitas (`.png`, `.jpg`, `.jpeg`, `.webp`, `.gif`, `.bmp`, `.tiff`).
- A saída `.md` gerada deve sempre residir no mesmo diretório com o mesmo nome base do arquivo de imagem original.

---

## 🛠️ Comandos Essenciais

| Comando | Descrição |
|---|---|
| `make build` | Compila o binário aplicando `-ldflags` para `buildinfo` em `./bin/my-scaner` |
| `make test` | Executa todos os testes unitários com race detector (`go test -race -v ./...`) |
| `make clean` | Remove o diretório `./bin` e artefatos de compilação |
| `make install` | Executa `./install.sh` para instalar binário, configs e autocompleções |
| `make uninstall` | Executa `./install.sh --uninstall` para remover o utilitário |
| `./bin/my-scaner scan -p <imagem>` | Extrai texto da imagem e gera o arquivo `.md` correspondente |
