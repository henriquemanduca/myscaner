# Guia para Agentes de IA (AGENTS.md) - my-scaner

Este documento contém diretrizes, arquitetura e convenções técnicas para guiar agentes de IA e desenvolvedores ao contribuir ou estender o projeto **my-scaner**.

---

## 📌 Visão Geral do Projeto

**my-scaner** é uma ferramenta de linha de comando (CLI) desenvolvida em **Go** para extração de texto de documentos **PDF** e **imagens** (OCR) estruturada em Markdown via **OpenRouter** (nuvem) ou **LM Studio** (local/offline), projetada com foco em alta performance, modularidade e manutenibilidade.

- **Linguagem**: Go 1.21+ (utilizando recursos modernos como `log/slog`).
- **CLI Framework**: [`spf13/cobra`](https://github.com/spf13/cobra).
- **Configuração**: [`spf13/viper`](https://github.com/spf13/viper) com suporte exclusivo a arquivos **YAML** (`config.yaml`), variáveis de ambiente (`OPENROUTER_API_KEY`, `MY_SCANER_*`) e valores padrão.
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
│   │   ├── config.go
│   │   └── config_test.go
│   └── logger/                  # Configuração do slog.Logger (níveis e formatos text/json)
├── pkg/                         # Código público e reutilizável
│   ├── fileutil/                # Validação de extensões, detecção de tipos, Base64 data URL e cálculo de path
│   │   └── fileutil.go
│   ├── llm/                     # Núcleo compartilhado de clientes compatíveis com API OpenAI
│   │   ├── client.go            # OpenAICompatibleClient (payloads, headers, OCR prompt, sanitização)
│   │   └── extractor.go         # Contrato comum Extractor
│   ├── lmstudio/                # Cliente para servidor LM Studio local
│   │   ├── client.go
│   │   └── client_test.go
│   ├── openrouter/              # Cliente para OpenRouter (nuvem)
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── image.go             # Reexportações para retrocompatibilidade
│   └── pdf/                     # Domínio de conversão e extração de PDF
│       ├── converter.go         # Interface Converter e renderizador poppler (pdftoppm)
│       └── converter_test.go    # Testes unitários do conversor de PDF
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
- **`pkg/`**: Destinado a domínios de negócio e regras reutilizáveis (`pkg/llm`, `pkg/openrouter`, `pkg/lmstudio`, `pkg/pdf`, `pkg/fileutil`). Não deve depender de pacotes dentro de `internal/` nem do Cobra/Viper diretamente.

### 3. Reaproveitamento de Código entre Provedores LLM
- Tanto o OpenRouter quanto o LM Studio implementam o formato de endpoints compatíveis com a OpenAI (`/chat/completions`).
- Qualquer alteração na lógica de montagem de payload, prompt base de OCR, tratamento de Data URL ou sanitização de Markdown deve ser feita no pacote compartilhado **`pkg/llm`**, evitando duplicação.
- Clientes específicos (`openrouter.Client`, `lmstudio.Client`) devem apenas configurar endpoints, headers e timeouts próprios e implementar a interface `llm.Extractor`.

### 4. Tratamento de Erros e Sinais (Graceful Shutdown)
- Propague sempre `context.Context` a partir de `cmd.Context()` até as chamadas de rede e comandos externos.
- Sinais `SIGINT` e `SIGTERM` cancelam requisições ativas de forma graciosa.
- Utilize embrulho de erros com `%w` (`fmt.Errorf("falha ao processar: %w", err)`).
- Não utilize `panic()` ou `os.Exit()` dentro de comandos Cobra; retorne erros no `RunE`.

### 5. Validação de Arquivos e Extensões
- Valide a existência do arquivo antes de iniciar qualquer conversão usando `fileutil.ValidateInputFile`. Formatos aceitos: `.pdf` e imagens (`.png`, `.jpg`, `.jpeg`, `.webp`, `.gif`, `.bmp`, `.tiff`, `.tif`).
- O arquivo `.md` consolidado deve sempre residir no mesmo diretório com o mesmo nome base do arquivo de entrada original.
- Para PDFs, as imagens temporárias devem ser limpas automaticamente ao final do processo via `defer os.RemoveAll(tempDir)`.

---

## 🛠️ Comandos Essenciais

| Comando | Descrição |
|---|---|
| `make build` | Compila o binário aplicando `-ldflags` para `buildinfo` em `./bin/my-scaner` |
| `make test` | Executa todos os testes unitários com race detector (`go test -race -v ./...`) |
| `make clean` | Remove o diretório `./bin` e artefatos de compilação |
| `make install` | Executa `./install.sh` para instalar binário, configs e autocompleções |
| `make uninstall` | Executa `./install.sh --uninstall` para remover o utilitário |
| `./bin/my-scaner scan -p <arquivo>` | Extrai texto de imagem ou PDF via OpenRouter |
| `./bin/my-scaner scan -p <arquivo> --provider lmstudio` | Extrai texto de imagem ou PDF via LM Studio local |
