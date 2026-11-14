package pdf

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// DefaultConversionTimeout é o tempo máximo padrão para converter um PDF em imagens.
const DefaultConversionTimeout = 5 * time.Minute

// Converter define a interface para conversão de documentos PDF em imagens por página.
type Converter interface {
	ConvertToImages(ctx context.Context, pdfPath string, outputDir string, dpi int) ([]string, error)
}

// PopplerConverter implementa a conversão de PDF utilizando o utilitário pdftoppm.
type PopplerConverter struct {
	binaryPath string
	timeout    time.Duration
}

var _ Converter = (*PopplerConverter)(nil)

// Option permite customizar a criação do conversor.
type Option func(*PopplerConverter)

// WithTimeout define o tempo máximo de execução do pdftoppm.
func WithTimeout(timeout time.Duration) Option {
	return func(c *PopplerConverter) {
		if timeout > 0 {
			c.timeout = timeout
		}
	}
}

// NewPopplerConverter instancia o conversor verificando se o utilitário pdftoppm está disponível.
func NewPopplerConverter(opts ...Option) (*PopplerConverter, error) {
	path, err := exec.LookPath("pdftoppm")
	if err != nil {
		return nil, fmt.Errorf("utilitário 'pdftoppm' não encontrado no PATH. Instale o pacote 'poppler-utils' (ex: sudo apt install poppler-utils): %w", err)
	}

	conv := &PopplerConverter{
		binaryPath: path,
		timeout:    DefaultConversionTimeout,
	}
	for _, opt := range opts {
		opt(conv)
	}
	return conv, nil
}

// ConvertToImages converte todas as páginas de um PDF em imagens PNG no outputDir.
// Retorna a lista de caminhos absolutos das imagens ordenadas numericamente pelas páginas.
func (c *PopplerConverter) ConvertToImages(ctx context.Context, pdfPath string, outputDir string, dpi int) ([]string, error) {
	if _, err := os.Stat(pdfPath); err != nil {
		return nil, fmt.Errorf("arquivo PDF não encontrado: %s", pdfPath)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório de saída para imagens do PDF: %w", err)
	}

	if dpi <= 0 {
		dpi = 150
	}

	prefix := filepath.Join(outputDir, "page")
	args := []string{
		"-png",
		"-r", strconv.Itoa(dpi),
		pdfPath,
		prefix,
	}

	timeout := c.timeout
	if timeout <= 0 {
		timeout = DefaultConversionTimeout
	}
	convCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(convCtx, c.binaryPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if convCtx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("conversão do PDF excedeu o tempo limite de %s: %w", timeout, convCtx.Err())
		}
		return nil, fmt.Errorf("falha ao executar pdftoppm (%w): %s", err, string(output))
	}

	// Localiza os arquivos de imagens gerados
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar imagens geradas do PDF: %w", err)
	}

	var imageFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "page-") && strings.HasSuffix(strings.ToLower(name), ".png") {
			imageFiles = append(imageFiles, filepath.Join(outputDir, name))
		}
	}

	if len(imageFiles) == 0 {
		return nil, fmt.Errorf("nenhuma imagem foi gerada a partir do PDF '%s'", pdfPath)
	}

	// Ordena numericamente pelas páginas (ex: page-1.png, page-2.png, page-10.png)
	sort.Slice(imageFiles, func(i, j int) bool {
		numI := extractPageNumber(imageFiles[i])
		numJ := extractPageNumber(imageFiles[j])
		return numI < numJ
	})

	return imageFiles, nil
}

// extractPageNumber extrai o número da página a partir de nomes no formato .../page-01.png ou .../page-1.png
func extractPageNumber(filename string) int {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	clean := strings.TrimSuffix(base, ext)
	parts := strings.Split(clean, "-")
	if len(parts) >= 2 {
		numStr := parts[len(parts)-1]
		if n, err := strconv.Atoi(numStr); err == nil {
			return n
		}
	}
	return 0
}
