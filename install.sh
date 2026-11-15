#!/usr/bin/env bash
# ==============================================================================
# Script de Instalação e Gerenciamento do scaner-cli para Sistemas Linux
# ==============================================================================

set -euo pipefail

APP_NAME="scaner-cli"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Configuração de cores
if [[ -t 1 ]] && [[ -z "${NO_COLOR:-}" ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    RESET='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    CYAN=''
    BOLD=''
    RESET=''
fi

info()    { echo -e "${CYAN}ℹ${RESET} $*"; }
success() { echo -e "${GREEN}✔${RESET} $*"; }
warn()    { echo -e "${YELLOW}⚠${RESET} $*"; }
error()   { echo -e "${RED}✖${RESET} $*" >&2; }
step()    { echo -e "\n${BOLD}${BLUE}==>${RESET} ${BOLD}$*${RESET}"; }

# Valores padrão de opções
ACTION="install"
INSTALL_MODE="auto"
BIN_DIR=""
CONFIG_DIR=""
INSTALL_CONFIG=true
INSTALL_COMPLETIONS=true
REBUILD=false
ASSUME_YES=false

show_help() {
    cat <<EOF
Uso: $(basename "$0") [OPÇÕES]

Script para instalação e desinstalação automatizada do ${APP_NAME} em sistemas Linux.

OPÇÕES DE INSTALAÇÃO:
  -u, --user              Instalação local para o usuário atual (~/.local/bin) [Padrão para não-root]
  -s, --system            Instalação em nível de sistema (/usr/local/bin) [Padrão para root/sudo]
  -p, --prefix <DIR>      Define um diretório base (<DIR>/bin e <DIR>/etc/scaner-cli)
  -b, --bindir <DIR>      Define explicitamente o diretório do binário
  -c, --config-dir <DIR>  Define explicitamente o diretório de configurações
      --no-config         Não cria o arquivo de configuração de exemplo
      --no-completion     Ignora a instalação de autocompleções para o shell
  -r, --rebuild           Força a recompilação do binário antes de instalar
  -y, --yes               Assume 'sim' para todas as confirmações (não interativo)

AÇÕES:
      --uninstall         Desinstala o binário, compleções e opcionalmente configurações
  -h, --help              Exibe esta mensagem de ajuda

EXEMPLOS:
  ./install.sh                    # Instala no escopo do usuário atual (~/.local/bin)
  sudo ./install.sh --system      # Instala globalmente no sistema (/usr/local/bin)
  ./install.sh --prefix /opt/app  # Instala sob /opt/app/bin
  ./install.sh --uninstall        # Remove os arquivos instalados
EOF
}

# Processamento de parâmetros CLI
while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)
            show_help
            exit 0
            ;;
        -u|--user)
            INSTALL_MODE="user"
            shift
            ;;
        -s|--system)
            INSTALL_MODE="system"
            shift
            ;;
        -p|--prefix)
            INSTALL_MODE="custom"
            BIN_DIR="$2/bin"
            CONFIG_DIR="$2/etc/$APP_NAME"
            shift 2
            ;;
        -b|--bindir)
            BIN_DIR="$2"
            shift 2
            ;;
        -c|--config-dir)
            CONFIG_DIR="$2"
            shift 2
            ;;
        --no-config)
            INSTALL_CONFIG=false
            shift
            ;;
        --no-completion)
            INSTALL_COMPLETIONS=false
            shift
            ;;
        -r|--rebuild)
            REBUILD=true
            shift
            ;;
        -y|--yes)
            ASSUME_YES=true
            shift
            ;;
        --uninstall)
            ACTION="uninstall"
            shift
            ;;
        *)
            error "Opção desconhecida: $1"
            echo "Execute '$0 --help' para visualizar as opções disponíveis."
            exit 1
            ;;
    esac
done

# Validação do Sistema Operacional
check_os() {
    local os
    os="$(uname -s)"
    if [[ "$os" != "Linux" ]]; then
        error "Sistema operacional não suportado: $os. Este script suporta apenas Linux."
        exit 1
    fi
}

# Resolução de diretórios conforme o modo
resolve_directories() {
    if [[ "$INSTALL_MODE" == "auto" ]]; then
        if [[ $EUID -eq 0 ]]; then
            INSTALL_MODE="system"
        else
            INSTALL_MODE="user"
        fi
    fi

    if [[ "$INSTALL_MODE" == "system" ]]; then
        BIN_DIR="${BIN_DIR:-/usr/local/bin}"
        CONFIG_DIR="${CONFIG_DIR:-/etc/$APP_NAME}"
    elif [[ "$INSTALL_MODE" == "user" ]]; then
        BIN_DIR="${BIN_DIR:-${XDG_BIN_HOME:-$HOME/.local/bin}}"
        CONFIG_DIR="${CONFIG_DIR:-${XDG_CONFIG_HOME:-$HOME/.config}/$APP_NAME}"
    else
        # Modo customizado (prefix/bindir definidos)
        BIN_DIR="${BIN_DIR:-/usr/local/bin}"
        CONFIG_DIR="${CONFIG_DIR:-/etc/$APP_NAME}"
    fi
}

# Utilitário para executar comandos com elevação (sudo) apenas se necessário
need_sudo() {
    local target_dir="$1"
    if [[ -d "$target_dir" ]]; then
        [[ ! -w "$target_dir" ]]
    else
        local parent="$target_dir"
        while [[ ! -d "$parent" && "$parent" != "/" ]]; do
            parent="$(dirname "$parent")"
        done
        [[ ! -w "$parent" ]]
    fi
}

run_cmd() {
    local target_dir="$1"
    shift
    if need_sudo "$target_dir" && [[ $EUID -ne 0 ]]; then
        if command -v sudo >/dev/null 2>&1; then
            sudo "$@"
        else
            error "Permissão negada para acessar $target_dir e 'sudo' não está disponível."
            exit 1
        fi
    else
        "$@"
    fi
}

# Compilação do binário
build_binary() {
    step "Verificando compilação do binário"

    local binary_path="$SCRIPT_DIR/bin/$APP_NAME"

    if [[ -f "$binary_path" ]] && [[ "$REBUILD" == false ]]; then
        info "Binário pré-existente encontrado em: $binary_path"
        return 0
    fi

    if ! command -v go >/dev/null 2>&1; then
        if [[ -f "$binary_path" ]]; then
            warn "Go não encontrado no PATH, utilizando binário pré-compilado em: $binary_path"
            return 0
        else
            error "O compilador 'go' não foi encontrado no sistema e nenhum binário foi localizado em bin/$APP_NAME."
            error "Instale o Go 1.27+ (https://golang.org/dl/) para compilar a partir do código-fonte."
            exit 1
        fi
    fi

    info "Compilando $APP_NAME..."
    if command -v make >/dev/null 2>&1 && [[ -f "$SCRIPT_DIR/Makefile" ]]; then
        make -C "$SCRIPT_DIR" build
    else
        local version commit date ldflags
        version="$(git -C "$SCRIPT_DIR" describe --tags --always --dirty 2>/dev/null || echo "v0.1.0-dev")"
        commit="$(git -C "$SCRIPT_DIR" rev-parse --short HEAD 2>/dev/null || echo "none")"
        date="$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
        ldflags="-w -s -X 'scaner-cli/internal/buildinfo.Version=${version}' -X 'scaner-cli/internal/buildinfo.Commit=${commit}' -X 'scaner-cli/internal/buildinfo.Date=${date}'"

        mkdir -p "$SCRIPT_DIR/bin"
        go build -C "$SCRIPT_DIR" -ldflags "$ldflags" -o "bin/$APP_NAME" ./cmd/scaner-cli/main.go
    fi

    success "Binário compilado com sucesso em: $binary_path"
}

# Instalação do binário
install_binary() {
    step "Instalando binário"
    local binary_src="$SCRIPT_DIR/bin/$APP_NAME"

    info "Diretório de destino: $BIN_DIR"
    run_cmd "$BIN_DIR" mkdir -p "$BIN_DIR"
    run_cmd "$BIN_DIR" install -m 755 "$binary_src" "$BIN_DIR/$APP_NAME"

    success "Binário instalado em: $BIN_DIR/$APP_NAME"

    # Verificação de PATH
    if [[ ":$PATH:" != *":$BIN_DIR:"* ]]; then
        warn "O diretório '$BIN_DIR' NÃO está presente no seu PATH atual."
        echo -e "${YELLOW}Para torná-lo acessível globalmente, adicione a seguinte linha ao seu shell (~/.bashrc ou ~/.zshrc):${RESET}"
        echo -e "    ${BOLD}export PATH=\"\$PATH:$BIN_DIR\"${RESET}\n"
    fi
}

# Instalação do arquivo de configuração (YAML)
install_configuration() {
    if [[ "$INSTALL_CONFIG" == false ]]; then
        info "Instalação de arquivo de configuração ignorada (--no-config)."
        return 0
    fi

    step "Configurando arquivo de ambiente"
    local example_yaml="$SCRIPT_DIR/configs/config.example.yaml"
    local target_yaml="$CONFIG_DIR/config.yaml"

    run_cmd "$CONFIG_DIR" mkdir -p "$CONFIG_DIR"

    if [[ -f "$target_yaml" ]]; then
        info "Arquivo de configuração existente mantido: $target_yaml"
    elif [[ -f "$example_yaml" ]]; then
        run_cmd "$CONFIG_DIR" cp "$example_yaml" "$target_yaml"
        run_cmd "$CONFIG_DIR" chmod 600 "$target_yaml"
        success "Modelo de configuração YAML instalado em: $target_yaml"
        info "Edite este arquivo para personalizar provedor, modelo e chave de API."
    fi
}

# Instalação de autocompleções para shell
install_shell_completions() {
    if [[ "$INSTALL_COMPLETIONS" == false ]]; then
        return 0
    fi

    step "Instalando autocompleções de shell"
    local bin="$BIN_DIR/$APP_NAME"

    if [[ ! -x "$bin" ]]; then
        return 0
    fi

    # 1. Bash
    local bash_completion_dir=""
    if [[ "$INSTALL_MODE" == "system" ]]; then
        for dir in "/usr/share/bash-completion/completions" "/etc/bash_completion.d"; do
            if [[ -d "$dir" ]]; then
                bash_completion_dir="$dir"
                break
            fi
        done
    else
        bash_completion_dir="${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions"
    fi

    if [[ -n "$bash_completion_dir" ]]; then
        if run_cmd "$bash_completion_dir" mkdir -p "$bash_completion_dir" 2>/dev/null; then
            if run_cmd "$bash_completion_dir" sh -c "\"$bin\" completion bash > \"$bash_completion_dir/$APP_NAME\"" 2>/dev/null; then
                success "Autocompleção Bash instalada em: $bash_completion_dir/$APP_NAME"
            fi
        fi
    fi

    # 2. Zsh
    local zsh_completion_dir=""
    if [[ "$INSTALL_MODE" == "system" ]]; then
        for dir in "/usr/local/share/zsh/site-functions" "/usr/share/zsh/site-functions"; do
            if [[ -d "$dir" ]]; then
                zsh_completion_dir="$dir"
                break
            fi
        done
    else
        zsh_completion_dir="${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions"
    fi

    if [[ -n "$zsh_completion_dir" ]]; then
        if run_cmd "$zsh_completion_dir" mkdir -p "$zsh_completion_dir" 2>/dev/null; then
            if run_cmd "$zsh_completion_dir" sh -c "\"$bin\" completion zsh > \"$zsh_completion_dir/_$APP_NAME\"" 2>/dev/null; then
                success "Autocompleção Zsh instalada em: $zsh_completion_dir/_$APP_NAME"
            fi
        fi
    fi

    # 3. Fish
    local fish_completion_dir=""
    if [[ "$INSTALL_MODE" == "system" ]]; then
        if [[ -d "/usr/share/fish/vendor_completions.d" ]]; then
            fish_completion_dir="/usr/share/fish/vendor_completions.d"
        fi
    else
        if [[ -d "$HOME/.config/fish" ]]; then
            fish_completion_dir="$HOME/.config/fish/completions"
        fi
    fi

    if [[ -n "$fish_completion_dir" ]]; then
        if run_cmd "$fish_completion_dir" mkdir -p "$fish_completion_dir" 2>/dev/null; then
            if run_cmd "$fish_completion_dir" sh -c "\"$bin\" completion fish > \"$fish_completion_dir/$APP_NAME.fish\"" 2>/dev/null; then
                success "Autocompleção Fish instalada em: $fish_completion_dir/$APP_NAME.fish"
            fi
        fi
    fi
}

# Desinstalação completa
uninstall_app() {
    step "Iniciando desinstalação do $APP_NAME"

    # Remoção do binário
    local target_bin="$BIN_DIR/$APP_NAME"
    if [[ -f "$target_bin" ]]; then
        info "Removendo binário: $target_bin"
        run_cmd "$BIN_DIR" rm -f "$target_bin"
        success "Binário removido com sucesso."
    else
        info "Nenhum binário encontrado em: $target_bin"
    fi

    # Remoção de compleções
    local completion_targets=(
        "/usr/share/bash-completion/completions/$APP_NAME"
        "/etc/bash_completion.d/$APP_NAME"
        "${XDG_DATA_HOME:-$HOME/.local/share}/bash-completion/completions/$APP_NAME"
        "/usr/local/share/zsh/site-functions/_$APP_NAME"
        "/usr/share/zsh/site-functions/_$APP_NAME"
        "${XDG_DATA_HOME:-$HOME/.local/share}/zsh/site-functions/_$APP_NAME"
        "/usr/share/fish/vendor_completions.d/$APP_NAME.fish"
        "$HOME/.config/fish/completions/$APP_NAME.fish"
    )

    for comp in "${completion_targets[@]}"; do
        if [[ -f "$comp" ]]; then
            info "Removendo autocompleção: $comp"
            local comp_dir
            comp_dir="$(dirname "$comp")"
            run_cmd "$comp_dir" rm -f "$comp"
        fi
    done

    # Configurações
    if [[ -d "$CONFIG_DIR" ]]; then
        local remove_cfg=false
        if [[ "$ASSUME_YES" == true ]]; then
            remove_cfg=false # Não apaga configurações por padrão mesmo com -y para proteger dados
        else
            echo -e "${YELLOW}Deseja remover o diretório de configurações ($CONFIG_DIR)? [s/N]${RESET} "
            read -r resp || resp="n"
            if [[ "$resp" =~ ^[sSyY]$ ]]; then
                remove_cfg=true
            fi
        fi

        if [[ "$remove_cfg" == true ]]; then
            info "Removendo configurações em: $CONFIG_DIR"
            run_cmd "$CONFIG_DIR" rm -rf "$CONFIG_DIR"
            success "Diretório de configurações removido."
        else
            info "Configurações preservadas em: $CONFIG_DIR"
        fi
    fi

    success "Desinstalação concluída com sucesso."
}

# Fluxo principal
main() {
    check_os
    resolve_directories

    if [[ "$ACTION" == "uninstall" ]]; then
        uninstall_app
        exit 0
    fi

    step "Iniciando instalação do $APP_NAME (Modo: $INSTALL_MODE)"
    info "Origem: $SCRIPT_DIR"
    info "Destino do binário: $BIN_DIR"
    info "Destino da configuração: $CONFIG_DIR"

    build_binary
    install_binary
    install_configuration
    install_shell_completions

    step "Verificando instalação"
    if command -v "$BIN_DIR/$APP_NAME" >/dev/null 2>&1; then
        local ver
        ver="$("$BIN_DIR/$APP_NAME" version 2>/dev/null || echo "instalado com sucesso")"
        success "Instalação concluída com sucesso! ($ver)"
    else
        success "Instalação finalizada."
    fi

    echo -e "\n${BOLD}Próximos passos:${RESET}"
    echo "  1. Configuração padrão instalada em:"
    echo "     $CONFIG_DIR/config.yaml"
    echo "     (edite este arquivo para ajustar provedor ativo, modelo e chave de API)"
    echo "  2. Teste o utilitário executando:"
    echo "     $APP_NAME --help"
    echo "     $APP_NAME -p /caminho/para/documento.pdf"
}

main "$@"
