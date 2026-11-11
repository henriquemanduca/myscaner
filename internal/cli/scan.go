package cli

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"my-scaner/pkg/openrouter"
)

func newScanCmd() *cobra.Command {
	var (
		imagePath string
		model     string
		apiKey    string
	)

	cmd := &cobra.Command{
		Use:   "scan",
		Short: "Extrai texto de uma imagem via OpenRouter e salva em formato Markdown (.md)",
		Long: `O comando scan processa a imagem indicada no parâmetro -p enviando-a ao OpenRouter
para extração precisa de texto (OCR) e estruturação em Markdown.
O arquivo Markdown gerado é salvo automaticamente no mesmo diretório e com o mesmo nome base da imagem.`,
		Example: `  # Extrair texto de uma imagem
  my-scaner scan -p /caminho/para/documento.png

  # Utilizar modelo customizado e passar chave de API diretamente
  my-scaner scan -p ./nota_fiscal.jpg --model anthropic/claude-3.5-sonnet --api-key sk-or-v1-...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			// 1. Validação estrita do arquivo de imagem (existência e extensão suportada)
			if err := openrouter.ValidateImageFile(imagePath); err != nil {
				return err
			}

			// 2. Resolução da chave de API
			key := apiKey
			if key == "" && cfg != nil {
				key = cfg.OpenRouter.APIKey
			}
			if key == "" {
				return fmt.Errorf("chave de API do OpenRouter não encontrada. Defina OPENROUTER_API_KEY, preencha no arquivo de configuração ou utilize a flag --api-key")
			}

			// 3. Resolução de modelo e endpoint
			selectedModel := model
			if selectedModel == "" && cfg != nil && cfg.OpenRouter.Model != "" {
				selectedModel = cfg.OpenRouter.Model
			}

			baseURL := openrouter.DefaultBaseURL
			if cfg != nil && cfg.OpenRouter.BaseURL != "" {
				baseURL = cfg.OpenRouter.BaseURL
			}

			client, err := openrouter.NewClient(key,
				openrouter.WithModel(selectedModel),
				openrouter.WithBaseURL(baseURL),
			)
			if err != nil {
				return fmt.Errorf("falha ao inicializar serviço OpenRouter: %w", err)
			}

			slog.Info("Processando imagem com OpenRouter", "imagem", imagePath, "modelo", selectedModel)

			// 4. Executa a extração de texto
			markdownContent, err := client.ExtractText(ctx, imagePath)
			if err != nil {
				return fmt.Errorf("erro durante extração de texto: %w", err)
			}

			// 5. Salva o arquivo .md no mesmo caminho com mesmo nome
			outputPath := openrouter.OutputMarkdownPath(imagePath)
			if err := os.WriteFile(outputPath, []byte(markdownContent), 0644); err != nil {
				return fmt.Errorf("falha ao gravar arquivo markdown em '%s': %w", outputPath, err)
			}

			slog.Info("Arquivo Markdown gerado com sucesso", "caminho", outputPath)
			fmt.Fprintf(cmd.OutOrStdout(), "✔ Texto extraído com sucesso!\nArquivo gerado: %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringVarP(&imagePath, "path", "p", "", "Caminho do arquivo de imagem (obrigatório: .png, .jpg, .jpeg, .webp, .gif, .bmp, .tiff)")
	cmd.Flags().StringVarP(&model, "model", "m", "", "Modelo do OpenRouter para visão/OCR (padrão: google/gemini-2.0-flash-001)")
	cmd.Flags().StringVar(&apiKey, "api-key", "", "Chave de API do OpenRouter")

	_ = cmd.MarkFlagRequired("path")

	return cmd
}
