# my-scaner

Estrutura base profissional para utilitários de linha de comando (CLI) em **Go**, seguindo os padrões canônicos da comunidade Go (*Standard Go Project Layout*) e arquitetura desacoplada.

---

## 📁 Estrutura do Projeto

```text
my-scaner/
├── cmd/
│   └── my-scaner/
│       └── main.go                 # Ponto de entrada mínimo (thin entrypoint)
├── internal/                       # Código privado da aplicação (não importável externamente)
│   ├── buildinfo/                  # Metadados de build injetados via ldflags (versão, commit, data)
│   │   └── buildinfo.go
│   ├── cli/                        # Camada de interface CLI (Cobra)
│   │   ├── root.go                 # Comando raiz, flags globais e graceful shutdown
│   │   ├── scan.go                 # Subcomando de exemplo 'scan'
│   │   └── version.go              # Subcomando 'version' (suporte a texto e JSON)
│   ├── config/                     # Gestão de configurações com Viper
│   │   └── config.go
│   └── logger/                     # Logging estruturado com log/slog (JSON/Text)
│       └── logger.go
├── pkg/                            # Código público e reutilizável por outras aplicações
│   └── scanner/
│       ├── scanner.go              # Lógica de negócio e interfaces do scanner
│       └── scanner_test.go         # Testes de unidade com race detector
├── configs/
│   └── config.example.yaml         # Template de configuração
├── Makefile                        # Tarefas de compilação, teste e execução
├── go.mod
├── go.sum
└── .gitignore
```

---

## 🛠️ Padrões e Boas Práticas Adotados

1. **Standard Go Project Layout (`cmd/`, `internal/`, `pkg/`)**:
   - `cmd/`: Apenas arquivos `main.go` mínimos responsáveis por delegar a execução.
   - `internal/`: Protege pacotes internos de importações indevidas por outros módulos externos.
   - `pkg/`: Contém bibliotecas ou utilitários reutilizáveis projetados com contratos claros (`interfaces`).

2. **Cobra + Viper Integration**:
   - Resolução de configurações com hierarquia: **Flags CLI > Variáveis de Ambiente (`MY_SCANER_*`) > Arquivo de Configuração > Valores Padrão**.

3. **Injeção de Metadados de Compilação (`ldflags`)**:
   - O `Makefile` injeta versão (`git describe`), hash do commit e data UTC diretamente nas variáveis do pacote `internal/buildinfo`.

4. **Graceful Shutdown**:
   - O comando raiz captura sinais do sistema (`SIGINT`, `SIGTERM`) através de `signal.NotifyContext` e propaga o contexto (`cmd.Context()`), permitindo interrupções limpas de rotinas em andamento.

5. **Logging Estruturado (`log/slog`)**:
   - Uso da biblioteca padrão de logging estruturado do Go (1.21+), com suporte dinâmico a formatos `text` e `json`, e níveis (`debug`, `info`, `warn`, `error`).

6. **Inversão de Dependências (DIP / Interfaces)**:
   - Pacotes de domínio em `pkg/` expõem interfaces (ex: `scanner.Scanner`), permitindo mocks limpos e testes sem dependências externas complexas.

---

## 🚀 Como Utilizar

### Compilar o binário
```bash
make build
```
O binário será gerado em `./bin/my-scaner`.

### Executar testes
```bash
make test
```

### Visualizar ajuda e comandos
```bash
./bin/my-scaner --help
```

### Consultar versão
```bash
# Formato texto legível
./bin/my-scaner version

# Formato JSON (ideal para automações/CI)
./bin/my-scaner version --json
```

### Executar varredura (Exemplo)
```bash
# Varredura básica
./bin/my-scaner scan --host 127.0.0.1 --ports 80,443,22

# Com logs verbosos estruturados em JSON
./bin/my-scaner scan --host 127.0.0.1 --ports 80,443 --verbose --log-format json
```
