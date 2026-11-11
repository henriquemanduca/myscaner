# my-scaner

Ferramenta CLI profissional em **Go** para extração de texto de imagens (OCR) via **OpenRouter** e estruturação automática em formato **Markdown (.md)**, seguindo os padrões canônicos da comunidade (*Standard Go Project Layout*) e arquitetura desacoplada.

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
│   │   ├── scan.go                 # Comando 'scan' (extração OCR e geração de .md)
│   │   └── version.go              # Subcomando 'version' (suporte a texto e JSON)
│   ├── config/                     # Gestão de configurações com Viper
│   │   └── config.go
│   └── logger/                     # Logging estruturado com log/slog (JSON/Text)
│       └── logger.go
├── pkg/                            # Código público e reutilizável por outras aplicações
│   └── openrouter/                 # Serviço de integração OpenRouter e processamento de imagem
│       ├── client.go               # Cliente HTTP multimodal e interface Extractor
│       ├── client_test.go          # Testes de unidade com httptest
│       └── image.go                # Validação de extensões, encoding data URL e caminhos
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
   - `pkg/openrouter`: Expõe o contrato `Extractor` e o cliente de comunicação via API.

2. **Cobra + Viper Integration**:
   - Hierarquia de resolução de credenciais e modelo: **Flags CLI (`--api-key`, `--model`) > Variáveis de Ambiente (`OPENROUTER_API_KEY`, `MY_SCANER_*`) > Arquivo de Configuração > Valores Padrão**.

3. **Injeção de Metadados de Compilação (`ldflags`)**:
   - O `Makefile` injeta versão Git, hash do commit e data UTC diretamente nas variáveis de `buildinfo`.

4. **Graceful Shutdown**:
   - Captura de sinais do sistema (`SIGINT`, `SIGTERM`) propagando `context.Context` para cancelar requisições HTTP em andamento.

5. **Logging Estruturado (`log/slog`)**:
   - Suporte dinâmico a `--verbose` e formatos de log `--log-format text|json`.

---

## 🚀 Instalação e Utilização

### Instalação no Linux

O repositório inclui o script [`install.sh`](file:///home/henrique/projetos/pessoal/my-scaner/install.sh) para automatizar a compilação, instalação do binário, configuração e autocompleções de shell (Bash, Zsh e Fish).

#### 1. Instalação rápida para o usuário (`~/.local/bin`)
Não requer privilégios de administrador (`sudo`):
```bash
./install.sh
# ou via Makefile
make install
```

#### 2. Instalação global no sistema (`/usr/local/bin`)
```bash
sudo ./install.sh --system
```

#### 3. Opções avançadas do instalador
```bash
./install.sh --prefix /opt/my-scaner   # Diretório customizado
./install.sh --no-config               # Não cria arquivo de configuração padrão
./install.sh --no-completion           # Pula registro de compleções no shell
./install.sh --uninstall               # Desinstala o binário e autocompleções
```

---

### Execução Local (sem instalar)
```bash
make build
```
O binário será gerado em `./bin/my-scaner`.

### 2. Executar testes
```bash
make test
```

### 3. Configurar a Chave do OpenRouter
Você pode definir a chave via variável de ambiente:
```bash
export OPENROUTER_API_KEY="sk-or-v1-..."
```
Ou criar um arquivo `config.yaml` baseado em `configs/config.example.yaml`.

### 4. Executar o comando `scan`
O comando `scan` exige a flag `-p, --path` apontando para a imagem. O arquivo `.md` será gravado no mesmo diretório com o mesmo nome base da imagem:

```bash
# Extração básica
./bin/my-scaner scan -p /caminho/para/documento.png

# O arquivo extraído será salvo em:
# /caminho/para/documento.md

# Usando com modelo específico e chave de API direta:
./bin/my-scaner scan -p ./nota.jpg --model anthropic/claude-3.5-sonnet --api-key sk-or-v1-...
```

### Formatos de imagem suportados:
- `.png`
- `.jpg`, `.jpeg`
- `.webp`
- `.gif`
- `.bmp`
- `.tiff`, `.tif`
