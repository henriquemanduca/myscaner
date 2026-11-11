package cli

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"my-scaner/pkg/fileutil"
	"my-scaner/pkg/llm"
	"my-scaner/pkg/lmstudio"
	"my-scaner/pkg/openrouter"
	"my-scaner/pkg/pdf"
)

func newScanCmd() *cobra.Command {
	var (
		inputPath  string
		provider   string
		model      string
		apiKey     string
		baseURL    string
		keepImages bool
		dpi        int
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Extrai texto de imagem ou PDF via IA (OpenRouter ou LM Studio) e salva em formato Markdown (.md)",
		Long: `O comando scan processa a imagem ou documento PDF indicado no parâmetro -p.
Suporta provedores de IA em nuvem (OpenRouter) ou modelos locais (LM Studio).
Para PDFs, converte cada página em imagem com alta resolução, envia para extração de OCR
e consolida todo o conteúdo em um único arquivo Markdown (.md) no mesmo diretório e com mesmo nome base.`,
		Example: `  # Extrair texto via OpenRouter (padrão)
  my-scaner scan -p /caminho/para/documento.png

  # Extrair texto usando LM Studio local (sem enviar dados para nuvem)
  my-scaner scan -p /caminho/para/documento.png --provider lmstudio

  # LM Studio com modelo local e endpoint customizado
  my-scaner scan -p ./relatorio.pdf --provider lmstudio --model qwen2.5-vl-7b-instruct --base-url http://localhost:1234/v1

  # PDF mantendo imagens temporárias das páginas
  my-scaner scan -p ./contrato.pdf --keep-images`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// 1. Validação estrita do arquivo de entrada (existência e extensão suportada)
			if err := fileutil.ValidateInputFile(inputPath); err != nil {
				return err
			}

			// 2. Determinação do provedor (flag > config > default "openrouter")
			selectedProvider := strings.ToLower(provider)
			if selectedProvider == "" && cfg != nil && cfg.Provider != "" {
				selectedProvider = strings.ToLower(cfg.Provider)
			}
			if selectedProvider == "" {
				selectedProvider = "openrouter"
			}

			// 3. Inicialização do Extractor de acordo com o provedor selecionado
			extractor, modelName, err := buildExtractor(selectedProvider, model, apiKey, baseURL)
			if err != nil {
				return err
			}

			outputPath := fileutil.OutputMarkdownPath(inputPath)

			// 4. Fluxo para PDF vs Imagem direta
			if fileutil.IsPDF(inputPath) {
				return processPDF(ctx, cmd, extractor, inputPath, outputPath, dpi, keepImages)
			}

			return processSingleImage(ctx, cmd, extractor, inputPath, outputPath, selectedProvider, modelName)
		},
	}

	cmd.Flags().StringVarP(&inputPath, "path", "p", "", "Caminho do arquivo (imagem suportada ou .pdf)")
	cmd.Flags().StringVar(&provider, "provider", "", "Provedor de IA para visão/OCR: 'openrouter' (nuvem) ou 'lmstudio' (local)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Modelo para visão/OCR (OpenRouter: google/gemini-2.0-flash-001; LM Studio: default ou nome carregado)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Chave de API (obrigatória para OpenRouter; opcional para LM Studio local)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "Endpoint base da API (OpenRouter: https://openrouter.ai/api/v1; LM Studio: http://localhost:1234/v1)")
	cmd.Flags().BoolVar(&keepImages, "keep-images", false, "Mantém as imagens temporárias geradas a partir do PDF")
	cmd.Flags().IntVar(&dpi, "dpi", 150, "Resolução em DPI para renderização de páginas de PDF")

	_ = cmd.MarkFlagRequired("path")

	return cmd
}

func buildExtractor(provider, modelFlag, apiKeyFlag, baseURLFlag string) (llm.Extractor, string, error) {
	switch provider {
	case "openrouter":
		key := apiKeyFlag
		if key == "" && cfg != nil {
			key = cfg.OpenRouter.APIKey
		}
		if key == "" {
			return nil, "", fmt.Errorf("chave de API do OpenRouter não encontrada. Defina OPENROUTER_API_KEY, preencha no arquivo de configuração ou utilize a flag --api-key")
		}

		selectedModel := modelFlag
		if selectedModel == "" && cfg != nil && cfg.OpenRouter.Model != "" {
			selectedModel = cfg.OpenRouter.Model
		}

		endpoint := baseURLFlag
		if endpoint == "" && cfg != nil && cfg.OpenRouter.BaseURL != "" {
			endpoint = cfg.OpenRouter.BaseURL
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
		if selectedModel == "" && cfg != nil && cfg.LMStudio.Model != "" {
			selectedModel = cfg.LMStudio.Model
		}

		endpoint := baseURLFlag
		if endpoint == "" && cfg != nil && cfg.LMStudio.BaseURL != "" {
			endpoint = cfg.LMStudio.BaseURL
		}

		key := apiKeyFlag
		if key == "" && cfg != nil {
			key = cfg.LMStudio.APIKey
		}

		client := lmstudio.NewClient(
			lmstudio.WithBaseURL(endpoint),
			lmstudio.WithModel(selectedModel),
			lmstudio.WithAPIKey(key),
		)
		return client, selectedModel, nil

	default:
		return nil, "", fmt.Errorf("provedor '%s' não suportado. Provedores válidos: 'openrouter', 'lmstudio'", provider)
	}
}

func processSingleImage(ctx context.Context, cmd *cobra.Command, extractor llm.Extractor, imagePath string, outputPath string, provider string, modelName string) error {
	slog.Info("Processando imagem com IA", "provedor", provider, "imagem", imagePath, "modelo", modelName)

	markdownContent, err := extractor.ExtractText(ctx, imagePath)
	if err != nil {
		return fmt.Errorf("erro durante extração de texto: %w", err)
	}

	if err := os.WriteFile(outputPath, []byte(markdownContent), 0644); err != nil {
		return fmt.Errorf("falha ao gravar arquivo markdown em '%s': %w", outputPath, err)
	}

	slog.Info("Arquivo Markdown gerado com sucesso", "caminho", outputPath)
	fmt.Fprintf(cmd.OutOrStdout(), "✔ Texto extraído com sucesso!\nArquivo gerado: %s\n", outputPath)
	return nil
}

func processPDF(ctx context.Context, cmd *cobra.Command, extractor llm.Extractor, pdfPath string, outputPath string, dpi int, keepImages bool) error {
	conv, err := pdf.NewPopplerConverter()
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
		tempDir, err = os.MkdirTemp("", "my-scaner-pdf-*")
		if err != nil {
			return fmt.Errorf("falha ao criar diretório temporário para conversão de PDF: %w", err)
		}
		defer func() {
			_ = os.RemoveAll(tempDir)
		}()
	}

	slog.Info("Convertendo páginas do PDF em imagens...", "pdf", pdfPath, "dpi", dpi)
	pageImages, err := conv.ConvertToImages(ctx, pdfPath, tempDir, dpi)
	if err != nil {
		return fmt.Errorf("erro na conversão das páginas do PDF: %w", err)
	}

	totalPages := len(pageImages)
	slog.Info("Páginas extraídas com sucesso", "total_paginas", totalPages)
	fmt.Fprintf(cmd.OutOrStdout(), "📄 Processando PDF com %d página(s)...\n", totalPages)

	var builder strings.Builder

	for idx, imgFile := range pageImages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pageNum := idx + 1
		slog.Info("Extraindo texto da página", "pagina", pageNum, "total", totalPages)
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

	slog.Info("Arquivo Markdown consolidado do PDF gerado com sucesso", "caminho", outputPath)
	if keepImages {
		fmt.Fprintf(cmd.OutOrStdout(), "🖼 Imagens das páginas salvas em: %s\n", tempDir)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "✔ PDF processado com sucesso!\nArquivo gerado: %s\n", outputPath)
	return nil
}
