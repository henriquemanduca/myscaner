# my-scaner

Ferramenta CLI profissional em **Go** para extração de texto de documentos **PDF** e **imagens** (OCR) via inteligência artificial com suporte a **OpenRouter** (nuvem) e **LM Studio** (local/offline), estruturando o resultado de forma automática e limpa em formato **Markdown (.md)**, seguindo os padrões canônicos da comunidade (*Standard Go Project Layout*) e arquitetura desacoplada.

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
│   │   ├── scan.go                 # Comando 'scan' (extração OCR de imagens/PDF e geração de .md)
│   │   └── version.go              # Subcomando 'version' (suporte a texto e JSON)
│   ├── config/                     # Gestão de configurações com Viper
│   │   ├── config.go
│   │   └── config_test.go
│   └── logger/                     # Logging estruturado com log/slog (JSON/Text)
│       └── logger.go
├── pkg/                            # Código público e reutilizável por outras aplicações
│   ├── fileutil/                   # Utilitários compartilhados de validação, detecção e base64
│   │   └── fileutil.go
│   ├── llm/                        # Abstração comum de clientes compatíveis com a API OpenAI
│   │   ├── client.go               # Cliente base reutilizado por OpenRouter e LM Studio
│   │   └── extractor.go            # Contrato comum Extractor
│   ├── lmstudio/                   # Cliente para LM Studio local
│   │   ├── client.go
│   │   └── client_test.go
│   ├── openrouter/                 # Cliente para OpenRouter (nuvem)
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── image.go
│   └── pdf/                        # Domínio de conversão e extração de PDF
│       ├── converter.go            # Interface Converter e renderizador poppler (pdftoppm)
│       └── converter_test.go       # Testes unitários do conversor de PDF
├── configs/
│   └── config.example.yaml         # Template de configuração
├── install.sh                      # Script de instalação/desinstalação para Linux
├── Makefile                        # Tarefas de compilação, teste e execução
├── AGENTS.md                       # Diretrizes de desenvolvimento para agentes de IA
├── go.mod
├── go.sum
└── .gitignore
```

---

## 🛠️ Padrões e Boas Práticas Adotados

1. **Standard Go Project Layout (`cmd/`, `internal/`, `pkg/`)**:
   - `cmd/`: Apenas `main.go` mínimo chamando `cli.Execute()`.
   - `internal/`: Protege componentes da CLI de acoplamento externo.
   - `pkg/llm`: Núcleo compartilhado que elimina duplicação de código entre provedores compatíveis com a especificação OpenAI.
   - `pkg/openrouter`, `pkg/lmstudio` e `pkg/pdf`: Implementações e interfaces reutilizáveis.

2. **Suporte a Provedores Híbridos (Nuvem e Local)**:
   - **OpenRouter**: Modelos de ponta em nuvem (`google/gemini-2.0-flash-001`, `anthropic/claude-3.5-sonnet`, etc.).
   - **LM Studio**: Execução 100% local e privada em `http://localhost:1234/v1` (sem envio de dados para servidores externos).

3. **Suporte Completo a PDF**:
   - Conversão página a página em alta resolução (DPI configurável) via `poppler-utils` (`pdftoppm`).
   - Ciclo de vida efêmero seguro com deleção automática das imagens temporárias.
   - Consolidação de documentos de múltiplas páginas em um único arquivo `.md`.

4. **Injeção de Metadados de Compilação (`ldflags`)**:
   - O `Makefile` injeta versão Git, hash do commit e data UTC diretamente nas variáveis de `buildinfo`.

5. **Graceful Shutdown**:
   - Captura de sinais do sistema (`SIGINT`, `SIGTERM`) propagando `context.Context` para cancelar requisições HTTP e subprocessos em andamento.

---

## 🚀 Instalação e Utilização

### Instalação no Linux

```bash
# Instalação rápida no diretório do usuário (~/.local/bin)
./install.sh
# ou via Makefile
make install
```

---

### Configuração

O `my-scaner` aceita **exclusivamente arquivos YAML**. A configuração é buscada, nesta ordem, em:

1. `./config.yaml` (diretório atual)
2. `~/.config/my-scaner/config.yaml`
3. `/etc/my-scaner/config.yaml`

Use [`configs/config.example.yaml`](configs/config.example.yaml) como modelo. Também é possível definir variáveis de ambiente com o prefixo `MY_SCANER_` (ex.: `MY_SCANER_PROVIDER`, `MY_SCANER_OPENROUTER_API_KEY`) ou usar a flag `--config` para indicar um arquivo YAML específico.

---

### Execução do Comando `scan`

O comando `scan` aceita imagens e PDFs, permitindo alternar facilmente entre nuvem e local:

#### 1. Usando OpenRouter (Padrão)
```bash
export OPENROUTER_API_KEY="sk-or-v1-..."
./bin/my-scaner scan -p /caminho/para/documento.pdf
```

#### 2. Usando LM Studio Local (100% Offline / Privado)
```bash
# Certifique-se de que o servidor local do LM Studio está iniciado na porta 1234
./bin/my-scaner scan -p /caminho/para/documento.png --provider lmstudio

# Para PDF com modelo local específico
./bin/my-scaner scan -p ./relatorio.pdf --provider lmstudio --model qwen2.5-vl-7b-instruct
```

#### 3. Preservar imagens temporárias das páginas do PDF (opcional)
```bash
./bin/my-scaner scan -p ./documento.pdf --keep-images
```

#### 4. Flags disponíveis no comando `scan`:
- `-p, --path`: Caminho do arquivo de imagem ou PDF (**obrigatório**).
- `--provider`: Provedor de IA (`openrouter` ou `lmstudio`).
- `-m, --model`: Nome ou identificador do modelo.
- `--api-key`: Chave de API.
- `--base-url`: Endpoint customizado da API.
- `--dpi`: Resolução em DPI para páginas de PDF (padrão: 150).
- `--keep-images`: Salva as páginas convertidas em pasta local.
