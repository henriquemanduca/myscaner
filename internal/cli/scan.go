package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"scaner-cli/pkg/fileutil"
	"scaner-cli/pkg/llm"
	"scaner-cli/pkg/lmstudio"
	"scaner-cli/pkg/openrouter"
)

// scanOptions agrupa as flags do fluxo principal de extração de texto (OCR).
type scanOptions struct {
	inputPath  string
	provider   string
	model      string
	apiKey     string
	baseURL    string
	keepImages bool
	dpi        int
}

// bind registra as flags de extração no comando raiz.
func (o *scanOptions) bind(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.inputPath, "path", "p", "", "Caminho do arquivo (imagem suportada ou .pdf)")
	cmd.Flags().StringVar(&o.provider, "provider", "", "Provedor de IA para visão/OCR: 'openrouter' (nuvem) ou 'lmstudio' (local)")
	cmd.Flags().StringVarP(&o.model, "model", "m", "", "Modelo para visão/OCR (OpenRouter: google/gemini-2.0-flash-001; LM Studio: default ou nome carregado)")
	cmd.Flags().StringVar(&o.apiKey, "api-key", "", "Chave de API (obrigatória para OpenRouter; opcional para LM Studio local)")
	cmd.Flags().StringVar(&o.baseURL, "base-url", "", "Endpoint base da API (OpenRouter: https://openrouter.ai/api/v1; LM Studio: http://localhost:1234/v1)")
	cmd.Flags().BoolVar(&o.keepImages, "keep-images", false, "Mantém as imagens temporárias geradas a partir do PDF")
	cmd.Flags().IntVar(&o.dpi, "dpi", 150, "Resolução em DPI para renderização de páginas de PDF")
}

// runScan executa o fluxo principal de extração para a imagem ou PDF informado.
func (a *app) runScan(cmd *cobra.Command, opts *scanOptions) error {
	ctx := cmd.Context()

	// 1. Validação estrita do arquivo de entrada (existência e extensão suportada)
	if err := fileutil.ValidateInputFile(opts.inputPath); err != nil {
		return err
	}

	// 2. Determinação do provedor (flag > config > default "openrouter")
	selectedProvider := strings.ToLower(opts.provider)
	if selectedProvider == "" && a.cfg != nil && a.cfg.Provider != "" {
		selectedProvider = strings.ToLower(a.cfg.Provider)
	}
	if selectedProvider == "" {
		selectedProvider = "openrouter"
	}

	// 3. Inicialização do Extractor de acordo com o provedor selecionado
	extractor, modelName, err := a.buildExtractor(selectedProvider, opts.model, opts.apiKey, opts.baseURL)
	if err != nil {
		return err
	}

	outputPath := fileutil.OutputMarkdownPath(opts.inputPath)

	// 4. Fluxo para PDF vs Imagem direta
	if fileutil.IsPDF(opts.inputPath) {
		return a.processPDF(ctx, cmd, extractor, opts.inputPath, outputPath, opts.dpi, opts.keepImages)
	}

	return a.processSingleImage(ctx, cmd, extractor, opts.inputPath, outputPath, selectedProvider, modelName)
}

func (a *app) buildExtractor(provider, modelFlag, apiKeyFlag, baseURLFlag string) (llm.Extractor, string, error) {
	switch provider {
	case "openrouter":
		key := apiKeyFlag
		if key == "" && a.cfg != nil {
			key = a.cfg.OpenRouter.APIKey
		}
		if key == "" {
			return nil, "", fmt.Errorf("chave de API do OpenRouter não encontrada. Defina OPENROUTER_API_KEY, preencha no arquivo de configuração ou utilize a flag --api-key")
		}

		selectedModel := modelFlag
		if selectedModel == "" && a.cfg != nil && a.cfg.OpenRouter.Model != "" {
			selectedModel = a.cfg.OpenRouter.Model
		}

		endpoint := baseURLFlag
		if endpoint == "" && a.cfg != nil && a.cfg.OpenRouter.BaseURL != "" {
			endpoint = a.cfg.OpenRouter.BaseURL
		}

		client, err := openrouter.NewClient(key,
			openrouter.WithModel(selectedModel),
			openrouter.WithBaseURL(endpoint),
		)
		if err != nil {
			return nil, "", fmt.Errorf("falha ao inicializar serviço OpenRouter: %w", err)
		}
		return client, selectedModel, nil

	case "lmstudio":
		selectedModel := modelFlag
		if selectedModel == "" && a.cfg != nil && a.cfg.LMStudio.Model != "" {
			selectedModel = a.cfg.LMStudio.Model
		}

		endpoint := baseURLFlag
		if endpoint == "" && a.cfg != nil && a.cfg.LMStudio.BaseURL != "" {
			endpoint = a.cfg.LMStudio.BaseURL
		}

		key := apiKeyFlag
		if key == "" && a.cfg != nil {
			key = a.cfg.LMStudio.APIKey
		}

		client, err := lmstudio.NewClient(
			lmstudio.WithBaseURL(endpoint),
			lmstudio.WithModel(selectedModel),
			lmstudio.WithAPIKey(key),
		)
		if err != nil {
			return nil, "", fmt.Errorf("falha ao inicializar serviço LM Studio: %w", err)
		}
		return client, selectedModel, nil

	default:
		return nil, "", fmt.Errorf("provedor '%s' não suportado. Provedores válidos: 'openrouter', 'lmstudio'", provider)
	}
}

func (a *app) processSingleImage(ctx context.Context, cmd *cobra.Command, extractor llm.Extractor, imagePath string, outputPath string, provider string, modelName string) error {
	a.log.Info("Processando imagem com IA", "provedor", provider, "imagem", imagePath, "modelo", modelName)

	markdownContent, err := extractor.ExtractText(ctx, imagePath)
	if err != nil {
		return fmt.Errorf("erro durante extração de texto: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(markdownContent), 0644); err != nil {
		return fmt.Errorf("falha ao gravar arquivo markdown em '%s': %w", outputPath, err)
	}

	a.log.Info("Arquivo Markdown gerado com sucesso", "caminho", outputPath)
	fmt.Fprintf(cmd.OutOrStdout(), "✔ Texto extraído com sucesso!\nArquivo gerado: %s\n", outputPath)
	return nil
}

func (a *app) processPDF(ctx context.Context, cmd *cobra.Command, extractor llm.Extractor, pdfPath string, outputPath string, dpi int, keepImages bool) error {
	conv, err := a.newConverter()
	if err != nil {
		return fmt.Errorf("conversor de PDF indisponível: %w", err)
	}

	var tempDir string
	if keepImages {
		baseDir := filepath.Dir(pdfPath)
		baseName := strings.TrimSuffix(filepath.Base(pdfPath), filepath.Ext(pdfPath))
		tempDir = filepath.Join(baseDir, baseName+"_pages")
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			return fmt.Errorf("falha ao criar diretório para salvar imagens do PDF: %w", err)
		}
	} else {
		tempDir, err = os.MkdirTemp("", "scaner-cli-pdf-*")
		if err != nil {
			return fmt.Errorf("falha ao criar diretório temporário para conversão de PDF: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()
	}

	a.log.Info("Convertendo páginas do PDF em imagens...", "pdf", pdfPath, "dpi", dpi)
	pageImages, err := conv.ConvertToImages(ctx, pdfPath, tempDir, dpi)
	if err != nil {
		return fmt.Errorf("erro na conversão das páginas do PDF: %w", err)
	}

	totalPages := len(pageImages)
	a.log.Info("Páginas extraídas com sucesso", "total_paginas", totalPages)
	fmt.Fprintf(cmd.OutOrStdout(), "📄 Processando PDF com %d página(s)...\n", totalPages)

	var builder strings.Builder

	for idx, imgFile := range pageImages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pageNum := idx + 1
		a.log.Info("Extraindo texto da página", "pagina", pageNum, "total", totalPages)
		fmt.Fprintf(cmd.OutOrStdout(), "  ↳ Extraindo página %d/%d...\n", pageNum, totalPages)

		pageMarkdown, err := extractor.ExtractText(ctx, imgFile)
		if err != nil {
			return fmt.Errorf("erro ao extrair texto da página %d (%s): %w", pageNum, filepath.Base(imgFile), err)
		}

		if totalPages > 1 {
			if idx > 0 {
				builder.WriteString("\n\n---\n\n")
			}
			builder.WriteString(fmt.Sprintf("<!-- Página %d -->\n\n", pageNum))
		}
		builder.WriteString(pageMarkdown)
	}

	consolidatedMarkdown := builder.String()
	if err := os.WriteFile(outputPath, []byte(consolidatedMarkdown), 0644); err != nil {
		return fmt.Errorf("falha ao gravar arquivo markdown consolidado em '%s': %w", outputPath, err)
	}

	a.log.Info("Arquivo Markdown consolidado do PDF gerado com sucesso", "caminho", outputPath)
	if keepImages {
		fmt.Fprintf(cmd.OutOrStdout(), "🖼 Imagens das páginas salvas em: %s\n", tempDir)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✔ PDF processado com sucesso!\nArquivo gerado: %s\n", outputPath)
	return nil
}
